//go:generate mockgen -source ${GOFILE} -package ${GOPACKAGE}_test -destination mocks_test.go
package node

import (
	"context"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type nodeService interface {
	GetByID(ctx context.Context, nodeID int) (domain.Node, error)
	GetNodePorts(ctx context.Context, nodeID int) ([]domain.Port, error)
}
