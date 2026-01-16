package core

type AggregateRoot struct {
	Events []Event
}

func (ar *AggregateRoot) ApplyEvent(e Event) {
	ar.Events = append(ar.Events, e)
}

func (ar *AggregateRoot) DomainEvents() []Event {
	return ar.Events
}
