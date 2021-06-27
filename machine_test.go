package brainy_test

import (
	"errors"
	"testing"

	"github.com/Devessier/brainy"
	"github.com/Devessier/brainy/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

const (
	OnState  brainy.StateType = "on"
	OffState brainy.StateType = "off"

	OnEvent  brainy.EventType = "on"
	OffEvent brainy.EventType = "off"
)

func TestOnOffMachine(t *testing.T) {
	assert := assert.New(t)

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OffState,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				On: brainy.Events{
					OffEvent: OffState,
				},
			},

			OffState: &brainy.StateNode{
				On: brainy.Events{
					OnEvent: OnState,
				},
			},
		},
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), OffState.String())

	nextState, err := onOffMachine.Send(OnEvent)
	assert.NoError(err)
	assert.Contains(nextState.Value(), OnState.String())
	assert.Contains(onOffMachine.Current().Value(), OnState.String())

	nextState, err = onOffMachine.Send(OnEvent)
	assert.Contains(nextState.Value(), OnState.String())
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)

	nextState, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)
	assert.Contains(nextState.Value(), OffState.String())
	assert.Contains(onOffMachine.Current().Value(), OffState.String())
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

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OffState,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				On: brainy.Events{
					OffEvent: OffState,
				},
			},

			OffState: &brainy.StateNode{
				On: brainy.Events{
					OnEvent: OnState,
				},
			},
		},
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))

	onOffMachine.Send(OnEvent)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OnState))
}

func TestCondTakesAGuardFunction(t *testing.T) {
	assert := assert.New(t)

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OnState,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				On: brainy.Events{
					OffEvent: brainy.Transition{
						Cond: func(c brainy.Context, e brainy.Event) bool {
							return true
						},
						Target: OffState,
					},
				},
			},

			OffState: &brainy.StateNode{
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
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OnState))

	_, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))

	_, err = onOffMachine.Send(OnEvent)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))
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

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OnState,

		Context: onOffMachineContext,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
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

			OffState: &brainy.StateNode{
				On: brainy.Events{
					OnEvent: brainy.Transition{
						Target: OnState,
					},
				},
			},
		},
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OnState))

	_, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))
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

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OnState,

		Context: onOffMachineContext,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
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

			OffState: &brainy.StateNode{
				On: brainy.Events{
					OnEvent: brainy.Transition{
						Target: OnState,
					},
				},
			},
		},
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OnState))

	nextState, err := onOffMachine.Send(OffEvent)
	assert.Contains(nextState.Value(), brainy.JoinStatesIDs(OnState))
	assert.Error(err)
	assert.ErrorIs(err, unexpectedError)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OnState))
	firstTransitionAction.AssertExpectations(t)
	secondTransitionAction.AssertExpectations(t)
	thirdTransitionAction.AssertNotCalled(t, "Execute", mock.Anything, mock.Anything)
}

func TestStateMachineWithTransitionsWithoutTargets(t *testing.T) {
	assert := assert.New(t)

	stateMachineContext := IncrementStateMachineContext{
		ToIncrement: 0,
	}

	incrementVariableInContextMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: IncrementState,

		Context: &stateMachineContext,

		States: brainy.StateNodes{
			IncrementState: &brainy.StateNode{
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
	})
	assert.NoError(err)

	assert.Equal(0, stateMachineContext.ToIncrement)

	nextState, err := incrementVariableInContextMachine.Send(NewIncrementEvent(2))
	assert.NotNil(nextState)
	assert.Contains(nextState.Value(), brainy.JoinStatesIDs(IncrementState))
	assert.NoError(err)
	assert.Equal(2, stateMachineContext.ToIncrement)

	nextState, err = incrementVariableInContextMachine.Send(IncrementEventType)
	assert.NotNil(nextState)
	assert.Contains(nextState.Value(), brainy.JoinStatesIDs(IncrementState))
	assert.NoError(err)
	assert.Equal(3, stateMachineContext.ToIncrement)

	nextState, err = incrementVariableInContextMachine.Send(UnknownEventType)
	assert.NotNil(nextState)
	assert.Contains(nextState.Value(), brainy.JoinStatesIDs(IncrementState))
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

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OffState,

		Context: onOffMachineContext,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{},

			OffState: &brainy.StateNode{
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
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))

	_, err = onOffMachine.Send(OnEvent)
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OnState))
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

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OffState,

		Context: onOffMachineContext,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					onEnterOnStateAction.Execute,
				},
			},

			OffState: &brainy.StateNode{
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
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))

	_, err = onOffMachine.Send(OnEvent)
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OnState))
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

	onOffMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OffState,

		Context: onOffMachineContext,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					failingOnEntryTransition.Execute,
				},
			},

			OffState: &brainy.StateNode{
				On: brainy.Events{
					OnEvent: OnState,
				},
			},
		},
	})
	assert.NoError(err)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))

	nextState, err := onOffMachine.Send(OnEvent)
	assert.Contains(nextState.Value(), brainy.JoinStatesIDs(OffState))
	assert.Error(err)
	assert.ErrorIs(err, failingOnExitTransitionError)

	assert.Contains(onOffMachine.Current().Value(), brainy.JoinStatesIDs(OffState))
	failingOnEntryTransition.AssertExpectations(t)
}

func TestPreemptivelyValidatesTransitions(t *testing.T) {
	assert := assert.New(t)

	invalidStateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OffState,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				On: brainy.Events{
					OffEvent: IncrementState,
				},
			},

			OffState: &brainy.StateNode{
				On: brainy.Events{
					OnEvent: OnState,
				},
			},
		},
	})
	assert.Nil(invalidStateMachine)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)
}

func TestCompoundStates(t *testing.T) {
	assert := assert.New(t)

	const (
		CompoundState brainy.StateType = "compound"
		NestedAState  brainy.StateType = "nested-a"
		NestedBState  brainy.StateType = "nested-b"
		AtomicState   brainy.StateType = "atomic"

		GoToNestedBState  brainy.EventType = "GO_TO_NESTED_B_STATE"
		ExitCompoundState brainy.EventType = "EXIT_COMPOUND_STATE"

		CompoundStateOnEntry string = "CompoundStateOnEntry"
		NestedAStateOnEntry  string = "CompoundStateOnEntry"
		NestedBStateOnEntry  string = "CompoundStateOnEntry"
		AtomicStateOnEntry   string = "AtomicStateOnEntry"
	)

	calledActions := make([]string, 0)

	compoundStateOnEntryAction := new(mocks.Action)
	compoundStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, CompoundStateOnEntry)
	})
	nestedAStateStateOnEntryAction := new(mocks.Action)
	nestedAStateStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, NestedAStateOnEntry)
	})
	nestedBStateStateOnEntryAction := new(mocks.Action)
	nestedBStateStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, NestedBStateOnEntry)
	})
	atomicStateOnEntryAction := new(mocks.Action)
	atomicStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, AtomicStateOnEntry)
	})

	compoundStateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: CompoundState,

		States: brainy.StateNodes{
			CompoundState: &brainy.StateNode{
				Initial: NestedAState,

				OnEntry: brainy.Actions{
					compoundStateOnEntryAction.Execute,
				},

				States: brainy.StateNodes{
					NestedAState: &brainy.StateNode{
						OnEntry: brainy.Actions{
							nestedAStateStateOnEntryAction.Execute,
						},

						On: brainy.Events{
							GoToNestedBState: NestedBState,
						},
					},

					NestedBState: &brainy.StateNode{
						OnEntry: brainy.Actions{
							nestedBStateStateOnEntryAction.Execute,
						},
					},
				},

				On: brainy.Events{
					ExitCompoundState: AtomicState,
				},
			},

			AtomicState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					atomicStateOnEntryAction.Execute,
				},

				On: brainy.Events{
					GoToNestedBState: brainy.CompoundTarget{
						CompoundState: NestedBState,
					},
				},
			},
		},
	})
	assert.NotNil(compoundStateMachine)
	assert.NoError(err)

	compoundStateOnEntryAction.AssertCalled(t, "Execute", nil, brainy.InitialTransitionEventType)
	nestedAStateStateOnEntryAction.AssertCalled(t, "Execute", nil, brainy.InitialTransitionEventType)
	assert.Contains(compoundStateMachine.Current().Value(), brainy.JoinStatesIDs(CompoundState, NestedAState))

	nextState, err := compoundStateMachine.Send(GoToNestedBState)
	assert.NoError(err)
	assert.Contains(nextState.Value(), brainy.JoinStatesIDs(CompoundState, NestedBState))

	nextState, err = compoundStateMachine.Send(ExitCompoundState)
	assert.NoError(err)
	assert.Contains(nextState.Value(), brainy.JoinStatesIDs(AtomicState))

	assert.Equal([]string{
		CompoundStateOnEntry,
		NestedAStateOnEntry,

		CompoundStateOnEntry,
		NestedBStateOnEntry,

		AtomicStateOnEntry,
	}, calledActions)
	compoundStateOnEntryAction.AssertExpectations(t)
	nestedAStateStateOnEntryAction.AssertExpectations(t)
	nestedBStateStateOnEntryAction.AssertExpectations(t)
	atomicStateOnEntryAction.AssertExpectations(t)
}
