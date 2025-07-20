package protobuffy

import (
	"net"
)

type ByteMessage struct {
	address net.Addr
	data    []byte
}

// NewMessage returns a new *ByteMessage with the specified fields.
func NewByteMessage(address net.Addr, data []byte) (msg *ByteMessage) {
	msg = &ByteMessage{address: address, data: data}
	return
}

// Address returns the address the message is from or to, or nil if msg is nil.
func (msg *ByteMessage) Address() (addr net.Addr) {
	if msg != nil {
		addr = msg.address
	}
	return
}

// Data returns the message's data, or nil if msg is nil.
func (msg *ByteMessage) Data() (data []byte) {
	if msg != nil {
		data = msg.data
	}
	return
}
