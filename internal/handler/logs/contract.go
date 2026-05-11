package logs

import (
	"context"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type logService interface {
	ParseLogFiles(ctx context.Context, path string) (int64, error)
	FetchLogMetadata(ctx context.Context, logID int) (domain.Log, error)
	GetTopology(ctx context.Context, logID int) (domain.Topology, error)
}