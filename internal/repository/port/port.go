package port

import (
	"context"

	sq "github.com/Masterminds/squirrel"
	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type PortRepository struct {
	conn connection
	sb   sq.StatementBuilderType
}

func NewPortRepository(conn connection) *PortRepository {
	return &PortRepository{
		conn: conn,
		sb:   sq.StatementBuilder.PlaceholderFormat(sq.Dollar),
	}
}

func (r *PortRepository) CreateMany(
	ctx context.Context,
	ports []domain.Port,
) error {
	if len(ports) == 0 {
		return nil
	}

	for _, port := range ports {
		nodeID, err := r.findNodeIDByGUID(ctx, port.NodeGUID)
		if err != nil {
			return err
		}

		query, args, err := r.sb.
			Insert("ports").
			Columns(
				"node_id",
				"port_guid",
				"port_num",
				"lid",
				"port_state",
				"port_phy_state",
				"link_width_active",
				"link_speed_active",
				"link_round_trip_latency",
			).
			Values(
				nodeID,
				port.PortGUID,
				port.PortNum,
				port.LID,
				port.PortState,
				port.PortPhyState,
				port.LinkWidthActive,
				port.LinkSpeedActive,
				port.LinkRoundTripLatency,
			).
			ToSql()
		if err != nil {
			return err
		}

		_, err = r.conn.Exec(ctx, query, args...)
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *PortRepository) GetByNodeID(
	ctx context.Context,
	nodeID int64,
) ([]domain.Port, error) {

	query, args, err := r.sb.
		Select(
			"n.node_guid",
			"p.port_guid",
			"p.port_num",
			"p.lid",
			"p.port_state",
			"p.port_phy_state",
			"p.link_width_active",
			"p.link_speed_active",
			"p.link_round_trip_latency",
		).
		From("ports p").
		Join("nodes n ON n.id = p.node_id").
		Where(sq.Eq{"p.node_id": nodeID}).
		OrderBy("p.port_num ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	defer rows.Close()

	portsDB := make([]portDB, 0)

	for rows.Next() {
		var row portDB

		err := rows.Scan(
			&row.NodeGUID,
			&row.PortGUID,
			&row.PortNum,
			&row.LID,
			&row.PortState,
			&row.PortPhyState,
			&row.LinkWidthActive,
			&row.LinkSpeedActive,
			&row.LinkRoundTripLatency,
		)
		if err != nil {
			return nil, err
		}

		portsDB = append(portsDB, row)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return mapPortsDBToDomain(portsDB), nil
}

func (r *PortRepository) findNodeIDByGUID(ctx context.Context, nodeGUID string) (int64, error) {
	query, args, err := r.sb.
		Select("id").
		From("nodes").
		Where(sq.Eq{"node_guid": nodeGUID}).
		Limit(1).
		ToSql()
	if err != nil {
		return 0, err
	}

	var nodeID int64
	if err := r.conn.QueryRow(ctx, query, args...).Scan(&nodeID); err != nil {
		return 0, err
	}
	return nodeID, nil
}
