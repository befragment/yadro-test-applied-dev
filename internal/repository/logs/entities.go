package logs

import "time"

import "github.com/befragment/yadro-test-applied-dev/internal/domain"

type logDB struct {
	ID         int64     `db:"id"`
	Status     string    `db:"status"`
	UploadedAt time.Time `db:"uploaded_at"`
	NodesCount int       `db:"nodes_count"`
	PortsCount int       `db:"ports_count"`
}

func mapLogDBToDomain(l logDB) domain.Log {
	return domain.Log{
		ID:         l.ID,
		Status:     l.Status,
		UploadedAt: l.UploadedAt,
		NodesCount: l.NodesCount,
		PortsCount: l.PortsCount,
	}
}
