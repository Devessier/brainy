package brainy

import (
	"errors"
	"log"
	"strconv"
	"sync"
)

// ErrUnexpectedBehavior represents an invalid operation we could not perform nor identify.
var ErrUnexpectedBehavior = errors.New("unexpected behavior")

// All reasons why a transition could not be performed.
var (
	ErrInvalidTransitionInvalidCurrentState = errors.New("current state is unexpected")
	ErrInvalidTransitionFinalState          = errors.New("final state reached")
	ErrInvalidTransitionNotImplemented      = errors.New("transition not implemented")
)

type ErrInvalidTransitionNextStateNotImplemented struct {
	NextState StateType
}

func (err *ErrInvalidTransitionNextStateNotImplemented) Error() string {
	return "next state is not implemented: " + string(err.NextState)
}

// ErrInvalidTransition means a transition could not be performed, holding the reason why.
type ErrInvalidTransition struct {
	Err error
}

func (err *ErrInvalidTransition) Unwrap() error {
	return err.Err
}

func (err ErrInvalidTransition) Error() string {
	return "invalid transition: " + err.Err.Error()
}

// ErrTransition represents an error that occured while transitioning to a new state.
type ErrTransition struct {
	Event EventType
	Err   error
}

func (err *ErrTransition) Unwrap() error {
	return err.Err
}

func (err *ErrTransition) Error() string {
	return "could not transition to: " + string(err.Event) + ": " + err.Err.Error()
}

type ErrAction struct {
	ID  int
	Err error
}

func (err *ErrAction) Unwrap() error {
	return err.Err
}

func (err *ErrAction) Error() string {
	return "error in action of index: " + strconv.Itoa(err.ID) + ": " + err.Err.Error()
}

// StateType represents a state described in the state machine.
type StateType string

// Basic states.
const (
	IdleState StateType = ""
	NoopState StateType = "noop"
)

// EventType represents an event transitioning from one state to another one.
type EventType string

// Basic events.
const (
	NoopEvent EventType = "noop"
)

// Context holds data passed to actions functions.
// It can contain what user wants.
type Context interface{}

// An Action is performed when the machine is transitioning to the node where it's defined.
// It takes machine context and returns an event to send to the machine itself, or NoopEvent.
type Action func(*Machine, Context) (EventType, error)

// Events map holds events to listen with the state to transition to when triggered.
type Events map[EventType]StateType

// A StateNode is a node of the state machine.
// It has actions and events to listen to.
//
// No actions can be specified.
// When no events are specified, the state node is of *final* type, which means once reached, the state
// machine can not be transitioned anymore.
type StateNode struct {
	Actions []Action
	On      Events
}

// A StateNodes holds all state nodes of a machine.
type StateNodes map[StateType]StateNode

// A Machine is a simple state machine.
type Machine struct {
	Context Context

	Initial StateType

	previous StateType
	current  StateType

	StateNodes StateNodes

	Debug bool

	lock sync.Mutex
}

// Init initializes the machine.
func (machine *Machine) Init() {
	machine.current = machine.Initial
}

// Previous returns previous state.
func (machine *Machine) Previous() StateType {
	machine.lock.Lock()
	defer machine.lock.Unlock()

	return machine.previous
}

// Current returns current state.
func (machine *Machine) Current() StateType {
	machine.lock.Lock()
	defer machine.lock.Unlock()
	return machine.current
}

// UnsafeCurrent returns current state without taking care of active lock.
func (machine *Machine) UnsafeCurrent() StateType {
	return machine.current
}

func (machine *Machine) getNextState(event EventType) (StateType, error) {
	currentState, ok := machine.StateNodes[machine.current]
	if !ok {
		return NoopState, &ErrInvalidTransition{
			Err: ErrInvalidTransitionInvalidCurrentState,
		}
	}
	if machine.Debug {
		log.Println("state machine: current state:", machine.current)
	}

	if currentState.On == nil {
		return NoopState, &ErrInvalidTransition{
			Err: ErrInvalidTransitionFinalState,
		}
	}

	nextState, ok := currentState.On[event]
	if !ok {
		return NoopState, &ErrInvalidTransition{
			Err: ErrInvalidTransitionNotImplemented,
		}
	}
	if machine.Debug {
		log.Println("state machine: next state:", nextState)
	}

	return nextState, nil
}

func (machine *Machine) executeActions(stateNode StateNode) (EventType, error) {
	for index, actionToRun := range stateNode.Actions {
		stateToReach, err := actionToRun(machine, machine.Context)
		if err != nil {
			return NoopEvent, &ErrAction{
				ID:  index,
				Err: err,
			}
		}
		if stateToReach == NoopEvent {
			continue
		}

		return stateToReach, nil
	}

	return NoopEvent, nil
}

// Send an event to the state machine.
// Returns the new state and an error if one occured, or nil.
func (machine *Machine) Send(event EventType) (StateType, error) {
	machine.lock.Lock()
	defer machine.lock.Unlock()

	for {
		nextState, err := machine.getNextState(event)
		if err != nil {
			return NoopState, &ErrTransition{
				Event: event,
				Err:   err,
			}
		}

		nextStateNode, ok := machine.StateNodes[nextState]
		if !ok {
			return NoopState, &ErrTransition{
				Event: event,
				Err: &ErrInvalidTransition{
					Err: &ErrInvalidTransitionNextStateNotImplemented{
						NextState: nextState,
					},
				},
			}
		}

		machine.previous = machine.current
		machine.current = nextState

		if len(nextStateNode.Actions) == 0 {
			return nextState, nil
		}

		eventToSend, err := machine.executeActions(nextStateNode)
		if err != nil {
			return NoopState, &ErrTransition{
				Event: event,
				Err:   err,
			}
		}
		if eventToSend == NoopEvent {
			return nextState, nil
		}

		event = eventToSend
	}
}
