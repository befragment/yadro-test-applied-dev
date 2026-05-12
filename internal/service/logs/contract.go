//go:generate mockgen -source ${GOFILE} -package ${GOPACKAGE}_test -destination mocks_test.go
package logs

import (
	"context"
	"errors"
	"time"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

var ErrInvalidLogArchive = errors.New("invalid log archive")

type logrepo interface {
	FetchMeta(ctx context.Context, logID int64) (domain.Log, error)
	Create(ctx context.Context, log domain.Log) (int64, error)
}

type noderepo interface {
	CreateMany(ctx context.Context, nodes []domain.Node) ([]domain.Node, error)
	CreateNodeInfos(ctx context.Context, nodeInfos []domain.NodeInfo) error
	GetTopology(ctx context.Context, logID int64) (domain.Topology, error)
}

type portrepo interface {
	CreateMany(ctx context.Context, ports []domain.Port) error
}

type logparser interface {
	ParseArchive(path string) (domain.ParsedLog, error)
}

type txmanager interface {
	InTransaction(ctx context.Context, fn func(ctx context.Context) error) error
}

type clock interface {
	Now() time.Time
}
