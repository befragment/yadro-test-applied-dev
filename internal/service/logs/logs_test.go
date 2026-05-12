package logs_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/befragment/yadro-test-applied-dev/internal/domain"
	"github.com/befragment/yadro-test-applied-dev/internal/service/logs"
)

func TestParseLogFiles(t *testing.T) {
	t.Parallel()

	fixedTime := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)

	validParsed := domain.ParsedLog{
		Nodes: []domain.Node{
			{NodeGUID: "guid-1"},
		},
		Ports: []domain.Port{
			{NodeGUID: "guid-1", PortGUID: "port-1"},
		},
		NodeInfos: []domain.NodeInfo{
			{NodeGUID: "guid-1"},
		},
	}

	errDB := errors.New("database error")

	tests := []struct {
		name    string
		path    string
		prepare func(
			logparser *Mocklogparser,
			logrepo   *Mocklogrepo,
			noderepo  *Mocknoderepo,
			portrepo  *Mockportrepo,
			txm       *Mocktxmanager,
			clock     *Mockclock,
		)
		wantLogID   int64
		wantErr     error
		wantErrWrap error
	}{
		{
			name: "success",
			path: "archive.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("archive.zip").
					Return(validParsed, nil)

				txm.EXPECT().
					InTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				clock.EXPECT().
					Now().
					Return(fixedTime)

				logrepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(42), nil)

				noderepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, nodes []domain.Node) ([]domain.Node, error) {
						result := make([]domain.Node, len(nodes))
						for i, n := range nodes {
							result[i] = n
							result[i].ID = int64(i + 1)
						}
						return result, nil
					})

				portrepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					Return(nil)

				noderepo.EXPECT().
					CreateNodeInfos(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantLogID: 42,
		},
		{
			name: "error: archive not found",
			path: "missing.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("missing.zip").
					Return(domain.ParsedLog{}, os.ErrNotExist)
			},
			wantErr: os.ErrNotExist,
		},
		{
			name: "error: invalid archive format",
			path: "broken.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("broken.zip").
					Return(domain.ParsedLog{}, fmt.Errorf("unexpected EOF"))
			},
			wantErrWrap: logs.ErrInvalidLogArchive,
		},
		{
			name: "error: logrepo.Create fails",
			path: "archive.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("archive.zip").
					Return(validParsed, nil)

				txm.EXPECT().
					InTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				clock.EXPECT().
					Now().
					Return(fixedTime)

				logrepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(0), errDB)
			},
			wantErr: errDB,
		},
		{
			name: "error: noderepo.CreateMany fails",
			path: "archive.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("archive.zip").
					Return(validParsed, nil)

				txm.EXPECT().
					InTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				clock.EXPECT().
					Now().
					Return(fixedTime)

				logrepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				noderepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					Return(nil, errDB)
			},
			wantErr: errDB,
		},
		{
			name: "error: port references unknown node_guid",
			path: "archive.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("archive.zip").
					Return(domain.ParsedLog{
						Nodes: []domain.Node{
							{NodeGUID: "guid-1"},
						},
						Ports: []domain.Port{
							{NodeGUID: "guid-unknown"},
						},
					}, nil)

				txm.EXPECT().
					InTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				clock.EXPECT().
					Now().
					Return(fixedTime)

				logrepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				noderepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, nodes []domain.Node) ([]domain.Node, error) {
						result := make([]domain.Node, len(nodes))
						for i, n := range nodes {
							result[i] = n
							result[i].ID = int64(i + 1)
						}
						return result, nil
					})
			},
			wantErrWrap: logs.ErrInvalidLogArchive,
		},
		{
			name: "error: portrepo.CreateMany fails",
			path: "archive.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("archive.zip").
					Return(validParsed, nil)

				txm.EXPECT().
					InTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				clock.EXPECT().
					Now().
					Return(fixedTime)

				logrepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				noderepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, nodes []domain.Node) ([]domain.Node, error) {
						result := make([]domain.Node, len(nodes))
						for i, n := range nodes {
							result[i] = n
							result[i].ID = int64(i + 1)
						}
						return result, nil
					})

				portrepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					Return(errDB)
			},
			wantErr: errDB,
		},
		{
			name: "error: node_info references unknown node_guid",
			path: "archive.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("archive.zip").
					Return(domain.ParsedLog{
						Nodes: []domain.Node{
							{NodeGUID: "guid-1"},
						},
						Ports: []domain.Port{
							{NodeGUID: "guid-1"},
						},
						NodeInfos: []domain.NodeInfo{
							{NodeGUID: "guid-unknown"},
						},
					}, nil)

				txm.EXPECT().
					InTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				clock.EXPECT().
					Now().
					Return(fixedTime)

				logrepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				noderepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, nodes []domain.Node) ([]domain.Node, error) {
						result := make([]domain.Node, len(nodes))
						for i, n := range nodes {
							result[i] = n
							result[i].ID = int64(i + 1)
						}
						return result, nil
					})

				portrepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					Return(nil)
			},
			wantErrWrap: logs.ErrInvalidLogArchive,
		},
		{
			name: "error: noderepo.CreateNodeInfos fails",
			path: "archive.zip",
			prepare: func(
				logparser *Mocklogparser,
				logrepo   *Mocklogrepo,
				noderepo  *Mocknoderepo,
				portrepo  *Mockportrepo,
				txm       *Mocktxmanager,
				clock     *Mockclock,
			) {
				logparser.EXPECT().
					ParseArchive("archive.zip").
					Return(validParsed, nil)

				txm.EXPECT().
					InTransaction(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, fn func(context.Context) error) error {
						return fn(ctx)
					})

				clock.EXPECT().
					Now().
					Return(fixedTime)

				logrepo.EXPECT().
					Create(gomock.Any(), gomock.Any()).
					Return(int64(1), nil)

				noderepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					DoAndReturn(func(ctx context.Context, nodes []domain.Node) ([]domain.Node, error) {
						result := make([]domain.Node, len(nodes))
						for i, n := range nodes {
							result[i] = n
							result[i].ID = int64(i + 1)
						}
						return result, nil
					})

				portrepo.EXPECT().
					CreateMany(gomock.Any(), gomock.Any()).
					Return(nil)

				noderepo.EXPECT().
					CreateNodeInfos(gomock.Any(), gomock.Any()).
					Return(errDB)
			},
			wantErr: errDB,
		},
	}

	for _, tc := range tests {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockLogParser := NewMocklogparser(ctrl)
			mockLogRepo := NewMocklogrepo(ctrl)
			mockNodeRepo := NewMocknoderepo(ctrl)
			mockPortRepo := NewMockportrepo(ctrl)
			mockTxm := NewMocktxmanager(ctrl)
			mockClock := NewMockclock(ctrl)

			if tc.prepare != nil {
				tc.prepare(mockLogParser, mockLogRepo, mockNodeRepo, mockPortRepo, mockTxm, mockClock)
			}

			svc := logs.NewLogsService(
				mockLogRepo,
				mockNodeRepo,
				mockPortRepo,
				mockTxm,
				mockLogParser,
				mockClock,
			)

			logID, err := svc.ParseLogFiles(context.Background(), tc.path)

			if tc.wantErr != nil {
				assert.ErrorIs(t, err, tc.wantErr)
				assert.Equal(t, int64(0), logID)
				return
			}

			if tc.wantErrWrap != nil {
				assert.ErrorIs(t, err, tc.wantErrWrap)
				assert.Equal(t, int64(0), logID)
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tc.wantLogID, logID)
		})
	}
}