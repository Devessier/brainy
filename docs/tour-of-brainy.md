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
    fmt.Printf("The current state of the state machine is: %s\n", currentState) // off
}
```

## Events

We saw how to use states to define which positions a state machine can take. To go from one state to another one, to _transition_ from one state to another one, we need another tool: events.

To transition from one state to another one, we need to tell brainy which events are expected to be received for each state. By default, if an event is received and it has not been defined for the current state, nothing happens.

Let's define a toggle event to go from `on` to `off` states, and vice versa.

### Define events

With brainy, to define the events we want to catch in a given state, we need to list them in `On` map and say to which state to transition. We must use `brainy.Events` type that is a `map[brainy.EventType]brainy.StateType`. `brainy.EventType` is a string type, and we can use string literals during prototyping.

```go{14-23}
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
                On: brainy.Events{
                    "toggle": "off",
                },
            },
            "off": brainy.StateNode{
                On: brainy.Events{
                    "toggle": "on",
                },
            },
        },
    }
    lightSwitchStateMachine.Init()

    currentState := lightSwitchStateMachine.Current()
    fmt.Printf("The current state of the state machine is: %s\n", currentState) // off
}
```

### Send events

Now that we defined the transitions, we need to figure out how to send these events to the state machine.

To send an event to the state machine, we call the `.Send()` method with the event we want to send.

```go{30-34}
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
                On: brainy.Events{
                    "toggle": "off",
                },
            },
            "off": brainy.StateNode{
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
```
#### Sending an unknown event

The `.Send()` method returns two things: the state reached after the event has been received, and an error that could occur during the transition.

In general we do not need to take care of these two returns, but the error can be used to know if the event we sent was handled by the state we were in. We can make use of [errors.Is](https://golang.org/pkg/errors/#Is) to determine if the returned error means the transition was impossible.

```go
import "errors"

nextState, err := lightSwitchStateMachine.Send("toogle")
if err != nil {
    // An error occured during the transition
    if errors.Is(err, brainy.ErrInvalidTransitionNotImplemented) {
        // The event `toggle` is not handled by the state we are in.

        return
    }

    // Another error occured.
    //
    // The only reason an error can be raised in a transition, except that the transition was not implemented,
    // is that an error occured in an `action` of the state we reached.
    // We will see `actions` in the following section.
}

```

When an unknown event is received, the state of the state machine remains the same. The philosophy behind that is that with state machines, we need to explicitly write what can occur. We must write which states are possible, which events lead to which events. Anything that has not been explicitly described should never modify the state machine.
