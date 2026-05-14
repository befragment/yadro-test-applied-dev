package routing

import (
	"net/http"

	"github.com/befragment/yadro-test-applied-dev/internal/handler/common"
	"github.com/befragment/yadro-test-applied-dev/internal/handler/middleware/logging"
)

func Router(
	logger logger,
	pathNormalizer pathNormalizer,
	nodeHandler nodeHandler,
	logHandler logHandler,
) http.Handler {
	mux := http.NewServeMux()

	// GET /api/v1/ping
	mux.HandleFunc("GET /api/v1/ping", common.Ping)

	// POST /api/v1/parse/
	mux.HandleFunc("POST /api/v1/parse/", logHandler.Parse)

	// GET /api/v1/topology/{log_id}
	mux.HandleFunc("GET /api/v1/topology/", logHandler.Topology)

	// GET /api/v1/node/{node_id}
	mux.HandleFunc("GET /api/v1/node/", nodeHandler.GetNode)

	// GET /api/v1/port/{node_id}
	mux.HandleFunc("GET /api/v1/port/", nodeHandler.GetPorts)

	// GET /api/v1/log/{log_id}
	mux.HandleFunc("GET /api/v1/log/", logHandler.GetLog)

	return logging.LoggingMiddleware(logger, pathNormalizer)(mux)
}
