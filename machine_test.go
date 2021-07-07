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

	assert.True(onOffMachine.Current().Matches(OffState))

	nextState, err := onOffMachine.Send(OnEvent)
	assert.NoError(err)
	assert.True(nextState.Matches(OnState))
	assert.True(onOffMachine.Current().Matches(OnState))

	nextState, err = onOffMachine.Send(OnEvent)
	assert.True(nextState.Matches(OnState))
	assert.Error(err)
	assert.ErrorIs(err, &brainy.ErrNoHandlerToHandleEvent{
		Event: OnEvent,
	})

	nextState, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)
	assert.True(nextState.Matches(OffState))
	assert.True(onOffMachine.Current().Matches(OffState))
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

	assert.True(onOffMachine.Current().Matches(OffState))

	onOffMachine.Send(OnEvent)

	assert.True(onOffMachine.Current().Matches(OnState))
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

	assert.True(onOffMachine.Current().Matches(OnState))

	_, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)

	assert.True(onOffMachine.Current().Matches(OffState))

	_, err = onOffMachine.Send(OnEvent)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrNoTransitionCouldBeRun)

	assert.True(onOffMachine.Current().Matches(OffState))
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
							brainy.ActionFn(firstTransitionAction.Execute),
							brainy.ActionFn(secondTransitionAction.Execute),
							brainy.ActionFn(thirdTransitionAction.Execute),
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

	assert.True(onOffMachine.Current().Matches(OnState))

	_, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)

	assert.True(onOffMachine.Current().Matches(OffState))
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
							brainy.ActionFn(firstTransitionAction.Execute),
							brainy.ActionFn(secondTransitionAction.Execute),
							brainy.ActionFn(thirdTransitionAction.Execute),
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

	assert.True(onOffMachine.Current().Matches(OnState))

	nextState, err := onOffMachine.Send(OffEvent)
	assert.True(nextState.Matches(OnState))
	assert.Error(err)
	assert.ErrorIs(err, unexpectedError)

	assert.True(onOffMachine.Current().Matches(OnState))
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
							Actions: brainy.Actions{
								brainy.ActionFn(
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
								),
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
	assert.True(nextState.Matches(IncrementState))
	assert.NoError(err)
	assert.Equal(2, stateMachineContext.ToIncrement)

	nextState, err = incrementVariableInContextMachine.Send(IncrementEventType)
	assert.NotNil(nextState)
	assert.True(nextState.Matches(IncrementState))
	assert.NoError(err)
	assert.Equal(3, stateMachineContext.ToIncrement)

	nextState, err = incrementVariableInContextMachine.Send(UnknownEventType)
	assert.NotNil(nextState)
	assert.True(nextState.Matches(IncrementState))
	assert.Error(err)
	assert.ErrorIs(err, &brainy.ErrNoHandlerToHandleEvent{
		Event: UnknownEventType,
	})
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

	assert.True(onOffMachine.Current().Matches(OffState))

	_, err = onOffMachine.Send(OnEvent)
	assert.NoError(err)

	assert.True(onOffMachine.Current().Matches(OnState))
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
					brainy.ActionFn(onEnterOnStateAction.Execute),
				},
			},

			OffState: &brainy.StateNode{
				OnExit: brainy.Actions{
					brainy.ActionFn(onExitOffStateAction.Execute),
				},

				On: brainy.Events{
					OnEvent: brainy.Transition{
						Target: OnState,
						Actions: brainy.Actions{
							brainy.ActionFn(transitionAction.Execute),
						},
					},
				},
			},
		},
	})
	assert.NoError(err)

	assert.True(onOffMachine.Current().Matches(OffState))

	_, err = onOffMachine.Send(OnEvent)
	assert.NoError(err)

	assert.True(onOffMachine.Current().Matches(OnState))
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
					brainy.ActionFn(failingOnEntryTransition.Execute),
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

	assert.True(onOffMachine.Current().Matches(OffState))

	nextState, err := onOffMachine.Send(OnEvent)
	assert.True(nextState.Matches(OffState))
	assert.Error(err)
	assert.ErrorIs(err, failingOnExitTransitionError)

	assert.True(onOffMachine.Current().Matches(OffState))
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

const (
	CompoundState brainy.StateType = "compound"
	NestedAState  brainy.StateType = "nested-a"
	NestedBState  brainy.StateType = "nested-b"
	AtomicState   brainy.StateType = "atomic"

	GoToNestedBStateEvent  brainy.EventType = "GO_TO_NESTED_B_STATE"
	ExitCompoundStateEvent brainy.EventType = "EXIT_COMPOUND_STATE"
)

func TestCompoundStates(t *testing.T) {
	assert := assert.New(t)

	const (
		RootOnEntry string = "RootOnEntry"
		RootOnExit  string = "RootOnExit"

		CompoundStateOnEntry string = "CompoundStateOnEntry"
		CompoundStateOnExit  string = "CompoundStateOnExit"

		NestedAStateOnEntry string = "NestedAStateOnEntry"
		NestedAStateOnExit  string = "NestedAStateOnExit"

		NestedBStateOnEntry string = "NestedBStateOnEntry"
		NestedBStateOnExit  string = "NestedBStateOnExit"

		AtomicStateOnEntry string = "AtomicStateOnEntry"
		AtomicStateOnExit  string = "AtomicStateOnExit"
	)

	calledActions := make([]string, 0)

	rootOnEntryAction := new(mocks.Action)
	rootOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, RootOnEntry)
	})
	rootOnExitAction := new(mocks.Action)
	rootOnExitAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, RootOnExit)
	})

	compoundStateOnEntryAction := new(mocks.Action)
	compoundStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, CompoundStateOnEntry)
	})
	compoundStateOnExitAction := new(mocks.Action)
	compoundStateOnExitAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, CompoundStateOnExit)
	})

	nestedAStateStateOnEntryAction := new(mocks.Action)
	nestedAStateStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, NestedAStateOnEntry)
	})
	nestedAStateStateOnExitAction := new(mocks.Action)
	nestedAStateStateOnExitAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, NestedAStateOnExit)
	})

	nestedBStateStateOnEntryAction := new(mocks.Action)
	nestedBStateStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, NestedBStateOnEntry)
	})
	nestedBStateStateOnExitAction := new(mocks.Action)
	nestedBStateStateOnExitAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, NestedBStateOnExit)
	})

	atomicStateOnEntryAction := new(mocks.Action)
	atomicStateOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, AtomicStateOnEntry)
	})
	atomicStateOnExitAction := new(mocks.Action)
	atomicStateOnExitAction.On("Execute", mock.Anything, mock.Anything).Return(nil).Run(func(args mock.Arguments) {
		calledActions = append(calledActions, AtomicStateOnExit)
	})

	compoundStateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: CompoundState,

		OnEntry: brainy.Actions{
			brainy.ActionFn(rootOnEntryAction.Execute),
		},

		OnExit: brainy.Actions{
			brainy.ActionFn(rootOnExitAction.Execute),
		},

		States: brainy.StateNodes{
			CompoundState: &brainy.StateNode{
				Initial: NestedAState,

				OnEntry: brainy.Actions{
					brainy.ActionFn(compoundStateOnEntryAction.Execute),
				},

				OnExit: brainy.Actions{
					brainy.ActionFn(compoundStateOnExitAction.Execute),
				},

				States: brainy.StateNodes{
					NestedAState: &brainy.StateNode{
						OnEntry: brainy.Actions{
							brainy.ActionFn(nestedAStateStateOnEntryAction.Execute),
						},

						OnExit: brainy.Actions{
							brainy.ActionFn(nestedAStateStateOnExitAction.Execute),
						},
					},

					NestedBState: &brainy.StateNode{
						OnEntry: brainy.Actions{
							brainy.ActionFn(nestedBStateStateOnEntryAction.Execute),
						},

						OnExit: brainy.Actions{
							brainy.ActionFn(nestedBStateStateOnExitAction.Execute),
						},
					},
				},

				On: brainy.Events{
					GoToNestedBStateEvent: brainy.CompoundTarget{
						CompoundState: NestedBState,
					},

					ExitCompoundStateEvent: AtomicState,
				},
			},

			AtomicState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					brainy.ActionFn(atomicStateOnEntryAction.Execute),
				},

				OnExit: brainy.Actions{
					brainy.ActionFn(atomicStateOnExitAction.Execute),
				},
			},
		},
	})
	assert.NotNil(compoundStateMachine)
	assert.NoError(err)

	compoundStateOnEntryAction.AssertCalled(t, "Execute", nil, brainy.InitialTransitionEventType)
	nestedAStateStateOnEntryAction.AssertCalled(t, "Execute", nil, brainy.InitialTransitionEventType)
	assert.True(compoundStateMachine.Current().Matches(CompoundState, NestedAState))

	nextState, err := compoundStateMachine.Send(GoToNestedBStateEvent)
	assert.NoError(err)
	assert.True(nextState.Matches(CompoundState, NestedBState))

	nextState, err = compoundStateMachine.Send(ExitCompoundStateEvent)
	assert.NoError(err)
	assert.True(nextState.Matches(AtomicState))

	assert.Equal([]string{
		RootOnEntry,
		CompoundStateOnEntry,
		NestedAStateOnEntry,

		NestedAStateOnExit,
		// We do not handle onentry and onexit states as described in SCXML specification.
		// We choose to do not take care of that for now.
		//
		// CompoundStateOnExit,
		// CompoundStateOnEntry,
		NestedBStateOnEntry,

		NestedBStateOnExit,
		CompoundStateOnExit,
		AtomicStateOnEntry,
	}, calledActions)
	compoundStateOnEntryAction.AssertExpectations(t)
	nestedAStateStateOnEntryAction.AssertExpectations(t)
	nestedBStateStateOnEntryAction.AssertExpectations(t)
	atomicStateOnEntryAction.AssertExpectations(t)
}

func TestCanSendEventsWithSendAction(t *testing.T) {
	assert := assert.New(t)

	const (
		SayHelloEvent brainy.EventType = "SAY_HELLO"
	)

	compoundStateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: CompoundState,

		States: brainy.StateNodes{
			CompoundState: &brainy.StateNode{
				Initial: NestedAState,

				States: brainy.StateNodes{
					NestedAState: &brainy.StateNode{
						On: brainy.Events{
							SayHelloEvent: brainy.Transition{
								Actions: brainy.Actions{
									brainy.Send(ExitCompoundStateEvent),
								},
							},
						},
					},
				},

				On: brainy.Events{
					ExitCompoundStateEvent: AtomicState,
				},
			},

			AtomicState: &brainy.StateNode{},
		},
	})
	assert.NotNil(compoundStateMachine)
	assert.NoError(err)

	assert.True(compoundStateMachine.Current().Matches(CompoundState, NestedAState))

	nextState, err := compoundStateMachine.Send(SayHelloEvent)
	assert.NoError(err)
	assert.True(nextState.Matches(AtomicState))
	assert.True(compoundStateMachine.Current().Matches(AtomicState))
}

func TestCanDisableLocking(t *testing.T) {
	assert := assert.New(t)

	stateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OnState,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{},
		},
	}, brainy.WithDisableLocking())
	assert.NotNil(stateMachine)
	assert.NoError(err)
}

func TestGivesGlobalContextDuringInitialTransition(t *testing.T) {
	assert := assert.New(t)

	stateMachineContext := &struct{}{}

	initialStateOnEntryAction := new(mocks.Action)
	initialStateOnEntryAction.On("Execute", stateMachineContext, brainy.InitialTransitionEventType).Return(nil)

	stateMachine, err := brainy.NewMachine(brainy.StateNode{
		Context: stateMachineContext,

		Initial: OnState,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					brainy.ActionFn(
						initialStateOnEntryAction.Execute,
					),
				},
			},
		},
	}, brainy.WithDisableLocking())
	assert.NotNil(stateMachine)
	assert.NoError(err)

	initialStateOnEntryAction.AssertExpectations(t)
}

func TestBlankTargetTriggersNoEntryActionsNorExitActions(t *testing.T) {
	assert := assert.New(t)

	const (
		InternalEvent brainy.EventType = "INTERNAL"
	)

	rootOnEntryAction := new(mocks.Action)
	rootOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil)

	rootOnExitAction := new(mocks.Action)
	rootOnExitAction.On("Execute", mock.Anything, mock.Anything).Return(nil)

	atomicOnEntryAction := new(mocks.Action)
	atomicOnEntryAction.On("Execute", mock.Anything, mock.Anything).Return(nil)

	atomicOnExitAction := new(mocks.Action)
	atomicOnExitAction.On("Execute", mock.Anything, mock.Anything).Return(nil)

	internalEventTransitionAction := new(mocks.Action)
	internalEventTransitionAction.On("Execute", mock.Anything, mock.Anything).Return(nil)

	stateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: AtomicState,

		OnEntry: brainy.Actions{
			brainy.ActionFn(rootOnEntryAction.Execute),
		},

		OnExit: brainy.Actions{
			brainy.ActionFn(rootOnExitAction.Execute),
		},

		States: brainy.StateNodes{
			AtomicState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					brainy.ActionFn(atomicOnEntryAction.Execute),
				},

				OnExit: brainy.Actions{
					brainy.ActionFn(atomicOnExitAction.Execute),
				},

				On: brainy.Events{
					InternalEvent: brainy.Transition{
						Target: nil,
						Actions: brainy.Actions{
							brainy.ActionFn(internalEventTransitionAction.Execute),
						},
					},
				},
			},
		},
	})
	assert.NotNil(stateMachine)
	assert.NoError(err)

	assert.True(stateMachine.Current().Matches(AtomicState))

	nextState, err := stateMachine.Send(InternalEvent)
	assert.NoError(err)
	assert.True(nextState.Matches(AtomicState))
	assert.True(stateMachine.Current().Matches(AtomicState))

	rootOnEntryAction.AssertExpectations(t)
	rootOnEntryAction.AssertNumberOfCalls(t, "Execute", 1)
	rootOnExitAction.AssertNumberOfCalls(t, "Execute", 0)
	atomicOnEntryAction.AssertExpectations(t)
	atomicOnEntryAction.AssertNumberOfCalls(t, "Execute", 1)
	atomicOnExitAction.AssertNumberOfCalls(t, "Execute", 0)
	internalEventTransitionAction.AssertExpectations(t)
	internalEventTransitionAction.AssertNumberOfCalls(t, "Execute", 1)
}

// Test that children states can only be targeted using a CompoundTarget.
// This is due to the fact that transitions can only occur between sibbling states.
// To target a child state, we must first target the state itself, that is itself sibbling,
// and then target the child state.
func TestResolvesChildTransitionsCorrectly(t *testing.T) {
	assert := assert.New(t)

	invalidCompoundStateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: CompoundState,

		States: brainy.StateNodes{
			CompoundState: &brainy.StateNode{
				Initial: NestedAState,

				States: brainy.StateNodes{
					NestedAState: &brainy.StateNode{},

					NestedBState: &brainy.StateNode{},
				},

				On: brainy.Events{
					GoToNestedBStateEvent: brainy.Transition{
						Target: NestedBState,
					},
				},
			},
		},
	})
	assert.Nil(invalidCompoundStateMachine)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)

	validCompoundStateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: CompoundState,

		States: brainy.StateNodes{
			CompoundState: &brainy.StateNode{
				Initial: NestedAState,

				States: brainy.StateNodes{
					NestedAState: &brainy.StateNode{},

					NestedBState: &brainy.StateNode{},
				},

				On: brainy.Events{
					GoToNestedBStateEvent: brainy.Transition{
						Target: brainy.CompoundTarget{
							CompoundState: NestedBState,
						},
					},
				},
			},
		},
	})
	assert.NotNil(validCompoundStateMachine)
	assert.NoError(err)
}
