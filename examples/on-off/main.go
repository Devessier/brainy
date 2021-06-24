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
	lightSwitchStateMachine := brainy.Machine{
		Initial: OffState,

		StateNodes: brainy.StateNodes{
			OnState: brainy.StateNode{
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

			OffState: brainy.StateNode{
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
	}
	lightSwitchStateMachine.Init()

	currentState := lightSwitchStateMachine.Current()
	fmt.Printf("The current state of the state machine is: %s\n", currentState) // off

	lightSwitchStateMachine.Send(ToggleEvent)

	stateAfterFirstToggle := lightSwitchStateMachine.Current()
	fmt.Printf("The state of the state machine after the first toogle is: %s\n", stateAfterFirstToggle) // on

	lightSwitchStateMachine.Send(ToggleEvent)

	stateAfterSecondToggle := lightSwitchStateMachine.Current()
	fmt.Printf("The state of the state machine after the second toogle is: %s\n", stateAfterSecondToggle) // off
}

// @@@SNIPEND
