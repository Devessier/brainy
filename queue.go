package brainy

type eventsQueue struct {
	events []Event
}

func newEventsQueue() *eventsQueue {
	return &eventsQueue{
		events: make([]Event, 0),
	}
}

func (q *eventsQueue) Add(event Event) {
	q.events = append(q.events, event)
}

func (q *eventsQueue) Poll() (Event, bool) {
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
