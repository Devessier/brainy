package brainy_test

import (
	"errors"
	"testing"

	"github.com/Devessier/brainy"
)

func TestOnOffMachine(t *testing.T) {
	const (
		OnState  brainy.StateType = "on"
		OffState brainy.StateType = "off"

		OnEvent  brainy.EventType = "on"
		OffEvent brainy.EventType = "off"
	)

	onOffMachine := brainy.Machine{
		Initial: OffState,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
				On: brainy.Events{
					OffEvent: OffState,
				},
			},
			OffState: brainy.StateNode{
				On: brainy.Events{
					OnEvent: OnState,
				},
			},
		},
	}
	onOffMachine.Init()

	if currentState := onOffMachine.Current(); currentState != OffState {
		t.Fatalf(
			"machine is in incorrect state %v; expected %v",
			currentState,
			OffState,
		)
	}

	nextState, err := onOffMachine.Send(OnEvent)
	if err != nil {
		t.Fatalf(
			"transition returned an unexpected error %v",
			err,
		)
	}
	if nextState != OnState {
		t.Fatalf(
			"machine is in incorrect state %v; expected %v",
			nextState,
			OnState,
		)
	}
	if currentState := onOffMachine.Current(); currentState != OnState {
		t.Fatalf(
			"machine is in incorrect state %v; expected %v",
			currentState,
			OnState,
		)
	}

	nextState, err = onOffMachine.Send(OnEvent)
	if err == nil {
		t.Error("returned no error when we expected one")
	}
	if !errors.Is(err, brainy.ErrInvalidTransitionNotImplemented) {
		t.Fatalf(
			"returned error is not caused by what we expected %v; expected %v",
			err,
			brainy.ErrInvalidTransitionNotImplemented,
		)
	}

	nextState, err = onOffMachine.Send(OffEvent)
	if err != nil {
		t.Fatalf(
			"transition returned an unexpected error %v",
			err,
		)
	}
	if nextState != OffState {
		t.Fatalf(
			"machine is in incorrect state %v; expected %v",
			nextState,
			OffState,
		)
	}
	if currentState := onOffMachine.Current(); currentState != OffState {
		t.Fatalf(
			"machine is in incorrect state %v; expected %v",
			currentState,
			OffState,
		)
	}
}
