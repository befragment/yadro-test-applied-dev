package logs

import (
	"errors"
	"net/http"
	"os"
	"strconv"

	"github.com/befragment/yadro-test-applied-dev/internal/handler/utils"
	logsservice "github.com/befragment/yadro-test-applied-dev/internal/service/logs"
	"github.com/jackc/pgx/v5"
)

type LogHandler struct {
	lsvc logService
}

func NewLogHandler(lsvc logService) *LogHandler {
	return &LogHandler{lsvc: lsvc}
}

func (h *LogHandler) Parse(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	path := r.URL.Query().Get("path")
	if path == "" {
		utils.RespondBadRequest(w, "path is required")
		return
	}

	logID, err := h.lsvc.ParseLogFiles(ctx, path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			utils.RespondNotFound(w, "file not found")
			return
		}
		if errors.Is(err, logsservice.ErrInvalidLogArchive) {
			utils.RespondBadRequest(w, "broken zip")
			return
		}
		utils.RespondInternalServerError(w, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusCreated, ParseLogFilesResponse{LogID: logID})
}

func (h *LogHandler) GetLog(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logID, err := strconv.Atoi(utils.LastPathParam(r.URL.Path))
	if err != nil || logID == 0 {
		utils.RespondBadRequest(w, "log_id is required")
		return
	}

	logData, err := h.lsvc.FetchLogMetadata(ctx, logID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.RespondNotFound(w, "log not found")
			return
		}
		utils.RespondInternalServerError(w, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, logData)
}

func (h *LogHandler) Topology(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	logID, err := strconv.Atoi(utils.LastPathParam(r.URL.Path))
	if err != nil || logID == 0 {
		utils.RespondBadRequest(w, "log_id is required")
		return
	}

	topology, err := h.lsvc.GetTopology(ctx, logID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			utils.RespondNotFound(w, "topology not found")
			return
		}
		utils.RespondInternalServerError(w, err)
		return
	}

	utils.RespondWithJSON(w, http.StatusOK, newTopologyResponse(topology))
}
