// @@@SNIPSTART on-off-example
package main

import (
	"fmt"

	"github.com/Devessier/brainy"
)

const (
	OnState  brainy.StateType = "on"
	OffState brainy.StateType = "off"

	ToggleEvent brainy.EventType = "TOGGLE"
)

func main() {
	lightSwitchStateMachine, err := brainy.NewMachine(brainy.StateNode{
		Initial: OffState,

		States: brainy.StateNodes{
			OnState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					func(c brainy.Context, e brainy.Event) error {
						fmt.Println("Reached `on` state")

						return nil
					},
				},

				On: brainy.Events{
					ToggleEvent: brainy.Transition{
						Target: OffState,
						Actions: brainy.Actions{
							func(c brainy.Context, e brainy.Event) error {
								fmt.Println("Go to `off` state")

								return nil
							},
						},
					},
				},
			},

			OffState: &brainy.StateNode{
				OnEntry: brainy.Actions{
					func(c brainy.Context, e brainy.Event) error {
						fmt.Println("Reached `off` state")

						return nil
					},
				},

				On: brainy.Events{
					ToggleEvent: brainy.Transition{
						Target: OnState,
						Actions: brainy.Actions{
							func(c brainy.Context, e brainy.Event) error {
								fmt.Println("Go to `on` state")

								return nil
							},
						},
					},
				},
			},
		},
	})
	if err != nil {
		fmt.Printf("invalid state machine declaration: %s\n", err)
		return
	}

	currentState := lightSwitchStateMachine.Current()
	fmt.Printf("The current state of the state machine is: %s\n", currentState.Value()) // (machine).off
	fmt.Printf("The light switch is off: %v\n", currentState.Matches(OffState))         // true

	lightSwitchStateMachine.Send(ToggleEvent)

	stateAfterFirstToggle := lightSwitchStateMachine.Current()
	fmt.Printf("The state of the state machine after the first toogle is: %s\n", stateAfterFirstToggle.Value()) // (machine).on
	fmt.Printf("The light switch is on: %v\n", stateAfterFirstToggle.Matches(OnState))                          // true

	lightSwitchStateMachine.Send(ToggleEvent)

	stateAfterSecondToggle := lightSwitchStateMachine.Current()
	fmt.Printf("The state of the state machine after the second toogle is: %s\n", stateAfterSecondToggle.Value()) // (machine).off
	fmt.Printf("Is the light switch on? %v\n", stateAfterSecondToggle.Matches(OnState))                           // false

}

// @@@SNIPEND
