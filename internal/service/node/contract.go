package node

import (
	"context"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type noderepo interface {
	GetByID(ctx context.Context, NodeID int) (domain.Node, error)
	GetNodePorts(ctx context.Context, NodeID int) ([]domain.Port, error)
	GetNodesByLogID(ctx context.Context, LogID int64) ([]domain.Node, error)
}
