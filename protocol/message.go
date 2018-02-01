package protocol

// Message is interface of bitcoin message.
type Message interface {
	CommandName() string
	Encode() []byte
}
