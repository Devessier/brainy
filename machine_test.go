package brainy_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"

	"github.com/Devessier/brainy"
	"github.com/Devessier/brainy/mocks"
)

const (
	OnState  brainy.StateType = "on"
	OffState brainy.StateType = "off"

	OnEvent  brainy.EventType = "on"
	OffEvent brainy.EventType = "off"
)

func TestOnOffMachine(t *testing.T) {
	assert := assert.New(t)

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

	assert.Equal(OffState, onOffMachine.Current())

	nextState, err := onOffMachine.Send(OnEvent)
	assert.NoError(err)
	assert.Equal(OnState, nextState)
	assert.Equal(OnState, onOffMachine.Current())

	nextState, err = onOffMachine.Send(OnEvent)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)

	nextState, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)
	assert.Equal(OffState, nextState)
	assert.Equal(OffState, onOffMachine.Current())
}

const (
	IncrementState brainy.StateType = "increment"

	IncrementEventType brainy.EventType = "INCREMENT"
	UnknownEventType   brainy.EventType = "UNKNOWN"
)

type IncrementEvent struct {
	brainy.EventWithType
	IncrementBy int
}

func NewIncrementEvent(incrementBy int) IncrementEvent {
	return IncrementEvent{
		EventWithType: brainy.EventWithType{
			Event: IncrementEventType,
		},
		IncrementBy: incrementBy,
	}
}

type IncrementStateMachineContext struct {
	ToIncrement int
}

func TestCondIsTrueByDefault(t *testing.T) {
	assert := assert.New(t)

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

	assert.Equal(OffState, onOffMachine.Current())

	onOffMachine.Send(OnEvent)

	assert.Equal(OnState, onOffMachine.Current())
}

func TestCondTakesAGuardFunction(t *testing.T) {
	assert := assert.New(t)

	onOffMachine := brainy.Machine{
		Initial: OnState,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
				On: brainy.Events{
					OffEvent: brainy.Transition{
						Cond: func(c brainy.Context, e brainy.Event) bool {
							return true
						},
						Target: OffState,
					},
				},
			},

			OffState: brainy.StateNode{
				On: brainy.Events{
					OnEvent: brainy.Transition{
						Cond: func(c brainy.Context, e brainy.Event) bool {
							return false
						},
						Target: OnState,
					},
				},
			},
		},
	}
	onOffMachine.Init()

	assert.Equal(OnState, onOffMachine.Current())

	_, err := onOffMachine.Send(OffEvent)
	assert.NoError(err)

	assert.Equal(OffState, onOffMachine.Current())

	_, err = onOffMachine.Send(OnEvent)
	assert.NoError(err)

	assert.Equal(OffState, onOffMachine.Current())
}

func TestAllActionsOfATransitionAreCalled(t *testing.T) {
	assert := assert.New(t)

	onOffMachineContext := &struct{}{}

	firstTransitionAction := new(mocks.Action)
	firstTransitionAction.On("Execute", onOffMachineContext, OffEvent).Return(nil)
	secondTransitionAction := new(mocks.Action)
	secondTransitionAction.On("Execute", onOffMachineContext, OffEvent).Return(nil)
	thirdTransitionAction := new(mocks.Action)
	thirdTransitionAction.On("Execute", onOffMachineContext, OffEvent).Return(nil)

	onOffMachine := brainy.Machine{
		Initial: OnState,

		Context: onOffMachineContext,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
				On: brainy.Events{
					OffEvent: brainy.Transition{
						Target: OffState,
						Actions: brainy.Actions{
							firstTransitionAction.Execute,
							secondTransitionAction.Execute,
							thirdTransitionAction.Execute,
						},
					},
				},
			},

			OffState: brainy.StateNode{
				On: brainy.Events{
					OnEvent: brainy.Transition{
						Target: OnState,
					},
				},
			},
		},
	}
	onOffMachine.Init()

	assert.Equal(OnState, onOffMachine.Current())

	_, err := onOffMachine.Send(OffEvent)
	assert.NoError(err)

	assert.Equal(OffState, onOffMachine.Current())
	firstTransitionAction.AssertExpectations(t)
	secondTransitionAction.AssertExpectations(t)
	thirdTransitionAction.AssertExpectations(t)
}
func TestAFailingActionShortCircuitsTransition(t *testing.T) {
	assert := assert.New(t)

	onOffMachineContext := &struct{}{}
	unexpectedError := errors.New("this action must fail")

	firstTransitionAction := new(mocks.Action)
	firstTransitionAction.On("Execute", onOffMachineContext, OffEvent).Return(nil)
	secondTransitionAction := new(mocks.Action)
	secondTransitionAction.On("Execute", onOffMachineContext, OffEvent).Return(unexpectedError)
	thirdTransitionAction := new(mocks.Action)

	onOffMachine := brainy.Machine{
		Initial: OnState,

		Context: onOffMachineContext,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
				On: brainy.Events{
					OffEvent: brainy.Transition{
						Target: OffState,
						Actions: brainy.Actions{
							firstTransitionAction.Execute,
							secondTransitionAction.Execute,
							thirdTransitionAction.Execute,
						},
					},
				},
			},

			OffState: brainy.StateNode{
				On: brainy.Events{
					OnEvent: brainy.Transition{
						Target: OnState,
					},
				},
			},
		},
	}
	onOffMachine.Init()

	assert.Equal(OnState, onOffMachine.Current())

	nextState, err := onOffMachine.Send(OffEvent)
	assert.Equal(OnState, nextState)
	assert.Error(err)
	assert.ErrorIs(err, unexpectedError)

	assert.Equal(OnState, onOffMachine.Current())
	firstTransitionAction.AssertExpectations(t)
	secondTransitionAction.AssertExpectations(t)
	thirdTransitionAction.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

func TestStateMachineWithTransitionsWithoutTargets(t *testing.T) {
	assert := assert.New(t)

	stateMachineContext := IncrementStateMachineContext{
		ToIncrement: 0,
	}

	incrementVariableInContextMachine := brainy.Machine{
		Initial: IncrementState,

		Context: &stateMachineContext,

		StateNodes: brainy.StateNodes{
			IncrementState: brainy.StateNode{
				On: brainy.Events{
					IncrementEventType: brainy.Transitions{
						brainy.Transition{
							Cond: func(c brainy.Context, e brainy.Event) bool {
								return true
							},
							Target: brainy.NoneState,
							Actions: brainy.Actions{
								func(c brainy.Context, e brainy.Event) error {
									ctx := c.(*IncrementStateMachineContext)

									switch ev := e.(type) {
									case IncrementEvent:
										ctx.ToIncrement += ev.IncrementBy
									default:
										ctx.ToIncrement += 1
									}

									return nil
								},
							},
						},
					},
				},
			},
		},
	}
	incrementVariableInContextMachine.Init()

	assert.Equal(0, stateMachineContext.ToIncrement)

	nextState, err := incrementVariableInContextMachine.Send(NewIncrementEvent(2))
	assert.Equal(IncrementState, nextState)
	assert.NoError(err)
	assert.Equal(2, stateMachineContext.ToIncrement)

	nextState, err = incrementVariableInContextMachine.Send(IncrementEventType)
	assert.Equal(IncrementState, nextState)
	assert.NoError(err)
	assert.Equal(3, stateMachineContext.ToIncrement)

	nextState, err = incrementVariableInContextMachine.Send(UnknownEventType)
	assert.Equal(IncrementState, nextState)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)
	assert.Equal(3, stateMachineContext.ToIncrement)
}

func TestOnlyOneTransitionCanBeTaken(t *testing.T) {
	assert := assert.New(t)

	onOffMachineContext := &struct{}{}

	firstGuardCondition := new(mocks.Cond)
	firstGuardCondition.On("Execute", onOffMachineContext, OnEvent).Return(true)
	secondGuardCondition := new(mocks.Cond)
	thirdGuardCondition := new(mocks.Cond)

	onOffMachine := brainy.Machine{
		Initial: OffState,

		Context: onOffMachineContext,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{},

			OffState: brainy.StateNode{
				On: brainy.Events{
					OnEvent: brainy.Transitions{
						{
							Cond:   firstGuardCondition.Execute,
							Target: OnState,
						},
						{
							Cond:   secondGuardCondition.Execute,
							Target: OnState,
						},
						{
							Cond:   thirdGuardCondition.Execute,
							Target: OnState,
						},
					},
				},
			},
		},
	}
	onOffMachine.Init()

	assert.Equal(OffState, onOffMachine.Current())

	_, err := onOffMachine.Send(OnEvent)
	assert.NoError(err)

	assert.Equal(OnState, onOffMachine.Current())
	firstGuardCondition.AssertExpectations(t)
	secondGuardCondition.AssertNumberOfCalls(t, "Execute", 0)
	thirdGuardCondition.AssertNumberOfCalls(t, "Execute", 0)
}

func TestOnEntryThenActionsThenOnExitAreCalled(t *testing.T) {
	assert := assert.New(t)

	onOffMachineContext := &struct{}{}

	actionsCallsOrder := []string{}

	onExitOffStateAction := new(mocks.Action)
	onExitOffStateAction.On("Execute", onOffMachineContext, OnEvent).Return(nil).Run(func(args mock.Arguments) {
		actionsCallsOrder = append(actionsCallsOrder, "onExit")
	})

	transitionAction := new(mocks.Action)
	transitionAction.On("Execute", onOffMachineContext, OnEvent).Return(nil).Run(func(args mock.Arguments) {
		actionsCallsOrder = append(actionsCallsOrder, "action")
	})

	onEnterOnStateAction := new(mocks.Action)
	onEnterOnStateAction.On("Execute", onOffMachineContext, OnEvent).Return(nil).Run(func(args mock.Arguments) {
		actionsCallsOrder = append(actionsCallsOrder, "onEnter")
	})

	onOffMachine := brainy.Machine{
		Initial: OffState,

		Context: onOffMachineContext,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
				OnEntry: brainy.Actions{
					onEnterOnStateAction.Execute,
				},
			},

			OffState: brainy.StateNode{
				OnExit: brainy.Actions{
					onExitOffStateAction.Execute,
				},

				On: brainy.Events{
					OnEvent: brainy.Transition{
						Target: OnState,
						Actions: brainy.Actions{
							transitionAction.Execute,
						},
					},
				},
			},
		},
	}
	onOffMachine.Init()

	assert.Equal(OffState, onOffMachine.Current())

	_, err := onOffMachine.Send(OnEvent)
	assert.NoError(err)

	assert.Equal(OnState, onOffMachine.Current())
	onExitOffStateAction.AssertExpectations(t)
	transitionAction.AssertExpectations(t)
	onEnterOnStateAction.AssertExpectations(t)
	assert.Equal([]string{
		"onExit",
		"action",
		"onEnter",
	}, actionsCallsOrder)
}

func TestFailingEntryActionAbortsTransition(t *testing.T) {
	assert := assert.New(t)

	onOffMachineContext := &struct{}{}
	failingOnExitTransitionError := errors.New("this action must fail")

	failingOnEntryTransition := new(mocks.Action)
	failingOnEntryTransition.On("Execute", onOffMachineContext, OnEvent).Return(failingOnExitTransitionError)

	onOffMachine := brainy.Machine{
		Initial: OffState,

		Context: onOffMachineContext,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
				OnEntry: brainy.Actions{
					failingOnEntryTransition.Execute,
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

	assert.Equal(OffState, onOffMachine.Current())

	nextState, err := onOffMachine.Send(OnEvent)
	assert.Equal(OffState, nextState)
	assert.Error(err)
	assert.ErrorIs(err, failingOnExitTransitionError)

	assert.Equal(OffState, onOffMachine.Current())
	failingOnEntryTransition.AssertExpectations(t)
}

func TestPreemptivelyValidatesTransitions(t *testing.T) {
	assert := assert.New(t)

	invalidStateMachine := brainy.Machine{
		Initial: OffState,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
				On: brainy.Events{
					OffEvent: IncrementState,
				},
			},

			OffState: brainy.StateNode{
				On: brainy.Events{
					OnEvent: OnState,
				},
			},
		},
	}
	err := invalidStateMachine.Init()
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)

	nextState, err := invalidStateMachine.Send(OnEvent)
	assert.Equal(brainy.NoneState, nextState)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)

	currentState := invalidStateMachine.Current()
	assert.Equal(brainy.NoneState, currentState)

	currentState = invalidStateMachine.UnsafeCurrent()
	assert.Equal(brainy.NoneState, currentState)

	previousState := invalidStateMachine.Current()
	assert.Equal(brainy.NoneState, previousState)
}
