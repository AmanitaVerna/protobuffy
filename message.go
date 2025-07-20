package protobuffy

import (
	"net"

	"google.golang.org/protobuf/proto"
)

type Message struct {
	address net.Addr
	msgId   uint32
	data    proto.Message
}

// NewMessage returns a new *Message with the specified fields.
func NewMessage(address net.Addr, msgId uint32, data proto.Message) (msg *Message) {
	msg = &Message{address: address, msgId: msgId, data: data}
	return
}

// Address returns the address the message is from or to, or nil if msg is nil.
func (msg *Message) Address() (addr net.Addr) {
	if msg != nil {
		addr = msg.address
	}
	return
}

// Data returns the message's data, or nil if msg is nil.
func (msg *Message) Data() (data proto.Message) {
	if msg != nil {
		data = msg.data
	}
	return
}

// MsgId returns the message ID, or 0 if msg is nil.
func (msg *Message) MsgId() (msgId uint32) {
	if msg != nil {
		msgId = msg.msgId
	}
	return
}
