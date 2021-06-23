# Tour of brainy

Currently brainy has a very simple API surface. You can create finite state machines, that is having states you are transitioning to thanks to events. When you enter in a state, its list of `actions` will be called sequentially.

Let's see what is possible to do with brainy.

## States

States are the base element of a state machine. A state is the position in which a state machine is. State machines can take a finite number of positions, that we need to define ourselves.

The simplest finite state machine is the on/off state machine, that could represent a light switch. This state machine has two states: `on` and `off`.

Let's see how we can implement this state machine with brainy.

### Define states

The first step is to import brainy and to define a brainy state machine, thanks to `Machine` struct.

```go
package main

import "github.com/Devessier/brainy"

func main() {
    lightSwitchStateMachine := brainy.Machine{}
}
```

We can now add the two states to the state machine:

```go{7-10}
package main

import "github.com/Devessier/brainy"

func main() {
    lightSwitchStateMachine := brainy.Machine{
        StateNodes: brainy.StateNodes{
            "on": brainy.StateNode{},
            "off": brainy.StateNode{},
        },
    }
}
```

With brainy we define states through `StateNodes` field. The `brainy.StateNodes` type is a `map[brainy.StateNode]brainy.StateNode`. As `brainy.StateNode` is a custom type that refers to a string, we can use string literals to define our states. Although, outside of prototyping, we would prefer to extract the states names to constants with the type `brainy.StateNode`, so we know instinctively what their purpose is.

### Define initial state

We defined our switch light state machine to have two states, `on` and `off`. But what is the initial state? Is the light on or off by default?

With brainy, we need to define the initial state explicitly. It can be done by using the `Initial` field of `Machine` type. Let's say that by default, the light is off:

```go{7-8}
package main

import "github.com/Devessier/brainy"

func main() {
    lightSwitchStateMachine := brainy.Machine{
        Initial: "off",

        StateNodes: brainy.StateNodes{
            "on": brainy.StateNode{},
            "off": brainy.StateNode{},
        },
    }
}
```

### Call `.Init()` method

For brainy to go directly to the `Initial` state, we need to call the `.Init()` method on the state machine. This method will perform some internal work that is necessary for the state machine to work properly.

```go{14}
package main

import "github.com/Devessier/brainy"

func main() {
    lightSwitchStateMachine := brainy.Machine{
        Initial: "off",

        StateNodes: brainy.StateNodes{
            "on": brainy.StateNode{},
            "off": brainy.StateNode{},
        },
    }
    lightSwitchStateMachine.Init()
}
```

### Access current state

State machines allow to encapsulate logic in a single piece of code, and to describe logical combinations explicitly so that the logical state is never in an unknown state. Although having put logic at a single place is great, we still need to know in which state we are. We still need to know in which state our `lightSwitchStateMachine` is, so that we can update the look of the toggle button linked to it, for example.

With brainy, to know what is the current state of a state machine, we can use the method `.Current()`.

```go{19-21}
package main

import (
    "fmt"

    "github.com/Devessier/brainy"
)

func main() {
    lightSwitchStateMachine := brainy.Machine{
        Initial: "off",

        StateNodes: brainy.StateNodes{
            "on": brainy.StateNode{},
            "off": brainy.StateNode{},
        },
    }
    lightSwitchStateMachine.Init()

    currentState := lightSwitchStateMachine.Current()
    fmt.Printf("The current state of the state machine is: %s\n", currentState)
}
```
