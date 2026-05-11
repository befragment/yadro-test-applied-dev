package domain

type Topology struct {
	LogID    int64
	Hosts    []TopologyNode
	Switches []TopologyNode
}

type TopologyNode struct {
	ID          int64
	GUID        string
	Description string
	PortsCount  int
}
