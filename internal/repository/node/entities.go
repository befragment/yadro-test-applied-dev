package node

import (
	"database/sql"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type nodeDB struct {
	ID              int64          `db:"id"`
	LogID           int64          `db:"log_id"`
	NodeGUID        string         `db:"node_guid"`
	SystemImageGUID sql.NullString `db:"system_image_guid"`
	PortGUID        sql.NullString `db:"port_guid"`
	Description     string         `db:"description"`
	NodeType        int            `db:"node_type"`
	NumPorts        int            `db:"num_ports"`
	ClassVersion    sql.NullInt32  `db:"class_version"`
	BaseVersion     sql.NullInt32  `db:"base_version"`
	LinearFDBCap    sql.NullInt32  `db:"linear_fdb_cap"`
	MulticastFDBCap sql.NullInt32  `db:"multicast_fdb_cap"`
	LifeTimeValue   sql.NullInt32  `db:"life_time_value"`
}

type portDB struct {
	ID                   int64          `db:"id"`
	NodeID               int64          `db:"node_id"`
	NodeGUID             sql.NullString `db:"node_guid"`
	PortGUID             string         `db:"port_guid"`
	PortNum              int            `db:"port_num"`
	LID                  sql.NullInt32  `db:"lid"`
	PortState            sql.NullInt32  `db:"port_state"`
	PortPhyState         sql.NullInt32  `db:"port_phy_state"`
	LinkWidthActive      sql.NullInt32  `db:"link_width_active"`
	LinkSpeedActive      sql.NullInt32  `db:"link_speed_active"`
	LinkRoundTripLatency sql.NullInt32  `db:"link_round_trip_latency"`
}

func mapNodeDBToDomain(n nodeDB) domain.Node {
	return domain.Node{
		Description:     n.Description,
		NodeGUID:        n.NodeGUID,
		SystemImageGUID: nullString(n.SystemImageGUID),
		PortGUID:        nullString(n.PortGUID),
		NodeType:        n.NodeType,
		NumPorts:        n.NumPorts,
		ClassVersion:    nullInt32(n.ClassVersion),
		BaseVersion:     nullInt32(n.BaseVersion),
		LinearFDBCap:    nullInt32Ptr(n.LinearFDBCap),
		MulticastFDBCap: nullInt32Ptr(n.MulticastFDBCap),
		LifeTimeValue:   nullInt32Ptr(n.LifeTimeValue),
	}
}

func mapPortDBToDomain(p portDB) domain.Port {
	return domain.Port{
		NodeGUID:             nullString(p.NodeGUID),
		PortGUID:             p.PortGUID,
		PortNum:              p.PortNum,
		LID:                  nullInt32(p.LID),
		PortState:            nullInt32(p.PortState),
		PortPhyState:         nullInt32(p.PortPhyState),
		LinkWidthActive:      nullInt32(p.LinkWidthActive),
		LinkSpeedActive:      nullInt32(p.LinkSpeedActive),
		LinkRoundTripLatency: nullInt32(p.LinkRoundTripLatency),
	}
}

func nullString(v sql.NullString) string {
	if !v.Valid {
		return ""
	}
	return v.String
}

func nullInt32(v sql.NullInt32) int {
	if !v.Valid {
		return 0
	}
	return int(v.Int32)
}

func nullInt32Ptr(v sql.NullInt32) *int {
	if !v.Valid {
		return nil
	}
	value := int(v.Int32)
	return &value
}
