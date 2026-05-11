package port

import (
	"database/sql"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
)

type portDB struct {
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

func mapPortsDBToDomain(items []portDB) []domain.Port {
	ports := make([]domain.Port, 0, len(items))
	for _, item := range items {
		ports = append(ports, mapPortDBToDomain(item))
	}
	return ports
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
