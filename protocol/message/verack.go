package message

// Verack means verack message
type Verack struct{}

// CommandName return "verack".
func (v *Verack) CommandName() string {
	return "verack"
}

// Encode encode verack.
func (v *Verack) Encode() []byte {
	return []byte{}
}
