package brainy

import (
	"errors"
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

func (s StateType) transitions() []Transition {
	return []Transition{
		{
			Target: s,
		},
	}
}

// Basic states.
const (
	NoneState StateType = ""
)

// EventType represents an event transitioning from one state to another one.
type EventType string

func (e EventType) eventType() EventType {
	return e
}

// Basic events.
const (
	NoopEvent EventType = "noop"
)

// Context holds data passed to actions functions.
// It can contain what user wants.
type Context interface{}

// An Action is performed when the machine is transitioning to the node where it's defined.
// It takes machine context and returns an event to send to the machine itself, or NoopEvent.
type Action func(Context, Event) error

// Actions is a slice of Action.
type Actions []Action

type EventWithType struct {
	Event EventType
}

func (e EventWithType) eventType() EventType {
	return e.Event
}

type Event interface {
	eventType() EventType
}

type Transitioner interface {
	transitions() []Transition
}

type Transitions []Transition

func (t Transitions) transitions() []Transition {
	return t
}

type Transition struct {
	Cond    func(Context, Event) bool
	Target  StateType
	Actions Actions
}

func (t Transition) transitions() []Transition {
	return []Transition{t}
}

// Events map holds events to listen with the state to transition to when triggered.
type Events map[EventType]Transitioner

// A StateNode is a node of the state machine.
// It has actions and events to listen to.
//
// No actions can be specified.
// When no events are specified, the state node is of *final* type, which means once reached, the state
// machine can not be transitioned anymore.
type StateNode struct {
	Actions Actions
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

func (machine *Machine) getTransitions(event EventType) ([]Transition, error) {
	currentState, ok := machine.StateNodes[machine.current]
	if !ok {
		return nil, &ErrInvalidTransition{
			Err: ErrInvalidTransitionInvalidCurrentState,
		}
	}

	if currentState.On == nil {
		return nil, &ErrInvalidTransition{
			Err: ErrInvalidTransitionFinalState,
		}
	}

	transitions, ok := currentState.On[event]
	if !ok {
		return nil, &ErrInvalidTransition{
			Err: ErrInvalidTransitionNotImplemented,
		}
	}

	return transitions.transitions(), nil
}

// Send an event to the state machine.
// Returns the new state and an error if one occured, or nil.
func (machine *Machine) Send(event Event) (StateType, error) {
	machine.lock.Lock()
	defer machine.lock.Unlock()

	eventType := event.eventType()

	transitions, err := machine.getTransitions(eventType)
	if err != nil {
		return machine.current, &ErrTransition{
			Event: eventType,
			Err:   err,
		}
	}

	for _, transition := range transitions {
		shouldCommitTransition := true
		if cond := transition.Cond; cond != nil {
			shouldCommitTransition = cond(machine.Context, event)
		}

		if !shouldCommitTransition {
			continue
		}

		if actions := transition.Actions; actions != nil {
			for _, action := range actions {
				if err := action(machine.Context, event); err != nil {
					return machine.current, err
				}
			}
		}

		if target := transition.Target; target != NoneState {
			machine.previous = machine.current
			machine.current = target
		}
	}

	return machine.current, nil
}
