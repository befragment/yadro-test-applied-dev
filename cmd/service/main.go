package main
 
import (
	"context"
	"errors"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/befragment/yadro-test-applied-dev/internal/config"
	handlerlogs "github.com/befragment/yadro-test-applied-dev/internal/handler/logs"
	handlernode "github.com/befragment/yadro-test-applied-dev/internal/handler/node"
	"github.com/befragment/yadro-test-applied-dev/internal/handler/routing"
	"github.com/befragment/yadro-test-applied-dev/internal/lib/logparser"
	"github.com/befragment/yadro-test-applied-dev/internal/lib/txmanager"
	logsrepo "github.com/befragment/yadro-test-applied-dev/internal/repository/logs"
	noderepo "github.com/befragment/yadro-test-applied-dev/internal/repository/node"
	portrepo "github.com/befragment/yadro-test-applied-dev/internal/repository/port"
	logsservice "github.com/befragment/yadro-test-applied-dev/internal/service/logs"
	nodeservice "github.com/befragment/yadro-test-applied-dev/internal/service/node"
	"github.com/befragment/yadro-test-applied-dev/pkg/clock"
	database "github.com/befragment/yadro-test-applied-dev/pkg/database/postgres/pgx"
	l "github.com/befragment/yadro-test-applied-dev/pkg/logger/zap"
	"github.com/befragment/yadro-test-applied-dev/pkg/shutdown"
)

func main() {
	ctx := shutdown.WaitForShutdown()

	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	logger, err := l.New(cfg.LogLevel)
	if err != nil {
		log.Fatalf("Failed to create logger: %v", err)
	}

	pool := database.MustInitPool(cfg.PostgresDSN(), logger)
	defer pool.Close()

	txm := txmanager.NewManager(pool)
	connProvider := txmanager.NewConnectionProvider(txm)
	logsRepository := logsrepo.NewLogsRepository(connProvider)
	nodeRepository := noderepo.NewNodeRepository(connProvider)
	portRepository := portrepo.NewPortRepository(connProvider)

	parser := logparser.NewLogFileParserAdapter()

	nodeSvc := nodeservice.NewNodeService(nodeRepository)
	logSvc := logsservice.NewLogsService(
		logsRepository,
		nodeRepository,
		portRepository,
		txm,
		parser,
		clock.NewClock(),
	)

	nodeHandler := handlernode.NewNodeHandler(nodeSvc)
	logHandler := handlerlogs.NewLogHandler(logSvc)

	router := routing.Router(logger, nodeHandler, logHandler)
	var wg sync.WaitGroup

	logger.Info("Starting service...")
	wg.Add(1)
	go startServer(ctx, cfg.Port, router, logger, &wg)
	<-ctx.Done()
	wg.Wait()
	logger.Info("Service stopped gracefully")
}

func startServer(ctx context.Context, port string, handler http.Handler, logger *l.Logger, wg *sync.WaitGroup) {
	defer wg.Done()

	srv := &http.Server{
		Addr:    port,
		Handler: handler,
	}

	serverErr := make(chan error, 1)
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
		}
		close(serverErr)
	}()

	select {
	case <-ctx.Done():
		logger.Info("Shutdown signal received")
	case err := <-serverErr:
		if err != nil {
			logger.Errorf("Server error: %v", err)
		}
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Errorf("error shutting down server: %v", err)
	}
}
