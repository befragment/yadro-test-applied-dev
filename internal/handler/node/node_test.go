package node_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
	"github.com/befragment/yadro-test-applied-dev/internal/handler/node"
	"github.com/golang/mock/gomock"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helpers

func getNodeRequest(rawID string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/api/v1/node/"+rawID, nil)
}

func getPortsRequest(rawID string) *http.Request {
	return httptest.NewRequest(http.MethodGet, "/api/v1/port/"+rawID, nil)
}

func decodeError(t *testing.T, body *bytes.Buffer) map[string]string {
	t.Helper()
	var got map[string]string
	require.NoError(t, json.NewDecoder(body).Decode(&got))
	return got
}

// TestNodeHandler_GetNode

func TestNodeHandler_GetNode(t *testing.T) {
	t.Parallel()

	errInternal := errors.New("storage unavailable")

	tests := []struct {
		name        string
		req         *http.Request
		prepare     func(m *MocknodeService)
		expectation func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success: node returned for valid node_id",
			req:  getNodeRequest("5"),
			prepare: func(m *MocknodeService) {
				m.EXPECT().
					GetByID(gomock.Any(), 5).
					Return(domain.Node{
						ID:              5,
						LogID:           1,
						NodeGUID:        "guid-5",
						SystemImageGUID: "sys-guid-5",
						PortGUID:        "port-guid-5",
						Description:     "compute node",
						NodeType:        1,
						NumPorts:        4,
						ClassVersion:    2,
						BaseVersion:     1,
					}, nil)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

				var got node.NodeDTO
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				assert.Equal(t, node.NodeDTO{
					Description:     "compute node",
					NodeGUID:        "guid-5",
					SystemImageGUID: "sys-guid-5",
					PortGUID:        "port-guid-5",
					NodeType:        1,
					NumPorts:        4,
					ClassVersion:    2,
					BaseVersion:     1,
				}, got)
			},
		},
		{
			name:    "bad request: node_id path param is non-numeric",
			req:     getNodeRequest("abc"),
			prepare: func(m *MocknodeService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "node_id is required", got["error"])
			},
		},
		{
			name:    "bad request: node_id is zero",
			req:     getNodeRequest("0"),
			prepare: func(m *MocknodeService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "node_id is required", got["error"])
			},
		},
		{
			name: "not found: node with given id does not exist",
			req:  getNodeRequest("99"),
			prepare: func(m *MocknodeService) {
				m.EXPECT().
					GetByID(gomock.Any(), 99).
					Return(domain.Node{}, pgx.ErrNoRows)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "node not found", got["error"])
			},
		},
		{
			name: "internal server error: service returns unexpected error",
			req:  getNodeRequest("3"),
			prepare: func(m *MocknodeService) {
				m.EXPECT().
					GetByID(gomock.Any(), 3).
					Return(domain.Node{}, errInternal)
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

			svc := NewMocknodeService(ctrl)
			tc.prepare(svc)

			h := node.NewNodeHandler(svc)
			rr := httptest.NewRecorder()
			h.GetNode(rr, tc.req)

			tc.expectation(t, rr)
		})
	}
}

// TestNodeHandler_GetPorts

func TestNodeHandler_GetPorts(t *testing.T) {
	t.Parallel()

	errInternal := errors.New("query execution failed")

	tests := []struct {
		name        string
		req         *http.Request
		prepare     func(m *MocknodeService)
		expectation func(t *testing.T, rr *httptest.ResponseRecorder)
	}{
		{
			name: "success: ports returned for valid node_id",
			req:  getPortsRequest("5"),
			prepare: func(m *MocknodeService) {
				m.EXPECT().
					GetNodePorts(gomock.Any(), 5).
					Return([]domain.Port{
						{
							NodeGUID:             "guid-5",
							PortGUID:             "port-guid-1",
							PortNum:              1,
							LID:                  100,
							PortState:            4,
							PortPhyState:         5,
							LinkWidthActive:      2,
							LinkSpeedActive:      4,
							LinkRoundTripLatency: 512,
						},
						{
							NodeGUID:             "guid-5",
							PortGUID:             "port-guid-2",
							PortNum:              2,
							LID:                  101,
							PortState:            4,
							PortPhyState:         5,
							LinkWidthActive:      2,
							LinkSpeedActive:      4,
							LinkRoundTripLatency: 480,
						},
					}, nil)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)
				assert.Contains(t, rr.Header().Get("Content-Type"), "application/json")

				var got []node.PortDTO
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				require.Len(t, got, 2)
				assert.Equal(t, "port-guid-1", got[0].PortGUID)
				assert.Equal(t, 1, got[0].PortNum)
				assert.Equal(t, "port-guid-2", got[1].PortGUID)
				assert.Equal(t, 2, got[1].PortNum)
			},
		},
		{
			name: "success: node has no ports, returns empty list",
			req:  getPortsRequest("8"),
			prepare: func(m *MocknodeService) {
				m.EXPECT().
					GetNodePorts(gomock.Any(), 8).
					Return([]domain.Port{}, nil)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusOK, rr.Code)

				var got []node.PortDTO
				require.NoError(t, json.NewDecoder(rr.Body).Decode(&got))
				assert.Empty(t, got)
			},
		},
		{
			name:    "bad request: node_id path param is non-numeric",
			req:     getPortsRequest("xyz"),
			prepare: func(m *MocknodeService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "node_id is required", got["error"])
			},
		},
		{
			name:    "bad request: node_id is zero",
			req:     getPortsRequest("0"),
			prepare: func(m *MocknodeService) {},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusBadRequest, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "node_id is required", got["error"])
			},
		},
		{
			name: "not found: no ports found for given node_id",
			req:  getPortsRequest("42"),
			prepare: func(m *MocknodeService) {
				m.EXPECT().
					GetNodePorts(gomock.Any(), 42).
					Return(nil, pgx.ErrNoRows)
			},
			expectation: func(t *testing.T, rr *httptest.ResponseRecorder) {
				assert.Equal(t, http.StatusNotFound, rr.Code)
				got := decodeError(t, rr.Body)
				assert.Equal(t, "ports not found", got["error"])
			},
		},
		{
			name: "internal server error: service returns unexpected error",
			req:  getPortsRequest("7"),
			prepare: func(m *MocknodeService) {
				m.EXPECT().
					GetNodePorts(gomock.Any(), 7).
					Return(nil, errInternal)
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

			svc := NewMocknodeService(ctrl)
			tc.prepare(svc)

			h := node.NewNodeHandler(svc)
			rr := httptest.NewRecorder()
			h.GetPorts(rr, tc.req)

			tc.expectation(t, rr)
		})
	}
}
