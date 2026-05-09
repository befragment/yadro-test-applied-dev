package routing

import (
	"net/http"

	"github.com/befragment/yadro-test-applied-dev/internal/handler/common"
)

func Router(
	logger logger,
	// parseHandler parseHandler,
	// topologyHandler topologyHandler,
	// nodeHandler nodeHandler,
	// logHandler logHandler,
) http.Handler {
	mux := http.NewServeMux()

	// GET /api/v1/info
	mux.HandleFunc("GET /api/v1/ping", common.Ping)

	// // POST /api/v1/parse/
	// mux.HandleFunc("POST /api/v1/parse/", parseHandler.Parse)

	// // GET /api/v1/topology/{log_id}
	// mux.HandleFunc("GET /api/v1/topology/", topologyHandler.GetTopology)

	// // GET /api/v1/node/{node_id}
	// mux.HandleFunc("GET /api/v1/node/", nodeHandler.GetNode)

	// // GET /api/v1/port/{node_id}
	// mux.HandleFunc("GET /api/v1/port/", nodeHandler.GetPorts)

	// // GET /api/v1/log/{log_id}
	// mux.HandleFunc("GET /api/v1/log/", logHandler.GetLog)

	return mux
}