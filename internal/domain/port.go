package domain

type Port struct {
	ID int64

	NodeID int64

	NodeGUID string

	PortGUID string

	PortNum int

	LID int

	PortState    int
	PortPhyState int

	LinkWidthActive int
	LinkSpeedActive int

	LinkRoundTripLatency int
}
