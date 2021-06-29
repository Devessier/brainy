package brainy

type sendActionEvent struct {
	SourceEvent Event
}

func (e sendActionEvent) run(Context, Event) error {
	return nil
}

// Send function creates a declarative action, that will send internally the event given
// as parameter to the state machine itself.
//
// Send action does not imperatively send an event to the state machine. It tells brainy to send
// an event to itself when the action must be executed, that is, when brainy finds it in a list of actions.
func Send(event Event) Actioner {
	return sendActionEvent{
		SourceEvent: event,
	}
}
