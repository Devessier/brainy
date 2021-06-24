package brainy_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Devessier/brainy"
)

func TestOnOffMachine(t *testing.T) {
	assert := assert.New(t)

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

	assert.Equal(OffState, onOffMachine.Current())

	nextState, err := onOffMachine.Send(OnEvent)
	assert.NoError(err)
	assert.Equal(OnState, nextState)
	assert.Equal(OnState, onOffMachine.Current())

	nextState, err = onOffMachine.Send(OnEvent)
	assert.Error(err)
	assert.ErrorIs(err, brainy.ErrInvalidTransitionNotImplemented)

	nextState, err = onOffMachine.Send(OffEvent)
	assert.NoError(err)
	assert.Equal(OffState, nextState)
	assert.Equal(OffState, onOffMachine.Current())
}
