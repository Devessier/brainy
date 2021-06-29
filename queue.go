package brainy

type EventsQueue struct {
	events []Event
}

func NewEventsQueue() *EventsQueue {
	return &EventsQueue{
		events: make([]Event, 0),
	}
}

func (q *EventsQueue) Add(event Event) {
	q.events = append(q.events, event)
}

func (q *EventsQueue) Poll() (Event, bool) {
	eventsCount := len(q.events)
	if eventsCount == 0 {
		return nil, false
	}

	event := q.events[0]

	if eventsCount == 1 {
		q.events = make([]Event, 0)
	} else {
		q.events = q.events[1:]
	}

	return event, true
}
