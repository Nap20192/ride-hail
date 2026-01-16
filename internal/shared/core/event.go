package core

type Event interface {
	Name() string
	Payload() []byte
}

type ShutdownEvent struct{}

func (s ShutdownEvent) Name() string {
	return "shutdown"
}

func (s ShutdownEvent) Payload() []byte {
	return []byte{}
}
