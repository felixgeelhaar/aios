package runtime

import (
	"context"
	"fmt"
	"path/filepath"
	"time"

	"github.com/felixgeelhaar/aios/internal/model"
	"github.com/felixgeelhaar/aios/internal/policy"
	"github.com/felixgeelhaar/fortify/circuitbreaker"
	"github.com/felixgeelhaar/fortify/retry"
)

type TokenStore interface {
	Put(ctx context.Context, key, value string) error
	Get(ctx context.Context, key string) (string, error)
}

type Runtime struct {
	workspace string
	store     TokenStore
	router    *model.Router
	policy    *policy.Engine

	connectorCB circuitbreaker.CircuitBreaker[struct{}]
	writeRetry  retry.Retry[struct{}]
}

type ExecutionRequest struct {
	SkillID    string
	Version    string
	Input      map[string]any
	UseCase    string
	Budget     string
	PolicyPack string
}

type ExecutionPlan struct {
	SkillID         string
	Version         string
	Model           string
	SanitizedInput  map[string]any
	PolicyTelemetry policy.RuntimeTelemetry
}

func New(workspace string, store TokenStore) *Runtime {
	return &Runtime{
		workspace: workspace,
		store:     store,
		router:    model.NewRouter(),
		policy:    policy.NewEngine(),
		connectorCB: circuitbreaker.New[struct{}](circuitbreaker.Config{
			MaxRequests: 2,
			Interval:    time.Minute,
			Timeout:     5 * time.Second,
			ReadyToTrip: func(counts circuitbreaker.Counts) bool {
				return counts.ConsecutiveFailures >= 3
			},
		}),
		writeRetry: retry.New[struct{}](retry.Config{
			MaxAttempts:   3,
			InitialDelay:  50 * time.Millisecond,
			Multiplier:    2.0,
			BackoffPolicy: retry.BackoffExponential,
			Jitter:        true,
		}),
	}
}

func (r *Runtime) RegistryDir() string {
	return filepath.Join(r.workspace, "registry", "skills")
}

func (r *Runtime) ConnectGoogleDrive(ctx context.Context, token string) error {
	if token == "" {
		return fmt.Errorf("token is required")
	}

	_, err := r.connectorCB.Execute(ctx, func(ctx context.Context) (struct{}, error) {
		_, retryErr := r.writeRetry.Do(ctx, func(ctx context.Context) (struct{}, error) {
			if putErr := r.store.Put(ctx, "gdrive", token); putErr != nil {
				return struct{}{}, putErr
			}
			return struct{}{}, nil
		})
		if retryErr != nil {
			return struct{}{}, retryErr
		}
		return struct{}{}, nil
	})
	return err
}

func (r *Runtime) PrepareExecution(req ExecutionRequest) (ExecutionPlan, error) {
	if req.SkillID == "" {
		return ExecutionPlan{}, fmt.Errorf("skill id is required")
	}
	useCase := req.UseCase
	if useCase == "" {
		useCase = req.SkillID
	}
	route, err := r.router.Decide(model.RouteRequest{
		UseCase:    useCase,
		Budget:     req.Budget,
		PolicyPack: req.PolicyPack,
	})
	if err != nil {
		return ExecutionPlan{}, err
	}
	sanitized, telemetry := r.policy.ApplyRuntimeHooks(req.Input)
	if telemetry.Blocked {
		return ExecutionPlan{}, fmt.Errorf("policy blocked execution: %v", telemetry.Violations)
	}
	return ExecutionPlan{
		SkillID:         req.SkillID,
		Version:         req.Version,
		Model:           route.Model,
		SanitizedInput:  sanitized,
		PolicyTelemetry: telemetry,
	}, nil
}
