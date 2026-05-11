package logs

import (
	"context"
	"github.com/Masterminds/squirrel"
	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type LogsRepository struct {
	conn connection
	sb   squirrel.StatementBuilderType
}

func NewLogsRepository(conn connection) *LogsRepository {
	return &LogsRepository{
		conn: conn, 
		sb: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *LogsRepository) FetchMeta(ctx context.Context, logID int64) (domain.Log, error) {
	query, args, err := r.sb.
		Select("id", "status", "uploaded_at", "nodes_count", "ports_count").
		From("logs").
		Where(squirrel.Eq{"id": logID}).
		ToSql()
	if err != nil {
		return domain.Log{}, err
	}

	var row logDB
	err = r.conn.QueryRow(ctx, query, args...).
		Scan(
			&row.ID,
			&row.Status,
			&row.UploadedAt,
			&row.NodesCount,
			&row.PortsCount,
		)
	if err != nil {
		return domain.Log{}, err
	}

	return mapLogDBToDomain(row), nil
} 

func (r *LogsRepository) Create(
	ctx context.Context,
	log domain.Log,
) (int64, error) {

	query, args, err := r.sb.
		Insert("logs").
		Columns(
			"status",
			"uploaded_at",
			"nodes_count",
			"ports_count",
		).
		Values(
			log.Status,
			log.UploadedAt,
			log.NodesCount,
			log.PortsCount,
		).
		Suffix("RETURNING id").
		ToSql()
	if err != nil {
		return 0, err
	}

	var id int64

	err = r.conn.QueryRow(ctx, query, args...).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}