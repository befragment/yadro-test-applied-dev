package node

import (
	"context"

	"github.com/Masterminds/squirrel"
	"github.com/befragment/yadro-test-applied-dev/internal/domain"
	"github.com/jackc/pgx/v5"
)

type NodeRepository struct {
	conn connection
	sb   squirrel.StatementBuilderType
}

func NewNodeRepository(conn connection) *NodeRepository {
	return &NodeRepository{
		conn: conn,
		sb:   squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}
}

func (r *NodeRepository) GetByID(ctx context.Context, NodeID int) (domain.Node, error) {
	query, args, err := r.sb.
		Select(
			"id",
			"log_id",
			"description",
			"node_guid",
			"system_image_guid",
			"port_guid",
			"node_type",
			"num_ports",
			"class_version",
			"base_version",
			"linear_fdb_cap",
			"multicast_fdb_cap",
			"life_time_value",
		).
		From("nodes").
		Where(squirrel.Eq{"id": NodeID}).
		ToSql()
	if err != nil {
		return domain.Node{}, err
	}

	var row nodeDB
	err = r.conn.QueryRow(ctx, query, args...).
		Scan(
			&row.ID,
			&row.LogID,
			&row.Description,
			&row.NodeGUID,
			&row.SystemImageGUID,
			&row.PortGUID,
			&row.NodeType,
			&row.NumPorts,
			&row.ClassVersion,
			&row.BaseVersion,
			&row.LinearFDBCap,
			&row.MulticastFDBCap,
			&row.LifeTimeValue,
		)
	if err != nil {
		return domain.Node{}, err
	}

	return mapNodeDBToDomain(row), nil
}

func (r *NodeRepository) GetNodePorts(ctx context.Context, NodeID int) ([]domain.Port, error) {
	query, args, err := r.sb.
		Select(
			"p.id",
			"p.node_id",
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
		Where(squirrel.Eq{"p.node_id": NodeID}).
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

	ports := make([]domain.Port, 0)
	for rows.Next() {
		var row portDB
		if err := rows.Scan(
			&row.ID,
			&row.NodeID,
			&row.NodeGUID,
			&row.PortGUID,
			&row.PortNum,
			&row.LID,
			&row.PortState,
			&row.PortPhyState,
			&row.LinkWidthActive,
			&row.LinkSpeedActive,
			&row.LinkRoundTripLatency,
		); err != nil {
			return nil, err
		}
		ports = append(ports, mapPortDBToDomain(row))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(ports) == 0 {
		return nil, pgx.ErrNoRows
	}

	return ports, nil
}

func (r *NodeRepository) GetNodesByLogID(ctx context.Context, LogID int64) ([]domain.Node, error) {
	query, args, err := r.sb.
		Select(
			"id",
			"log_id",
			"description",
			"node_guid",
			"system_image_guid",
			"port_guid",
			"node_type",
			"num_ports",
			"class_version",
			"base_version",
			"linear_fdb_cap",
			"multicast_fdb_cap",
			"life_time_value",
		).
		From("nodes").
		Where(squirrel.Eq{"log_id": LogID}).
		OrderBy("id ASC").
		ToSql()
	if err != nil {
		return nil, err
	}

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	nodes := make([]domain.Node, 0)
	for rows.Next() {
		var row nodeDB
		if err := rows.Scan(
			&row.ID,
			&row.LogID,
			&row.Description,
			&row.NodeGUID,
			&row.SystemImageGUID,
			&row.PortGUID,
			&row.NodeType,
			&row.NumPorts,
			&row.ClassVersion,
			&row.BaseVersion,
			&row.LinearFDBCap,
			&row.MulticastFDBCap,
			&row.LifeTimeValue,
		); err != nil {
			return nil, err
		}
		nodes = append(nodes, mapNodeDBToDomain(row))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return nodes, nil
}

func (r *NodeRepository) GetTopology(ctx context.Context, logID int64) (domain.Topology, error) {
	const (
		nodeTypeHost   = 1
		nodeTypeSwitch = 2
	)

	query, args, err := r.sb.
		Select(
			"id",
			"node_guid",
			"description",
			"num_ports",
			"node_type",
		).
		From("nodes").
		Where(squirrel.Eq{"log_id": logID}).
		OrderBy("id ASC").
		ToSql()
	if err != nil {
		return domain.Topology{}, err
	}

	rows, err := r.conn.Query(ctx, query, args...)
	if err != nil {
		return domain.Topology{}, err
	}
	defer rows.Close()

	topology := domain.Topology{
		LogID:    logID,
		Hosts:    make([]domain.TopologyNode, 0),
		Switches: make([]domain.TopologyNode, 0),
	}

	for rows.Next() {
		var (
			id          int64
			guid        string
			description string
			portsCount  int
			nodeType    int
		)

		if err := rows.Scan(
			&id,
			&guid,
			&description,
			&portsCount,
			&nodeType,
		); err != nil {
			return domain.Topology{}, err
		}

		node := domain.TopologyNode{
			ID:          id,
			GUID:        guid,
			Description: description,
			PortsCount:  portsCount,
		}

		switch nodeType {
		case nodeTypeHost:
			topology.Hosts = append(topology.Hosts, node)
		case nodeTypeSwitch:
			topology.Switches = append(topology.Switches, node)
		}
	}

	if err := rows.Err(); err != nil {
		return domain.Topology{}, err
	}
	if len(topology.Hosts)+len(topology.Switches) == 0 {
		return domain.Topology{}, pgx.ErrNoRows
	}

	return topology, nil
}

func (r *NodeRepository) CreateMany(
	ctx context.Context,
	nodes []domain.Node,
) ([]domain.Node, error) {

	result := make([]domain.Node, 0, len(nodes))

	for _, node := range nodes {

		query, args, err := r.sb.
			Insert("nodes").
			Columns(
				"log_id",
				"node_guid",
				"system_image_guid",
				"port_guid",
				"description",
				"node_type",
				"num_ports",
				"class_version",
				"base_version",
				"linear_fdb_cap",
				"multicast_fdb_cap",
				"life_time_value",
			).
			Values(
				node.LogID,
				node.NodeGUID,
				node.SystemImageGUID,
				node.PortGUID,
				node.Description,
				node.NodeType,
				node.NumPorts,
				node.ClassVersion,
				node.BaseVersion,
				nullableInt(node.LinearFDBCap),
				nullableInt(node.MulticastFDBCap),
				nullableInt(node.LifeTimeValue),
			).
			Suffix("RETURNING id").
			ToSql()
		if err != nil {
			return nil, err
		}

		var id int64

		err = r.conn.QueryRow(ctx, query, args...).Scan(&id)
		if err != nil {
			return nil, err
		}

		node.ID = id

		result = append(result, node)
	}

	return result, nil
}

func nullableInt(v *int) any {
	if v == nil {
		return nil
	}
	return *v
}

func (r *NodeRepository) CreateNodeInfos(
	ctx context.Context,
	nodeInfos []domain.NodeInfo,
) error {
	if len(nodeInfos) == 0 {
		return nil
	}

	for _, ni := range nodeInfos {
		query, args, err := r.sb.
			Insert("nodes_info").
			Columns(
				"node_id",
				"serial_number",
				"part_number",
				"revision",
				"product_name",
				"endianness",
				"enable_endianness_per_job",
				"reproducibility_disable",
			).
			Values(
				ni.NodeID,
				ni.SerialNumber,
				ni.PartNumber,
				ni.Revision,
				ni.ProductName,
				ni.Endianness,
				ni.EnableEndiannessPerJob,
				ni.ReproducibilityDisable,
			).
			ToSql()
		if err != nil {
			return err
		}

		if _, err := r.conn.Exec(ctx, query, args...); err != nil {
			return err
		}
	}

	return nil
}
