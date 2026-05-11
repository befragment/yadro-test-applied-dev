package logs

import "github.com/befragment/yadro-test-applied-dev/internal/domain"

type ParseLogResponse struct {
	Nodes     []NodeDTO     `json:"nodes"`
	Ports     []PortDTO     `json:"ports"`
	NodeInfos []NodeInfoDTO `json:"node_infos"`
}

type NodeDTO struct {
	Description     string `json:"description"`
	NodeGUID        string `json:"node_guid"`
	SystemImageGUID string `json:"system_image_guid"`
	PortGUID        string `json:"port_guid"`
	NodeType        int    `json:"node_type"`
	NumPorts        int    `json:"num_ports"`
	ClassVersion    int    `json:"class_version"`
	BaseVersion     int    `json:"base_version"`
}

type PortDTO struct {
	NodeGUID             string `json:"node_guid"`
	PortGUID             string `json:"port_guid"`
	PortNum              int    `json:"port_num"`
	LID                  int    `json:"lid"`
	PortState            int    `json:"port_state"`
	PortPhyState         int    `json:"port_phy_state"`
	LinkWidthActive      int    `json:"link_width_active"`
	LinkSpeedActive      int    `json:"link_speed_active"`
	LinkRoundTripLatency int    `json:"link_round_trip_latency"`
}

type NodeInfoDTO struct {
	NodeGUID               string `json:"node_guid"`
	SerialNumber           string `json:"serial_number"`
	PartNumber             string `json:"part_number"`
	Revision               string `json:"revision"`
	ProductName            string `json:"product_name"`
	Endianness             int    `json:"endianness"`
	EnableEndiannessPerJob int    `json:"enable_endianness_per_job"`
	ReproducibilityDisable int    `json:"reproducibility_disable"`
}

type TopologyResponse struct {
	LogID    int64             `json:"log_id"`
	Hosts    []TopologyNodeDTO `json:"hosts"`
	Switches []TopologyNodeDTO `json:"switches"`
}

type TopologyNodeDTO struct {
	ID          int64  `json:"id"`
	GUID        string `json:"guid"`
	Description string `json:"description"`
	PortsCount  int    `json:"ports_count"`
}

type ParseLogFilesResponse struct {
	LogID int64 `json:"log_id"`
}

func newParseLogResponse(parsed *domain.ParsedLog) ParseLogResponse {
	if parsed == nil {
		return ParseLogResponse{}
	}

	nodes := make([]NodeDTO, 0, len(parsed.Nodes))
	for _, n := range parsed.Nodes {
		nodes = append(nodes, NodeDTO{
			Description:     n.Description,
			NodeGUID:        n.NodeGUID,
			SystemImageGUID: n.SystemImageGUID,
			PortGUID:        n.PortGUID,
			NodeType:        n.NodeType,
			NumPorts:        n.NumPorts,
			ClassVersion:    n.ClassVersion,
			BaseVersion:     n.BaseVersion,
		})
	}

	ports := make([]PortDTO, 0, len(parsed.Ports))
	for _, p := range parsed.Ports {
		ports = append(ports, PortDTO{
			NodeGUID:             p.NodeGUID,
			PortGUID:             p.PortGUID,
			PortNum:              p.PortNum,
			LID:                  p.LID,
			PortState:            p.PortState,
			PortPhyState:         p.PortPhyState,
			LinkWidthActive:      p.LinkWidthActive,
			LinkSpeedActive:      p.LinkSpeedActive,
			LinkRoundTripLatency: p.LinkRoundTripLatency,
		})
	}

	nodeInfos := make([]NodeInfoDTO, 0, len(parsed.NodeInfos))
	for _, ni := range parsed.NodeInfos {
		nodeInfos = append(nodeInfos, NodeInfoDTO{
			NodeGUID:               ni.NodeGUID,
			SerialNumber:           ni.SerialNumber,
			PartNumber:             ni.PartNumber,
			Revision:               ni.Revision,
			ProductName:            ni.ProductName,
			Endianness:             ni.Endianness,
			EnableEndiannessPerJob: ni.EnableEndiannessPerJob,
			ReproducibilityDisable: ni.ReproducibilityDisable,
		})
	}

	return ParseLogResponse{
		Nodes:     nodes,
		Ports:     ports,
		NodeInfos: nodeInfos,
	}
}

func newTopologyResponse(topology domain.Topology) TopologyResponse {
	hosts := make([]TopologyNodeDTO, 0, len(topology.Hosts))
	for _, host := range topology.Hosts {
		hosts = append(hosts, TopologyNodeDTO{
			ID:          host.ID,
			GUID:        host.GUID,
			Description: host.Description,
			PortsCount:  host.PortsCount,
		})
	}

	switches := make([]TopologyNodeDTO, 0, len(topology.Switches))
	for _, sw := range topology.Switches {
		switches = append(switches, TopologyNodeDTO{
			ID:          sw.ID,
			GUID:        sw.GUID,
			Description: sw.Description,
			PortsCount:  sw.PortsCount,
		})
	}

	return TopologyResponse{
		LogID:    topology.LogID,
		Hosts:    hosts,
		Switches: switches,
	}
}
