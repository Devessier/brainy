# brainy

Simple state machine library for Golang, that aims to be compliant with [SCXML specification](https://www.w3.org/TR/scxml/), with an API similar to the amazing [XState](https://github.com/davidkpiano/xstate) library.

## Quick start

```bash
go get -u github.com/Devessier/brainy
```

<!--SNIPSTART on-off-example-->
```go
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

```
<!--SNIPEND-->

## Credits

The original implementation of this library was based on this very great article: https://venilnoronha.io/a-simple-state-machine-framework-in-go.
