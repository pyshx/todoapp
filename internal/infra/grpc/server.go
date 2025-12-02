package grpc

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/grpchealth"
	"connectrpc.com/grpcreflect"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"

	"github.com/pyshx/todoapp/gen/todo/v1/todov1connect"
	"github.com/pyshx/todoapp/pkg/auth"
	"github.com/pyshx/todoapp/pkg/idempotency"
	"github.com/pyshx/todoapp/pkg/user"
)

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func NewServer(port int, handler *TaskHandler, userRepo user.Repo, jwtService *auth.JWTService, idempotencyStore idempotency.Store, logger *slog.Logger) *Server {
	interceptors := connect.WithInterceptors(
		NewRecoveryInterceptor(logger),
		NewMetricsInterceptor(),
		NewRequestIDInterceptor(),
		NewLoggingInterceptor(logger),
		NewAuthInterceptor(jwtService, userRepo, logger),
		NewIdempotencyInterceptor(idempotencyStore, logger),
	)

	mux := http.NewServeMux()

	path, httpHandler := todov1connect.NewTodoServiceHandler(handler, interceptors)
	mux.Handle(path, httpHandler)

	checker := grpchealth.NewStaticChecker(todov1connect.TodoServiceName)
	mux.Handle(grpchealth.NewHandler(checker))

	reflector := grpcreflect.NewStaticReflector(todov1connect.TodoServiceName)
	mux.Handle(grpcreflect.NewHandlerV1(reflector))
	mux.Handle(grpcreflect.NewHandlerV1Alpha(reflector))

	mux.Handle("/metrics", promhttp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})

	httpServer := &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: h2c.NewHandler(mux, &http2.Server{}),
	}

	return &Server{httpServer: httpServer, logger: logger}
}

func (s *Server) Start() error {
	s.logger.Info("server starting",
		"addr", s.httpServer.Addr,
		"reflection", true,
		"health", true,
		"metrics", "/metrics",
	)
	return s.httpServer.ListenAndServe()
}

func (s *Server) Shutdown(ctx context.Context) error {
	s.logger.Info("server shutting down")
	return s.httpServer.Shutdown(ctx)
}

func (s *Server) GracefulShutdown(timeout time.Duration) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	return s.Shutdown(ctx)
}
