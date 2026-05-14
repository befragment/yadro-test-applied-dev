package logs_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
	"github.com/befragment/yadro-test-applied-dev/internal/handler/logs"
	logsservice "github.com/befragment/yadro-test-applied-dev/internal/service/logs"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func parseRequest(path string) *http.Request {
	url := "/api/v1/parse/"
	if path != "" {
		url = fmt.Sprintf("/api/v1/parse/?path=%s", path)
	}
	return httptest.NewRequest(http.MethodPost, url, nil)
}

func getLogRequest(rawID string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/api/v1/log/"+rawID, nil)
}

func topologyRequest(rawID string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/api/v1/topology/"+rawID, nil)
}

func decodeError(t *testing.T, body *bytes.Buffer) map[string]string {
	t.Helper()
	var got map[string]string
	require.NoError(t, json.NewDecoder(body).Decode(&got))
	return got
}

// TestLogHandler_Parse

func TestLogHandler_Parse(t *testing.T) {
	t.Parallel()

	errInternal := errors.New("unexpected db failure")

	tests := []struct {
		name        string
		req         *http.Request
		prepare     func(m *MocklogService)
		expectation func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success: archive parsed, returns 201 with log_id",
			req:  parseRequest("/data/archive.zip"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					ParseLogFiles(gomock.Any(), "/data/archive.zip").
					Return(int64(42), nil)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusCreated, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

				var got logs.ParseLogFilesResponse
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				assert.Equal(t, logs.ParseLogFilesResponse{LogID: 42}, got)
			},
		},
		{
			name:    "bad request: path query param is missing",
			req:     parseRequest(""),
			prepare: func(m *MocklogService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "path is required", got["error"])
			},
		},
		{
			name: "not found: archive file does not exist on disk",
			req:  parseRequest("/missing/archive.zip"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					ParseLogFiles(gomock.Any(), "/missing/archive.zip").
					Return(int64(0), os.ErrNotExist)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "file not found", got["error"])
			},
		},
		{
			name: "bad request: archive is corrupted or has invalid format",
			req:  parseRequest("/data/broken.zip"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					ParseLogFiles(gomock.Any(), "/data/broken.zip").
					Return(int64(0), fmt.Errorf("%w: unexpected EOF", logsservice.ErrInvalidLogArchive))
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "broken zip", got["error"])
			},
		},
		{
			name: "internal server error: service returns unexpected error",
			req:  parseRequest("/data/archive.zip"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					ParseLogFiles(gomock.Any(), "/data/archive.zip").
					Return(int64(0), errInternal)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				got := decodeError(t, rr.Body)
				assert.NotEmpty(t, got["error"])
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMocklogService(ctrl)
			tc.prepare(svc)

			h := logs.NewLogHandler(svc)
			rr := httptest.NewRecorder()
			h.Parse(rr, tc.req)

			tc.expectation(t, rr)
		})
	}
}

// TestLogHandler_GetLog

func TestLogHandler_GetLog(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2024, 6, 1, 12, 0, 0, 0, time.UTC)
	errInternal := errors.New("db connection lost")

	tests := []struct {
		name        string
		req         *http.Request
		prepare     func(m *MocklogService)
		expectation func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success: log metadata returned for valid log_id",
			req:  getLogRequest("7"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					FetchLogMetadata(gomock.Any(), 7).
					Return(domain.Log{
						ID:         7,
						Status:     "parsed",
						UploadedAt: fixedTime,
						NodesCount: 3,
						PortsCount: 12,
					}, nil)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

				var got domain.Log
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				assert.Equal(t, int64(7), got.ID)
				assert.Equal(t, "parsed", got.Status)
				assert.Equal(t, 3, got.NodesCount)
				assert.Equal(t, 12, got.PortsCount)
			},
		},
		{
			name:    "bad request: log_id path param is non-numeric",
			req:     getLogRequest("abc"),
			prepare: func(m *MocklogService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "log_id is required", got["error"])
			},
		},
		{
			name:    "bad request: log_id is zero",
			req:     getLogRequest("0"),
			prepare: func(m *MocklogService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "log_id is required", got["error"])
			},
		},
		{
			name: "not found: log with given id does not exist",
			req:  getLogRequest("99"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					FetchLogMetadata(gomock.Any(), 99).
					Return(domain.Log{}, pgx.ErrNoRows)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "log not found", got["error"])
			},
		},
		{
			name: "internal server error: service returns unexpected error",
			req:  getLogRequest("5"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					FetchLogMetadata(gomock.Any(), 5).
					Return(domain.Log{}, errInternal)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				got := decodeError(t, rr.Body)
				assert.NotEmpty(t, got["error"])
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMocklogService(ctrl)
			tc.prepare(svc)

			h := logs.NewLogHandler(svc)
			rr := httptest.NewRecorder()
			h.GetLog(rr, tc.req)

			tc.expectation(t, rr)
		})
	}
}

// TestLogHandler_Topology

func TestLogHandler_Topology(t *testing.T) {
	t.Parallel()

	errInternal := errors.New("query timeout")

	tests := []struct {
		name        string
		req         *http.Request
		prepare     func(m *MocklogService)
		expectation func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success: topology returned for valid log_id",
			req:  topologyRequest("3"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					GetTopology(gomock.Any(), 3).
					Return(domain.Topology{
						LogID: 3,
						Hosts: []domain.TopologyNode{
							{ID: 1, GUID: "guid-host-1", Description: "host1", PortsCount: 2},
						},
						Switches: []domain.TopologyNode{
							{ID: 2, GUID: "guid-sw-1", Description: "sw1", PortsCount: 8},
						},
					}, nil)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

				var got logs.TopologyResponse
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				assert.Equal(t, int64(3), got.LogID)
				require.Len(t, got.Hosts, 1)
				assert.Equal(t, "guid-host-1", got.Hosts[0].GUID)
				require.Len(t, got.Switches, 1)
				assert.Equal(t, "guid-sw-1", got.Switches[0].GUID)
			},
		},
		{
			name:    "bad request: log_id path param is non-numeric",
			req:     topologyRequest("abc"),
			prepare: func(m *MocklogService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "log_id is required", got["error"])
			},
		},
		{
			name:    "bad request: log_id is zero",
			req:     topologyRequest("0"),
			prepare: func(m *MocklogService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "log_id is required", got["error"])
			},
		},
		{
			name: "not found: no topology data for given log_id",
			req:  topologyRequest("10"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					GetTopology(gomock.Any(), 10).
					Return(domain.Topology{}, pgx.ErrNoRows)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "topology not found", got["error"])
			},
		},
		{
			name: "internal server error: service returns unexpected error",
			req:  topologyRequest("2"),
			prepare: func(m *MocklogService) {
				m.EXPECT().
					GetTopology(gomock.Any(), 2).
					Return(domain.Topology{}, errInternal)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusInternalServerError, rr.Code)
				got := decodeError(t, rr.Body)
				assert.NotEmpty(t, got["error"])
			},
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			svc := NewMocklogService(ctrl)
			tc.prepare(svc)

			h := logs.NewLogHandler(svc)
			rr := httptest.NewRecorder()
			h.Topology(rr, tc.req)

			tc.expectation(t, rr)
		})
	}
}
