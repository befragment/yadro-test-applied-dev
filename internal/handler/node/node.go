package node

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/befragment/yadro-test-applied-dev/internal/handler/utils"
	"github.com/jackc/pgx/v5"
)

type NodeHandler struct {
	svc nodeService
}

func NewNodeHandler(svc nodeService) *NodeHandler {
	return &NodeHandler{svc: svc}
}

func (h *NodeHandler) GetNode(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nodeID, err := strconv.Atoi(utils.LastPathParam(r.URL.Path))
	if err != nil || nodeID == 0 {
		utils.RespondBadRequest(w, "node_id is required")
		return
	}

	node, err := h.svc.GetByID(ctx, nodeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.RespondNotFound(w, "node not found")
			return
		}
		utils.RespondInternalServerError(w, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, newNodeDTO(node))
}

func (h *NodeHandler) GetPorts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	nodeID, err := strconv.Atoi(utils.LastPathParam(r.URL.Path))
	if err != nil || nodeID == 0 {
		utils.RespondBadRequest(w, "node_id is required")
		return
	}

	ports, err := h.svc.GetNodePorts(ctx, nodeID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.RespondNotFound(w, "ports not found")
			return
		}
		utils.RespondInternalServerError(w, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, newPortDTOs(ports))
}
