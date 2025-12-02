package grpc

import (
	"context"
	"log/slog"
	"runtime/debug"
	"strings"
	"time"

	"connectrpc.com/connect"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"

	"github.com/pyshx/todoapp/pkg/apperr"
	"github.com/pyshx/todoapp/pkg/auth"
	"github.com/pyshx/todoapp/pkg/id"
	"github.com/pyshx/todoapp/pkg/idempotency"
	"github.com/pyshx/todoapp/pkg/user"
)

var (
	requestCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "grpc_requests_total",
			Help: "Total gRPC requests by method, status, and error kind",
		},
		[]string{"method", "status", "error_kind"},
	)

	requestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "grpc_request_duration_seconds",
			Help:    "Request duration by method and status",
			Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5},
		},
		[]string{"method", "status"},
	)
)

type AuthInterceptor struct {
	jwtService *auth.JWTService
	userRepo   user.Repo
	logger     *slog.Logger
}

func NewAuthInterceptor(jwtService *auth.JWTService, userRepo user.Repo, logger *slog.Logger) *AuthInterceptor {
	return &AuthInterceptor{jwtService: jwtService, userRepo: userRepo, logger: logger}
}

func (i *AuthInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		// Try JWT authentication first
		authHeader := req.Header().Get("Authorization")
		if authHeader != "" {
			return i.authenticateWithJWT(ctx, req, next, authHeader)
		}

		// Fall back to x-user-id header for backward compatibility
		userIDStr := req.Header().Get("x-user-id")
		if userIDStr == "" {
			return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("authorization header or x-user-id header is required"))
		}

		userID, err := id.ParseUserID(userIDStr)
		if err != nil {
			return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("invalid user ID format"))
		}

		u, err := i.userRepo.FindByID(ctx, userID)
		if err != nil {
			if apperr.IsNotFound(err) {
				return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("user not found"))
			}
			i.logger.Error("failed to find user", "error", err, "user_id", userIDStr)
			return nil, connect.NewError(connect.CodeInternal, err)
		}

		ctx = ContextWithUser(ctx, u)
		return next(ctx, req)
	}
}

func (i *AuthInterceptor) authenticateWithJWT(ctx context.Context, req connect.AnyRequest, next connect.UnaryFunc, authHeader string) (connect.AnyResponse, error) {
	parts := strings.SplitN(authHeader, " ", 2)
	if len(parts) != 2 || strings.ToLower(parts[0]) != "bearer" {
		return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("invalid authorization header format"))
	}

	token := parts[1]
	claims, err := i.jwtService.ValidateToken(token)
	if err != nil {
		switch err {
		case auth.ErrTokenExpired:
			return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("token expired"))
		case auth.ErrInvalidSignature:
			return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("invalid token signature"))
		default:
			return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("invalid token"))
		}
	}

	u, err := i.userRepo.FindByID(ctx, claims.UserID)
	if err != nil {
		if apperr.IsNotFound(err) {
			return nil, connect.NewError(connect.CodeUnauthenticated, apperr.NewErrUnauthenticated("user not found"))
		}
		i.logger.Error("failed to find user", "error", err, "user_id", claims.UserID.String())
		return nil, connect.NewError(connect.CodeInternal, err)
	}

	ctx = ContextWithUser(ctx, u)
	return next(ctx, req)
}

func (i *AuthInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *AuthInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

type LoggingInterceptor struct {
	logger *slog.Logger
}

func NewLoggingInterceptor(logger *slog.Logger) *LoggingInterceptor {
	return &LoggingInterceptor{logger: logger}
}

func (i *LoggingInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()
		method := req.Spec().Procedure

		requestID, _ := RequestIDFromContext(ctx)

		i.logger.Info("request started",
			"method", method,
			"request_id", requestID,
		)

		resp, err := next(ctx, req)

		duration := time.Since(start)
		status := "ok"
		if err != nil {
			status = connect.CodeOf(err).String()
		}

		i.logger.Info("request completed",
			"method", method,
			"request_id", requestID,
			"duration_ms", duration.Milliseconds(),
			"status", status,
		)

		return resp, err
	}
}

func (i *LoggingInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *LoggingInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

type RequestIDInterceptor struct{}

func NewRequestIDInterceptor() *RequestIDInterceptor {
	return &RequestIDInterceptor{}
}

func (i *RequestIDInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		requestID := uuid.New().String()
		ctx = ContextWithRequestID(ctx, requestID)
		resp, err := next(ctx, req)
		if err == nil && resp != nil {
			resp.Header().Set("x-request-id", requestID)
		}
		return resp, err
	}
}

func (i *RequestIDInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *RequestIDInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

type RecoveryInterceptor struct {
	logger *slog.Logger
}

func NewRecoveryInterceptor(logger *slog.Logger) *RecoveryInterceptor {
	return &RecoveryInterceptor{logger: logger}
}

func (i *RecoveryInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (resp connect.AnyResponse, err error) {
		defer func() {
			if r := recover(); r != nil {
				requestID, _ := RequestIDFromContext(ctx)
				i.logger.Error("panic recovered",
					"panic", r,
					"method", req.Spec().Procedure,
					"request_id", requestID,
					"stack", string(debug.Stack()),
				)
				err = connect.NewError(connect.CodeInternal, nil)
			}
		}()
		return next(ctx, req)
	}
}

func (i *RecoveryInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *RecoveryInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

type MetricsInterceptor struct{}

func NewMetricsInterceptor() *MetricsInterceptor {
	return &MetricsInterceptor{}
}

func (i *MetricsInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		start := time.Now()
		method := req.Spec().Procedure

		resp, err := next(ctx, req)

		duration := time.Since(start)
		status := "ok"
		errorKind := ""
		if err != nil {
			code := connect.CodeOf(err)
			status = code.String()
			errorKind = errorKindFromConnectCode(code)
		}

		requestCounter.WithLabelValues(method, status, errorKind).Inc()
		requestDuration.WithLabelValues(method, status).Observe(duration.Seconds())

		return resp, err
	}
}

func errorKindFromConnectCode(code connect.Code) string {
	switch code {
	case connect.CodeNotFound:
		return string(apperr.ErrorKindNotFound)
	case connect.CodePermissionDenied, connect.CodeUnauthenticated:
		return string(apperr.ErrorKindAuth)
	case connect.CodeAborted:
		return string(apperr.ErrorKindConflict)
	case connect.CodeInvalidArgument:
		return string(apperr.ErrorKindValidation)
	default:
		return string(apperr.ErrorKindInternal)
	}
}

func (i *MetricsInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *MetricsInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}

// IdempotencyInterceptor handles idempotent requests for mutation operations
type IdempotencyInterceptor struct {
	store  idempotency.Store
	logger *slog.Logger
}

func NewIdempotencyInterceptor(store idempotency.Store, logger *slog.Logger) *IdempotencyInterceptor {
	return &IdempotencyInterceptor{store: store, logger: logger}
}

func (i *IdempotencyInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		idempotencyKey := req.Header().Get("Idempotency-Key")
		if idempotencyKey == "" {
			return next(ctx, req)
		}

		method := req.Spec().Procedure
		if !i.isMutationMethod(method) {
			return next(ctx, req)
		}

		u, ok := UserFromContext(ctx)
		if !ok {
			return next(ctx, req)
		}

		cacheKey := idempotency.GenerateKey(u.ID().String(), method, []byte(idempotencyKey))

		if _, ok := i.store.Get(ctx, cacheKey); ok {
			i.logger.Info("returning cached idempotent response",
				"method", method,
				"idempotency_key", idempotencyKey,
			)
			return nil, connect.NewError(connect.CodeAlreadyExists, apperr.NewErrAlreadyExists("request", "already processed with this idempotency key"))
		}

		resp, err := next(ctx, req)

		if err == nil {
			i.store.Set(ctx, cacheKey, &idempotency.Response{
				StatusCode: 200,
				Body:       []byte(idempotencyKey),
			})
		}

		return resp, err
	}
}

func (i *IdempotencyInterceptor) isMutationMethod(method string) bool {
	mutationMethods := []string{
		"/todo.v1.TodoService/CreateTask",
		"/todo.v1.TodoService/UpdateTask",
		"/todo.v1.TodoService/DeleteTask",
	}
	for _, m := range mutationMethods {
		if method == m {
			return true
		}
	}
	return false
}

func (i *IdempotencyInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return next
}

func (i *IdempotencyInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return next
}
