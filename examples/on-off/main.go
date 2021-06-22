package main

import (
	"fmt"

	"github.com/Devessier/brainy"
)

func main() {
	// @@@SNIPSTART on-off-example
	const (
		OnState  brainy.StateType = "on"
		OffState brainy.StateType = "off"

		OnEvent  brainy.EventType = "on"
		OffEvent brainy.EventType = "off"
	)

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

	initialState := onOffMachine.Current() // off
	fmt.Printf("initial state of the state machine is: %s\n", initialState)

	onOffMachine.Send(OnEvent)

	stateAfterOnEvent := onOffMachine.Current() // on
	fmt.Printf("state of the state machine after receiving an on event from off state is: %s\n", stateAfterOnEvent)

	onOffMachine.Send(OnEvent)

	stateAfterASecondOnEvent := onOffMachine.Current() // on
	fmt.Printf("state of the state machine after receiving an on event from on state is: %s\n", stateAfterASecondOnEvent)
	// @@@SNIPEND
}
