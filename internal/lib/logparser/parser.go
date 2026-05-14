package logparser

import (
	"archive/zip"
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type LogFileParserAdapter struct{}

var ErrBrokenZip = errors.New("broken zip")

func NewLogFileParserAdapter() *LogFileParserAdapter {
	return &LogFileParserAdapter{}
}

func (l *LogFileParserAdapter) ParseArchive(path string) (_ domain.ParsedLog, retErr error) {
	r, err := zip.OpenReader(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return domain.ParsedLog{}, fmt.Errorf("logparser: open zip %q: %w", path, err)
		}

		return domain.ParsedLog{}, fmt.Errorf("%w: open zip %q: %v", ErrBrokenZip, path, err)
	}
	defer func() {
		if err := r.Close(); err != nil && retErr == nil {
			retErr = fmt.Errorf("logparser: close zip %q: %w", path, err)
		}
	}()

	dbCSV, sharpInfo, err := findRequiredFiles(r)
	if err != nil {
		return domain.ParsedLog{}, fmt.Errorf("%w: %v", ErrBrokenZip, err)
	}

	// ── parse ibdiagnet2.db_csv ──
	dbData, err := readZipFile(dbCSV)
	if err != nil {
		return domain.ParsedLog{}, fmt.Errorf("%w: read db_csv: %v", ErrBrokenZip, err)
	}
	nodes, ports, sysInfoRows, err := parseDBCSV(dbData)
	if err != nil {
		return domain.ParsedLog{}, fmt.Errorf("%w: parse db_csv: %v", ErrBrokenZip, err)
	}

	sharpData, err := readZipFile(sharpInfo)
	if err != nil {
		return domain.ParsedLog{}, fmt.Errorf("%w: read sharp_an_info: %v", ErrBrokenZip, err)
	}
	sharpMap, err := parseSharpAnInfo(sharpData)
	if err != nil {
		return domain.ParsedLog{}, fmt.Errorf("%w: parse sharp_an_info: %v", ErrBrokenZip, err)
	}

	nodeInfos := mergeNodeInfos(sysInfoRows, sharpMap)

	return domain.ParsedLog{
		Nodes:     nodes,
		Ports:     ports,
		NodeInfos: nodeInfos,
	}, nil
}

const (
	fileDBCSV     = "ibdiagnet2.db_csv"
	fileSharpInfo = "ibdiagnet2.sharp_an_info"
)

func findRequiredFiles(r *zip.ReadCloser) (dbCSV, sharpInfo *zip.File, err error) {
	for _, f := range r.File {
		base := baseName(f.Name)
		switch base {
		case fileDBCSV:
			dbCSV = f
		case fileSharpInfo:
			sharpInfo = f
		}
	}
	if dbCSV == nil {
		return nil, nil, fmt.Errorf("logparser: %s not found in archive", fileDBCSV)
	}
	if sharpInfo == nil {
		return nil, nil, fmt.Errorf("logparser: %s not found in archive", fileSharpInfo)
	}
	return dbCSV, sharpInfo, nil
}

// baseName returns the last path component of a zip entry name.
func baseName(name string) string {
	idx := strings.LastIndexByte(name, '/')
	if idx < 0 {
		return name
	}
	return name[idx+1:]
}

func readZipFile(f *zip.File) ([]byte, error) {
	rc, err := f.Open()
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(rc)
	if closeErr := rc.Close(); closeErr != nil && err == nil {
		return nil, closeErr
	}
	return data, err
}

type dbSection int

const (
	sectionNone dbSection = iota
	sectionNodes
	sectionPorts
	sectionSwitches
	sectionSysInfo
)

type sysInfoRow struct {
	nodeGUID     string
	serialNumber string
	partNumber   string
	revision     string
	productName  string
}

type switchRow struct {
	nodeGUID        string
	linearFDBCap    int
	multicastFDBCap int
	lifeTimeValue   int
}

func parseDBCSV(data []byte) ([]domain.Node, []domain.Port, []sysInfoRow, error) {
	const nodeTypeSwitch = 2

	scanner := bufio.NewScanner(bytes.NewReader(data))

	var (
		nodes    []domain.Node
		ports    []domain.Port
		sysInfos []sysInfoRow
		switches map[string]switchRow

		section dbSection
		headers []string
	)

	switches = make(map[string]switchRow)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		switch line {
		case "START_NODES":
			section = sectionNodes
			headers = nil
			continue
		case "END_NODES":
			section = sectionNone
			continue
		case "START_PORTS":
			section = sectionPorts
			headers = nil
			continue
		case "END_PORTS":
			section = sectionNone
			continue
		case "START_SWITCHES":
			section = sectionSwitches
			headers = nil
			continue
		case "END_SWITCHES":
			section = sectionNone
			continue
		case "START_SYSTEM_GENERAL_INFORMATION":
			section = sectionSysInfo
			headers = nil
			continue
		case "END_SYSTEM_GENERAL_INFORMATION":
			section = sectionNone
			continue
		}

		if section == sectionNone {
			continue
		}

		// header row
		if headers == nil {
			headers = splitCSVLine(line)
			continue
		}

		// data rows
		values := splitCSVLine(line)
		row := mapRow(headers, values)

		switch section {
		case sectionNodes:
			n, err := parseNode(row)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("parse node row: %w", err)
			}
			nodes = append(nodes, n)

		case sectionPorts:
			p, err := parsePort(row)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("parse port row: %w", err)
			}
			ports = append(ports, p)

		case sectionSwitches:
			sw, err := parseSwitchRow(row)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("parse switch row: %w", err)
			}
			switches[sw.nodeGUID] = sw

		case sectionSysInfo:
			si, err := parseSysInfoRow(row)
			if err != nil {
				return nil, nil, nil, fmt.Errorf("parse system_general_information row: %w", err)
			}
			sysInfos = append(sysInfos, si)
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, nil, nil, fmt.Errorf("scanner: %w", err)
	}

	for i := range nodes {
		if nodes[i].NodeType != nodeTypeSwitch {
			continue
		}

		sw, ok := switches[nodes[i].NodeGUID]
		if !ok {
			continue
		}

		nodes[i].LinearFDBCap = intPtr(sw.linearFDBCap)
		nodes[i].MulticastFDBCap = intPtr(sw.multicastFDBCap)
		nodes[i].LifeTimeValue = intPtr(sw.lifeTimeValue)
	}

	return nodes, ports, sysInfos, nil
}

func mapRow(headers, values []string) map[string]string {
	m := make(map[string]string, len(headers))
	for i, h := range headers {
		if i < len(values) {
			m[h] = strings.TrimSpace(values[i])
		} else {
			m[h] = ""
		}
	}
	return m
}

func splitCSVLine(line string) []string {
	var fields []string
	var cur strings.Builder
	inQuote := false

	for i := 0; i < len(line); i++ {
		c := line[i]
		switch {
		case c == '"':
			inQuote = !inQuote
		case c == ',' && !inQuote:
			fields = append(fields, cur.String())
			cur.Reset()
		default:
			cur.WriteByte(c)
		}
	}
	fields = append(fields, cur.String())
	return fields
}

func atoi(s string) (int, error) {
	s = strings.TrimSpace(s)
	if s == "" || strings.EqualFold(s, "n/a") {
		return 0, nil
	}
	// Handle hex prefixed values like 0xffff
	if strings.HasPrefix(s, "0x") || strings.HasPrefix(s, "0X") {
		v, err := strconv.ParseInt(s[2:], 16, 64)
		if err != nil {
			return 0, err
		}
		return int(v), nil
	}
	return strconv.Atoi(s)
}

func atoiNA(s string) (int, error) {
	s = strings.TrimSpace(s)
	if strings.EqualFold(s, "n/a") {
		return -1, nil
	}
	return atoi(s)
}

func parseNode(row map[string]string) (domain.Node, error) {
	numPorts, err := atoi(row["NumPorts"])
	if err != nil {
		return domain.Node{}, fmt.Errorf("NumPorts %q: %w", row["NumPorts"], err)
	}
	nodeType, err := atoi(row["NodeType"])
	if err != nil {
		return domain.Node{}, fmt.Errorf("NodeType %q: %w", row["NodeType"], err)
	}
	classVersion, err := atoi(row["ClassVersion"])
	if err != nil {
		return domain.Node{}, fmt.Errorf("ClassVersion %q: %w", row["ClassVersion"], err)
	}
	baseVersion, err := atoi(row["BaseVersion"])
	if err != nil {
		return domain.Node{}, fmt.Errorf("BaseVersion %q: %w", row["BaseVersion"], err)
	}

	return domain.Node{
		Description:     strings.TrimSpace(row["NodeDesc"]),
		NodeGUID:        strings.TrimSpace(row["NodeGUID"]),
		SystemImageGUID: strings.TrimSpace(row["SystemImageGUID"]),
		PortGUID:        strings.TrimSpace(row["PortGUID"]),
		NodeType:        nodeType,
		NumPorts:        numPorts,
		ClassVersion:    classVersion,
		BaseVersion:     baseVersion,
	}, nil
}

func parsePort(row map[string]string) (domain.Port, error) {
	portNum, err := atoi(row["PortNum"])
	if err != nil {
		return domain.Port{}, fmt.Errorf("PortNum %q: %w", row["PortNum"], err)
	}
	lid, err := atoi(row["LID"])
	if err != nil {
		return domain.Port{}, fmt.Errorf("LID %q: %w", row["LID"], err)
	}
	portState, err := atoi(row["PortState"])
	if err != nil {
		return domain.Port{}, fmt.Errorf("PortState %q: %w", row["PortState"], err)
	}
	portPhyState, err := atoi(row["PortPhyState"])
	if err != nil {
		return domain.Port{}, fmt.Errorf("PortPhyState %q: %w", row["PortPhyState"], err)
	}
	linkWidthActv, err := atoi(row["LinkWidthActv"])
	if err != nil {
		return domain.Port{}, fmt.Errorf("LinkWidthActv %q: %w", row["LinkWidthActv"], err)
	}
	linkSpeedActv, err := atoi(row["LinkSpeedActv"])
	if err != nil {
		return domain.Port{}, fmt.Errorf("LinkSpeedActv %q: %w", row["LinkSpeedActv"], err)
	}
	latency, err := atoiNA(row["LinkRoundTripLatency"])
	if err != nil {
		return domain.Port{}, fmt.Errorf("LinkRoundTripLatency %q: %w", row["LinkRoundTripLatency"], err)
	}

	return domain.Port{
		NodeGUID:             strings.TrimSpace(row["NodeGuid"]),
		PortGUID:             strings.TrimSpace(row["PortGuid"]),
		PortNum:              portNum,
		LID:                  lid,
		PortState:            portState,
		PortPhyState:         portPhyState,
		LinkWidthActive:      linkWidthActv,
		LinkSpeedActive:      linkSpeedActv,
		LinkRoundTripLatency: latency,
	}, nil
}

func parseSysInfoRow(row map[string]string) (sysInfoRow, error) {
	return sysInfoRow{
		nodeGUID:     strings.TrimSpace(row["NodeGuid"]),
		serialNumber: strings.TrimSpace(row["SerialNumber"]),
		partNumber:   strings.TrimSpace(row["PartNumber"]),
		revision:     strings.TrimSpace(row["Revision"]),
		productName:  strings.TrimSpace(row["ProductName"]),
	}, nil
}

func parseSwitchRow(row map[string]string) (switchRow, error) {
	linearFDBCap, err := atoi(row["LinearFDBCap"])
	if err != nil {
		return switchRow{}, fmt.Errorf("LinearFDBCap %q: %w", row["LinearFDBCap"], err)
	}
	multicastFDBCap, err := atoi(row["MCastFDBCap"])
	if err != nil {
		return switchRow{}, fmt.Errorf("MCastFDBCap %q: %w", row["MCastFDBCap"], err)
	}
	lifeTimeValue, err := atoi(row["LifeTimeValue"])
	if err != nil {
		return switchRow{}, fmt.Errorf("LifeTimeValue %q: %w", row["LifeTimeValue"], err)
	}

	return switchRow{
		nodeGUID:        strings.TrimSpace(row["NodeGUID"]),
		linearFDBCap:    linearFDBCap,
		multicastFDBCap: multicastFDBCap,
		lifeTimeValue:   lifeTimeValue,
	}, nil
}

func intPtr(v int) *int {
	value := v
	return &value
}

type sharpEntry struct {
	endianness             int
	enableEndiannessPerJob int
	reproducibilityDisable int
}

func parseSharpAnInfo(data []byte) (map[string]sharpEntry, error) {
	result := make(map[string]sharpEntry)

	scanner := bufio.NewScanner(bytes.NewReader(data))

	var (
		currentGUID string
		pending     sharpEntry
		hasPending  bool
	)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "---") {
			continue
		}

		if strings.HasPrefix(line, "SW_GUID=") {
			if hasPending && currentGUID != "" {
				result[currentGUID] = pending
			}
			currentGUID = strings.TrimSpace(strings.TrimPrefix(line, "SW_GUID="))
			pending = sharpEntry{}
			hasPending = true
			continue
		}

		eqIdx := strings.IndexByte(line, '=')
		if eqIdx < 0 || currentGUID == "" {
			continue
		}
		key := strings.TrimSpace(line[:eqIdx])
		val := strings.TrimSpace(line[eqIdx+1:])

		switch key {
		case "endianness":
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("sharp_an_info: endianness %q: %w", val, err)
			}
			pending.endianness = v
		case "enable_endianness_per_job":
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("sharp_an_info: enable_endianness_per_job %q: %w", val, err)
			}
			pending.enableEndiannessPerJob = v
		case "reproducibility_disable":
			v, err := strconv.Atoi(val)
			if err != nil {
				return nil, fmt.Errorf("sharp_an_info: reproducibility_disable %q: %w", val, err)
			}
			pending.reproducibilityDisable = v
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("sharp_an_info scanner: %w", err)
	}

	// flush last block
	if hasPending && currentGUID != "" {
		result[currentGUID] = pending
	}
	return result, nil
}

func mergeNodeInfos(rows []sysInfoRow, sharpMap map[string]sharpEntry) []domain.NodeInfo {
	infos := make([]domain.NodeInfo, 0, len(rows))

	for _, r := range rows {
		ni := domain.NodeInfo{
			NodeGUID:     r.nodeGUID,
			SerialNumber: r.serialNumber,
			PartNumber:   r.partNumber,
			Revision:     r.revision,
			ProductName:  r.productName,
		}

		// Try to find the matching sharp entry:
		// 1. exact GUID match
		// 2. strip leading "0x" prefix
		if entry, ok := sharpMap[r.nodeGUID]; ok {
			ni.Endianness = entry.endianness
			ni.EnableEndiannessPerJob = entry.enableEndiannessPerJob
			ni.ReproducibilityDisable = entry.reproducibilityDisable
		} else {
			stripped := strings.TrimPrefix(strings.ToLower(r.nodeGUID), "0x")
			if entry, ok := sharpMap[stripped]; ok {
				ni.Endianness = entry.endianness
				ni.EnableEndiannessPerJob = entry.enableEndiannessPerJob
				ni.ReproducibilityDisable = entry.reproducibilityDisable
			}
		}

		infos = append(infos, ni)
	}
	return infos
}
