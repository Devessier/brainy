# brainy

Simple state machine library for Golang.

> ⚠️ Disclaimer ⚠️
> 
> The code and the API of this library are heavily inspired by this great article: https://venilnoronha.io/a-simple-state-machine-framework-in-go.

The objective of this library is to have a minimal port of [XState](https://github.com/davidkpiano/xstate) to Golang.

## Quick start

```bash
go get -u github.com/Devessier/brainy
```

```go
package main

import (
	"fmt"

	"github.com/Devessier/brainy"
)

func main() {
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
}
```
