package node

import "github.com/befragment/yadro-test-applied-dev/internal/domain"

type NodeDTO struct {
	Description     string `json:"description"`
	NodeGUID        string `json:"node_guid"`
	SystemImageGUID string `json:"system_image_guid"`
	PortGUID        string `json:"port_guid"`
	NodeType        int    `json:"node_type"`
	NumPorts        int    `json:"num_ports"`
	ClassVersion    int    `json:"class_version"`
	BaseVersion     int    `json:"base_version"`
}

type PortDTO struct {
	NodeGUID             string `json:"node_guid"`
	PortGUID             string `json:"port_guid"`
	PortNum              int    `json:"port_num"`
	LID                  int    `json:"lid"`
	PortState            int    `json:"port_state"`
	PortPhyState         int    `json:"port_phy_state"`
	LinkWidthActive      int    `json:"link_width_active"`
	LinkSpeedActive      int    `json:"link_speed_active"`
	LinkRoundTripLatency int    `json:"link_round_trip_latency"`
}

func newNodeDTO(n domain.Node) NodeDTO {
	return NodeDTO{
		Description:     n.Description,
		NodeGUID:        n.NodeGUID,
		SystemImageGUID: n.SystemImageGUID,
		PortGUID:        n.PortGUID,
		NodeType:        n.NodeType,
		NumPorts:        n.NumPorts,
		ClassVersion:    n.ClassVersion,
		BaseVersion:     n.BaseVersion,
	}
}

func newPortDTOs(ports []domain.Port) []PortDTO {
	items := make([]PortDTO, 0, len(ports))
	for _, p := range ports {
		items = append(items, PortDTO{
			NodeGUID:             p.NodeGUID,
			PortGUID:             p.PortGUID,
			PortNum:              p.PortNum,
			LID:                  p.LID,
			PortState:            p.PortState,
			PortPhyState:         p.PortPhyState,
			LinkWidthActive:      p.LinkWidthActive,
			LinkSpeedActive:      p.LinkSpeedActive,
			LinkRoundTripLatency: p.LinkRoundTripLatency,
		})
	}
	return items
}