package sync

import (
	"fmt"
	"os"

	"github.com/felixgeelhaar/statekit"
)

const (
	eventDrift  = "DRIFT"
	eventStable = "STABLE"
	eventRepair = "REPAIR"
)

type syncContext struct {
	DriftCount int
}

type Engine struct {
	interp *statekit.Interpreter[syncContext]
}

func NewEngine() *Engine {
	machine, err := statekit.NewMachine[syncContext]("sync").
		WithInitial("clean").
		WithContext(syncContext{}).
		State("clean").
		On(eventDrift).Target("drifted").
		Done().
		State("drifted").
		On(eventRepair).Target("repairing").
		On(eventStable).Target("clean").
		Done().
		State("repairing").
		On(eventDrift).Target("drifted").
		On(eventStable).Target("clean").
		Done().
		Build()
	if err != nil {
		return &Engine{}
	}

	interp := statekit.NewInterpreter(machine)
	interp.Start()
	return &Engine{interp: interp}
}

func (e *Engine) EnsurePath(path string) error {
	if path == "" {
		return fmt.Errorf("path is required")
	}
	return os.MkdirAll(path, 0o750)
}

func (e *Engine) DetectDrift(expected, current map[string]string) []string {
	var drift []string
	for key, exp := range expected {
		if cur, ok := current[key]; !ok || cur != exp {
			drift = append(drift, key)
		}
	}
	if e.interp != nil {
		if len(drift) > 0 {
			e.interp.Send(statekit.Event{Type: eventDrift})
		} else {
			e.interp.Send(statekit.Event{Type: eventStable})
		}
	}
	return drift
}

func (e *Engine) MarkRepairing() {
	if e.interp != nil {
		e.interp.Send(statekit.Event{Type: eventRepair})
	}
}

func (e *Engine) MarkDrifted() {
	if e.interp != nil {
		e.interp.Send(statekit.Event{Type: eventDrift})
	}
}

func (e *Engine) MarkStable() {
	if e.interp != nil {
		e.interp.Send(statekit.Event{Type: eventStable})
	}
}

func (e *Engine) CurrentState() string {
	if e.interp == nil {
		return "unknown"
	}
	return string(e.interp.State().Value)
}
