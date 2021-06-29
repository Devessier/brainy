package brainy

import (
	"errors"
	"strconv"
	"strings"
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
	ErrBlankInitialStateForCompoundState    = errors.New("expected an initial state for a compound state node")
	// ErrInvalidTransitionFinalState is returned by Send method when the current state of the state
	// machine does not have any event handler.
	ErrInvalidTransitionFinalState = errors.New("final state reached")
	// ErrInvalidTransitionNotImplemented is returned when the target of a transition does not reference
	// a state that exists in the state machine.
	//
	// Transitions validity is checked when .Init() method is called.
	ErrInvalidTransitionNotImplemented = errors.New("transition not implemented")
	// ErrNoTransitionCouldBeRun is returned when an event could be handled, that is, a state had an handler for this event,
	// but all transition guards returned false.
	// This is usually not an issue.
	ErrNoTransitionCouldBeRun = errors.New("no transition could be run, due to all guards having returned false")
)

type ErrInvalidInitialState struct {
	InvalidInitialState StateType
}

func (err *ErrInvalidInitialState) Error() string {
	return "initial state references an invalid state node: " + string(err.InvalidInitialState)
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

func (s StateType) target() []StateType {
	return []StateType{
		s,
	}
}

func (s StateType) String() string {
	return string(s)
}

// NoneState describes a transition to a state itself.
// As the zero value of a string is an empty string, it allows us to let the Target field
// blank.
const NoneState StateType = ""

type CompoundTarget map[StateType]Targeter

func (c CompoundTarget) transitions() []Transition {
	return []Transition{
		{
			Target: c,
		},
	}
}

func (c CompoundTarget) target() []StateType {
	states := make([]StateType, 0)

	parentStateIndex := 0
	for parentState, nodes := range c {
		// Currently we only want to handle transitions to a single branch of states.
		if parentStateIndex > 0 {
			break
		}

		states = append(states, parentState)
		states = append(states, nodes.target()...)

		parentStateIndex++
	}

	return removeDuplicatesFromStateTypeSlice(states)
}

func (c CompoundTarget) String() string {
	targets := c.target()
	targetsAsString := make([]string, 0, len(targets))

	for _, target := range targets {
		targetsAsString = append(targetsAsString, target.String())
	}

	return strings.Join(targetsAsString, ".")
}

type Targeter interface {
	target() []StateType
	String() string
}

// EventType is the base of an event.
// EventType implements the Event interface, so that we can send an EventType to a state machine.
type EventType string

func (e EventType) eventType() EventType {
	return e
}

const InitialTransitionEventType EventType = "_INITIAL_TRANSITION"

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
	Target  Targeter
	Actions Actions
}

func (t Transition) transitions() []Transition {
	return []Transition{t}
}

// Events map holds which events to listen to and their corresponding Transitions.
// We can use as values a single Transition as well as a Transitions slice.
type Events map[EventType]Transitioner

// JoinStatesIDs takes an unlimited number of state types and returns
// them as a concatenated string, each of them separed by a "." character.
func JoinStatesIDs(statesIDs ...StateType) string {
	concatenatedIDs := ""

	for index, id := range statesIDs {
		concatenatedIDs += id.String()
		if index < len(statesIDs)-1 {
			concatenatedIDs += "."
		}
	}

	return concatenatedIDs
}

// A StateNode is a node of the state machine.
// It has a map of Events to listen to, OnEntry actions to run when the state is entered and
// OnExit actions to run when the state is exited.
//
// All these fields are optional.
// When no events are specified, the state node is of *final* type, which means once reached, the state
// machine can not be transitioned anymore.
type StateNode struct {
	id string

	Context Context

	Initial StateType

	States StateNodes

	OnEntry Actions
	OnExit  Actions

	On Events

	parentStateNode *StateNode
	machineID       StateType
}

func (s *StateNode) Value() string {
	return s.id
}

// Matches returns whether or not the state node is a descendant of the parent state value.
// It takes the parent state value as a variadic list of StateType.
//
// Given that the id of the StateNode is `compound.atomic`:
//  state.Matches(CompoundState)
//  // => true
//
//  state.Matches(CompoundState, AtomicState)
//  // => true
//
//  state.Matches(UnknownState)
//  // => false
func (s *StateNode) Matches(stateSelectors ...StateType) bool {
	selectorsWithMachineID := make([]StateType, 0, len(stateSelectors)+1)
	selectorsWithMachineID = append(selectorsWithMachineID, s.machineID)
	for _, stateSelector := range stateSelectors {
		selectorsWithMachineID = append(selectorsWithMachineID, stateSelector)
	}

	rebuiltStateID := JoinStatesIDs(selectorsWithMachineID...)

	doesMatch := strings.HasPrefix(s.id, rebuiltStateID)
	return doesMatch
}

func (s *StateNode) setChildrenStateNodesIDs(parentStateNodeID string, machineID StateType) {
	for childStateNodeName, childStateNode := range s.States {
		childStateNode.id = parentStateNodeID + "." + childStateNodeName.String()
		childStateNode.machineID = machineID

		if childStateNode.isCompound() {
			childStateNode.setChildrenStateNodesIDs(childStateNode.id, machineID)
		}
	}
}

func (s StateNode) isAtomic() bool {
	return s.States == nil || len(s.States) == 0
}

func (s StateNode) isCompound() bool {
	return !s.isAtomic()
}

func (s *StateNode) resolveMostNestedInitialStateNode() *StateNode {
	if !s.isCompound() {
		return s
	}

	initialStateNode := s.States[s.Initial]
	return initialStateNode.resolveMostNestedInitialStateNode()
}

func (s *StateNode) getTarget(t Targeter) (*StateNode, bool) {
	targetID := t.String()

	for _, childStateNode := range s.States {
		if stateNodeIDEndsWithTargetID := strings.HasSuffix(childStateNode.id, targetID); stateNodeIDEndsWithTargetID {
			return childStateNode.resolveMostNestedInitialStateNode(), true
		}

		if matchingChildStateNode, ok := childStateNode.getTarget(t); ok {
			return matchingChildStateNode.resolveMostNestedInitialStateNode(), true
		}
	}

	return nil, false
}

func (s *StateNode) executeOnEntryActions(c Context, e Event) error {
	actionsToCall := make([]Action, 0)

	stateNodeToEntry := s

	for stateNodeToEntry != nil {
		if onEntryActions := stateNodeToEntry.OnEntry; onEntryActions != nil {
			actionsToCall = append(actionsToCall, onEntryActions...)
		}

		stateNodeToEntry = stateNodeToEntry.parentStateNode
	}

	countOfActions := len(actionsToCall)
	actionsToCallInReverseOrder := make([]Action, countOfActions)

	for index, action := range actionsToCall {
		actionsToCallInReverseOrder[countOfActions-index-1] = action
	}

	for index, action := range actionsToCallInReverseOrder {
		if err := action(c, e); err != nil {
			return &ErrAction{
				Type: onEntryActionType,
				ID:   index,
				Err:  err,
			}
		}
	}

	return nil
}

func (s *StateNode) executeOnExitActions(c Context, e Event) error {
	stateNodeToExit := s

	for stateNodeToExit != nil {
		if onExitActions := stateNodeToExit.OnExit; onExitActions != nil {
			for index, action := range onExitActions {
				if err := action(c, e); err != nil {
					return &ErrAction{
						Type: onExitActionType,
						ID:   index,
						Err:  err,
					}
				}
			}
		}

		stateNodeToExit = stateNodeToExit.parentStateNode
	}

	return nil
}

func (s *StateNode) validate(m *Machine) error {
	if s.isAtomic() {
		return nil
	}

	if s.Initial == NoneState {
		return ErrBlankInitialStateForCompoundState
	}

	if _, ok := s.States[s.Initial]; !ok {
		return &ErrInvalidInitialState{
			InvalidInitialState: s.Initial,
		}
	}

	for _, stateNode := range s.States {
		// Set the parentStateNode of each state node
		stateNode.parentStateNode = s

		// Recursively validate children states
		if stateNode.isCompound() {
			if err := stateNode.validate(m); err != nil {
				return err
			}
		}

		handlers := stateNode.On
		if handlers == nil {
			continue
		}

		for _, events := range handlers {
			transitions := events.transitions()

			for _, transition := range transitions {
				target := transition.Target
				if target == nil || target == NoneState {
					continue
				}

				if _, hasTarget := s.getTarget(target); !hasTarget {
					return ErrInvalidTransitionNotImplemented
				}
			}
		}

	}

	return nil
}

// A StateNodes holds all state nodes of a machine.
type StateNodes map[StateType]*StateNode

// NewMachine takes a StateNode configuration and returns a Machine if one could be created from the given configuration.
// The configuration is validated so that impossible transitions are not possible at runtime.
// If the state machine could not be created, the validation error is returned.
func NewMachine(config StateNode) (*Machine, error) {
	machine := &Machine{
		StateNode: &config,
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
	ID string

	StateNode *StateNode

	previous *StateNode
	current  *StateNode

	lock sync.Mutex
}

func (machine *Machine) setStateNodesIDs() {
	var rootID StateType = StateType(machine.ID)
	if rootID == "" {
		rootID = "(machine)"
	}

	machine.StateNode.id = string(rootID)
	machine.StateNode.machineID = rootID

	for stateNodeName, stateNode := range machine.StateNode.States {
		stateNode.id = machine.StateNode.id + "." + stateNodeName.String()
		stateNode.machineID = rootID

		stateNode.setChildrenStateNodesIDs(stateNode.id, rootID)
	}
}

// Validate ensures all transitions targets are valid states.
func (machine *Machine) validate() error {
	machine.setStateNodesIDs()

	err := machine.StateNode.validate(machine)

	return err
}

// Init initializes the machine and validates transitions target state.
func (machine *Machine) init() error {
	if err := machine.validate(); err != nil {
		return err
	}

	machine.current = machine.StateNode.resolveMostNestedInitialStateNode()
	if err := machine.current.executeOnEntryActions(machine.current.Context, InitialTransitionEventType); err != nil {
		return err
	}

	return nil
}

// Previous returns previous state.
func (machine *Machine) Previous() *StateNode {
	machine.lock.Lock()
	defer machine.lock.Unlock()

	return machine.previous
}

// Current returns current state.
func (machine *Machine) Current() *StateNode {
	machine.lock.Lock()
	defer machine.lock.Unlock()

	return machine.current
}

// UnsafeCurrent returns current state without taking care of active lock.
func (machine *Machine) UnsafeCurrent() *StateNode {
	return machine.current
}

func (machine *Machine) selectTransition(transitions []Transition, event Event) (Transition, bool) {
	for _, transition := range transitions {
		shouldCommitTransition := true
		if cond := transition.Cond; cond != nil {
			shouldCommitTransition = cond(machine.StateNode.Context, event)
		}

		if shouldCommitTransition {
			return transition, true
		}
	}

	return Transition{}, false
}

func (machine *Machine) resolveStateNodeToEnter(stateNodeWithHandler *StateNode, transitionToExecute Transition) (*StateNode, error) {
	stateNodeToEnter := stateNodeWithHandler
	if target := transitionToExecute.Target; target != nil && target != NoneState {
		// Get parent node to be able to target sibbling state nodes.
		parentStateNode := stateNodeWithHandler.parentStateNode
		if parentStateNode == nil {
			return nil, errors.New("parent state node is nil")
		}

		resolvedTargetStateNode, ok := parentStateNode.getTarget(target)
		if !ok {
			return nil, errors.New("could not resolve target")
		}

		stateNodeToEnter = resolvedTargetStateNode
	}

	return stateNodeToEnter, nil
}

func (machine *Machine) executeMicrotask(stateNodeToEnter *StateNode, transitionToExecute Transition, event Event) error {
	if err := machine.current.executeOnExitActions(machine.StateNode.Context, event); err != nil {
		return err
	}

	if actions := transitionToExecute.Actions; actions != nil {
		for index, action := range actions {
			if err := action(machine.StateNode.Context, event); err != nil {
				return &ErrAction{
					Type: transitionActionActionType,
					ID:   index,
					Err:  err,
				}
			}
		}
	}

	if err := stateNodeToEnter.executeOnEntryActions(machine.StateNode.Context, event); err != nil {
		return err
	}

	return nil
}

func (machine *Machine) resolveStateNodeWithHandler(eventType EventType) (*StateNode, Transitioner) {
	stateNode := machine.current

	for stateNode != nil {
		handlers := stateNode.On
		if handlers == nil {
			stateNode = stateNode.parentStateNode
			continue
		}

		eventHandler := handlers[eventType]
		if eventHandler == nil {
			stateNode = stateNode.parentStateNode
			continue
		}

		return stateNode, eventHandler
	}

	return nil, nil
}

// Send an event to the state machine.
// Returns the new state and an error if one occured, or nil.
func (machine *Machine) Send(event Event) (*StateNode, error) {
	machine.lock.Lock()
	defer machine.lock.Unlock()

	eventType := event.eventType()
	stateNodeWithHandler, eventHandler := machine.resolveStateNodeWithHandler(eventType)
	if stateNodeWithHandler == nil {
		return machine.current, ErrInvalidTransitionNotImplemented
	}

	transitions := eventHandler.transitions()
	transitionToExecute, ok := machine.selectTransition(transitions, event)
	if !ok {
		return machine.current, ErrNoTransitionCouldBeRun
	}

	stateNodeToEnter, err := machine.resolveStateNodeToEnter(stateNodeWithHandler, transitionToExecute)
	if err != nil {
		return machine.current, err
	}

	if err := machine.executeMicrotask(stateNodeToEnter, transitionToExecute, event); err != nil {
		return machine.current, err
	}

	machine.previous = machine.current
	machine.current = stateNodeToEnter

	return machine.current, nil
}

func removeDuplicatesFromStateTypeSlice(s []StateType) []StateType {
	encounteredKeys := make(map[StateType]bool)
	uniqueValues := make([]StateType, 0, len(s))

	for _, value := range s {
		_, ok := encounteredKeys[value]
		if ok {
			continue
		}

		uniqueValues = append(uniqueValues, value)

		encounteredKeys[value] = true
	}

	return uniqueValues
}
