package domain

type Node struct {
	ID    int64
	LogID int64

	Description string

	NodeGUID        string
	SystemImageGUID string
	PortGUID        string

	NodeType     int
	NumPorts     int
	ClassVersion int
	BaseVersion  int

	// Switch-only fields. Nil means "not applicable / not present".
	LinearFDBCap    *int
	MulticastFDBCap *int
	LifeTimeValue   *int
}

type NodeInfo struct {
	ID int64

	NodeID int64

	NodeGUID string

	SerialNumber string
	PartNumber   string
	Revision     string
	ProductName  string

	Endianness             int
	EnableEndiannessPerJob int
	ReproducibilityDisable int
}
