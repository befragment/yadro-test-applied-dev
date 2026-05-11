package logs

import (
	"context"
	"errors"
	"fmt"
	"os"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type LogsService struct {
	logrepo  logrepo
	noderepo noderepo
	portrepo portrepo
	txm      txmanager
	logp     logparser
	clock    clock
}

func NewLogsService(
	lr logrepo,
	nr noderepo,
	pr portrepo,
	txm txmanager,
	logp logparser,
	clock clock,
) *LogsService {
	return &LogsService{
		logrepo:  lr,
		noderepo: nr,
		portrepo: pr,
		txm:      txm,
		logp:     logp,
		clock:    clock,
	}
}

func (s *LogsService) FetchLogMetadata(ctx context.Context, logID int) (domain.Log, error) {
	return s.logrepo.FetchMeta(ctx, int64(logID))
}

func (s *LogsService) GetTopology(ctx context.Context, logID int) (domain.Topology, error) {
	return s.noderepo.GetTopology(ctx, int64(logID))
}

func (s *LogsService) ParseLogFiles(
	ctx context.Context,
	path string,
) (int64, error) {
	parsed, err := s.logp.ParseArchive(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return 0, err
		}
		return 0, fmt.Errorf("%w: %v", ErrInvalidLogArchive, err)
	}

	var logID int64
	err = s.txm.InTransaction(ctx, func(ctx context.Context) error {
		logEntity := domain.Log{
			Status:     "parsed",
			UploadedAt: s.clock.Now(),
			NodesCount: len(parsed.Nodes),
			PortsCount: len(parsed.Ports),
		}

		id, err := s.logrepo.Create(ctx, logEntity)
		if err != nil {
			return err
		}

		logID = id

		for i := range parsed.Nodes {
			parsed.Nodes[i].LogID = logID
		}

		nodes, err := s.noderepo.CreateMany(ctx, parsed.Nodes)
		if err != nil {
			return err
		}

		guidToID := make(map[string]int64, len(nodes))

		for _, node := range nodes {
			guidToID[node.NodeGUID] = node.ID
		}

		for i := range parsed.Ports {
			nodeID, ok := guidToID[parsed.Ports[i].NodeGUID]
			if !ok {
				return fmt.Errorf(
					"%w: port has unknown node_guid %q",
					ErrInvalidLogArchive,
					parsed.Ports[i].NodeGUID,
				)
			}
			parsed.Ports[i].NodeID = nodeID
		}

		if err := s.portrepo.CreateMany(
			ctx,
			parsed.Ports,
		); err != nil {
			return err
		}

		for i := range parsed.NodeInfos {
			nodeID, ok := guidToID[parsed.NodeInfos[i].NodeGUID]
			if !ok {
				return fmt.Errorf(
					"%w: node_info has unknown node_guid %q",
					ErrInvalidLogArchive,
					parsed.NodeInfos[i].NodeGUID,
				)
			}
			parsed.NodeInfos[i].NodeID = nodeID
		}

		if err := s.noderepo.CreateNodeInfos(
			ctx,
			parsed.NodeInfos,
		); err != nil {
			return err
		}

		return nil
	})

	if err != nil {
		return 0, err
	}

	return logID, nil
}
