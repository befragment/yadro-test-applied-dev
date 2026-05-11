package node

import (
	"context"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type NodeService struct {
	repo noderepo
}

func NewNodeService(repo noderepo) *NodeService {
	return &NodeService{repo: repo}
}

func (s *NodeService) GetByID(ctx context.Context, nodeID int) (domain.Node, error) {
	return s.repo.GetByID(ctx, nodeID)
}

func (s *NodeService) GetNodePorts(ctx context.Context, nodeID int) ([]domain.Port, error) {
	return s.repo.GetNodePorts(ctx, nodeID)
}

func (s *NodeService) GetNodesByLogID(ctx context.Context, logID int64) ([]domain.Node, error) {
	return s.repo.GetNodesByLogID(ctx, logID)
}

