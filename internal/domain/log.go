package domain

import "time"

type Log struct {
	ID         int64
	Status     string
	UploadedAt time.Time
	NodesCount int
	PortsCount int
}

// ParsedLog is the top-level result returned by ParseArchive.
type ParsedLog struct {
	Nodes     []Node
	Ports     []Port
	NodeInfos []NodeInfo
}
