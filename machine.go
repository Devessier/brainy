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
	// ErrInvalidTransitionInvalidCurrentState is returned when the current state of a state machine
	// is not valid.
	// It can occur when the initial state of a state machine does not match any state node, such
	// as with this state machine definition:
	//  invalidInitialStateStateMachine := Machine{
	//  	Initial: "invalid-state",
	//
	//  	StateNodes: StateNodes{
	//  		"state-1": StateNode{},
	//
	//  		"state-2": StateNode{},
	//  	},
	//  }
	ErrInvalidTransitionInvalidCurrentState = errors.New("current state is unexpected")
	// ErrInvalidTransitionFinalState is returned by Send method when the current state of the state
	// machine does not have any event handler.
	ErrInvalidTransitionFinalState = errors.New("final state reached")
	// ErrInvalidTransitionNotImplemented is returned when the target of a transition does not reference
	// a state that exists in the state machine.
	//
	// Transitions validity is checked when .Init() method is called.
	ErrInvalidTransitionNotImplemented = errors.New("transition not implemented")
)

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

// actionType represents the different types of actions that are possible in a state machine.
type actionType string

// Currently we implemented three types of actions:
//
// 1. onentry actions, that are run when a state is entered
//
// 2. transition actions, that are run during a state transition
//
// 3. onexit actions, that are run when a state is exited
const (
	onEntryActionType          actionType = "onEntry"
	transitionActionActionType actionType = "transitionAction"
	onExitActionType           actionType = "onExit"
)

// ErrAction holds the error that occured in an action (entry action, transition action or exit action)
// as well as the index of this action in the slice of transitions and the type of the action.
type ErrAction struct {
	Type actionType
	ID   int
	Err  error
}

func (err *ErrAction) Unwrap() error {
	return err.Err
}

func (err *ErrAction) Error() string {
	return "error in " + string(err.Type) + " action of index: " + strconv.Itoa(err.ID) + ": " + err.Err.Error()
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

// NoneState describes a transition to a state itself.
// As the zero value of a string is an empty string, it allows us to let the Target field
// blank.
const NoneState StateType = ""

// EventType is the base of an event.
// EventType implements the Event interface, so that we can send an EventType to a state machine.
type EventType string

func (e EventType) eventType() EventType {
	return e
}

// Context holds data passed to actions and guards functions.
// It can contain what user wants.
type Context interface{}

// An Action is performed when the machine is transitioning to the node where it's defined.
// It takes machine context and returns an event to send to the machine itself, or NoopEvent.
type Action func(Context, Event) error

// Actions is a slice of Action.
type Actions []Action

// EventWithType is meant to be embedded in a struct to represent an event with a payload.
// Because EventWithType implements the Event interface, if it is embedded within a struct, this struct
// will also implement the Event interface, and it can be sent to a state machine.
//
//  const AddUserEventType EventType = "ADD_USER"
//
//  eventWithPayload := struct{
//  	EventWithType
//  	Username string
//  }{
//  	EventWithType: EventWithType{
//  		Event: AddUserEventType,
//  	},
//  	Username: "Kim",
//  }
//
//  stateMachine.Send(eventWithPayload)
type EventWithType struct {
	Event EventType
}

func (e EventWithType) eventType() EventType {
	return e.Event
}

// The Event interface describes a type that has a method that returns an EventType.
// Thanks to this method, we can directly send an EventType to a state machine, as well as a struct with a payload
// in addition to the type of the event itself.
type Event interface {
	eventType() EventType
}

// The Transitioner interface requires a type to have a method that returns a slice of Transition.
// It allows us to provide a single Transition as well as a slice of Transition, through Transitions type,
// given that these both types implement the Transitioner interface.
type Transitioner interface {
	transitions() []Transition
}

// Transitions represents a slice of Transition, that implements the Transitioner interface.
//
// Describing several transitions for an event allows to use conditional guards. Each guard will be tested
// and the first one that matches will validate its transition. If conditional guards are not required,
// or if you need only one guard, using a unique Transition is sufficient.
type Transitions []Transition

func (t Transitions) transitions() []Transition {
	return t
}

// A Cond is a function that takes the context of the state machine and the event that triggered the transition
// and returns a boolean that indicates whether to validate or not the transition.
type Cond func(Context, Event) bool

// Transition describe how to go from one state to another one.
//
// The Target is the state that the transition points to. If the Target is left blank, the Transition will be
// a self Transition, that is, the Transition will reenter the current state of the state machine.
// It can be useful to perform some actions according to an event, while remaining in the same state.
//
// The Cond is a Cond function that returns whether or not the transition must be taken. If the Cond is left blank,
// the transition will be validated.
// It is possible to have a slice of Transition and none of them returning true. No Transition will be taken.
//
// The Actions is a slice of Actions functions, that are run when the transition is taken. These functions
// can be used to do fire-and-forget actions, or to assign values to the context of the state machine,
// as currently there is no built-in assign action in brainy.
type Transition struct {
	Cond    Cond
	Target  StateType
	Actions Actions
}

func (t Transition) transitions() []Transition {
	return []Transition{t}
}

// Events map holds which events to listen to and their corresponding Transitions.
// We can use as values a single Transition as well as a Transitions slice.
type Events map[EventType]Transitioner

// A StateNode is a node of the state machine.
// It has a map of Events to listen to, OnEntry actions to run when the state is entered and
// OnExit actions to run when the state is exited.
//
// All these fields are optional.
// When no events are specified, the state node is of *final* type, which means once reached, the state
// machine can not be transitioned anymore.
type StateNode struct {
	Context Context

	Initial StateType

	States StateNodes

	OnEntry Actions
	OnExit  Actions

	On Events
}

// A StateNodes holds all state nodes of a machine.
type StateNodes map[StateType]StateNode

// NewMachine takes a StateNode configuration and returns a Machine if one could be created from the given configuration.
// The configuration is validated so that impossible transitions are not possible at runtime.
// If the state machine could not be created, the validation error is returned.
func NewMachine(config StateNode) (*Machine, error) {
	machine := &Machine{
		StateNode: config,
	}
	if err := machine.init(); err != nil {
		return nil, err
	}

	return machine, nil
}

// A Machine is a simple finite state machine.
// State machines should be instanciated through NewMachine function, that will validate state nodes configuration.
//
// The long-term objective of this library is to have a Golang implementation of state charts as defined by the SCXML specification.
type Machine struct {
	StateNode

	previous StateType
	current  StateType

	lock sync.Mutex
}

// Validate ensures all transitions targets are valid states.
func (machine *Machine) validate() error {
	for _, stateNode := range machine.States {
		handlers := stateNode.On
		if handlers == nil {
			continue
		}

		for _, events := range handlers {
			transitions := events.transitions()

			for _, transition := range transitions {
				target := transition.Target
				if target == NoneState {
					continue
				}

				if _, ok := machine.States[target]; !ok {
					return ErrInvalidTransitionNotImplemented
				}
			}
		}
	}

	return nil
}

// Init initializes the machine and validates transitions target state.
func (machine *Machine) init() error {
	if err := machine.validate(); err != nil {
		return err
	}

	machine.current = machine.Initial

	return nil
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
	currentState, ok := machine.States[machine.current]
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

		currentStateNode := machine.States[machine.current]
		if onExitActions := currentStateNode.OnExit; onExitActions != nil {
			for index, action := range onExitActions {
				if err := action(machine.Context, event); err != nil {
					return machine.current, &ErrAction{
						Type: onExitActionType,
						ID:   index,
						Err:  err,
					}
				}
			}
		}

		if actions := transition.Actions; actions != nil {
			for index, action := range actions {
				if err := action(machine.Context, event); err != nil {
					return machine.current, &ErrAction{
						Type: transitionActionActionType,
						ID:   index,
						Err:  err,
					}
				}
			}
		}

		if target := transition.Target; target != NoneState {
			nextStateNode := machine.States[target]
			if onEntryActions := nextStateNode.OnEntry; onEntryActions != nil {
				for index, action := range onEntryActions {
					if err := action(machine.Context, event); err != nil {
						return machine.current, &ErrAction{
							Type: onEntryActionType,
							ID:   index,
							Err:  err,
						}
					}
				}
			}

			machine.previous = machine.current
			machine.current = target
		}

		break
	}

	return machine.current, nil
}
