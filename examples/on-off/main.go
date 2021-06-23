// @@@SNIPSTART on-off-example
package main

import (
	"fmt"

	"github.com/Devessier/brainy"
)

func main() {
	lightSwitchStateMachine := brainy.Machine{
		Initial: "off",

		StateNodes: brainy.StateNodes{
			"on": brainy.StateNode{
				Actions: brainy.Actions{
					func(m *brainy.Machine, c brainy.Context) (brainy.EventType, error) {
						fmt.Println("Reached on state")

						return brainy.NoopEvent, nil
					},
				},

				On: brainy.Events{
					"toggle": "off",
				},
			},
			"off": brainy.StateNode{
				Actions: brainy.Actions{
					func(m *brainy.Machine, c brainy.Context) (brainy.EventType, error) {
						fmt.Println("Reached off state")

						return brainy.NoopEvent, nil
					},
				},

				On: brainy.Events{
					"toggle": "on",
				},
			},
		},
	}
	lightSwitchStateMachine.Init()

	currentState := lightSwitchStateMachine.Current()
	fmt.Printf("The current state of the state machine is: %s\n", currentState) // off

	lightSwitchStateMachine.Send("toogle")

	stateAfterFirstToggle := lightSwitchStateMachine.Current()
	fmt.Printf("The state of the state machine after the first toogle is: %s\n", stateAfterFirstToggle) // on

	lightSwitchStateMachine.Send("toogle")

	stateAfterSecondToggle := lightSwitchStateMachine.Current()
	fmt.Printf("The state of the state machine after the second toogle is: %s\n", stateAfterSecondToggle) // off
}

// @@@SNIPEND
