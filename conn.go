package protobuffy

import (
	"encoding/binary"
	"fmt"
	"net"

	"google.golang.org/protobuf/proto"
)

const BufferSize int = 65536

// Conn represents a connection to a remote client.
type Conn struct {
	nconn     net.Conn
	addr      net.Addr
	closeChan chan<- net.Addr
	buff      []byte
	msgChan   chan<- *Message
	gnmc      func(messageType uint32) proto.Message
}

func newConn(msgChan chan<- *Message, closeChan chan<- net.Addr, nconn net.Conn, gnmc func(messageType uint32) proto.Message) (conn *Conn) {
	if nconn != nil {
		conn = &Conn{msgChan: msgChan, closeChan: closeChan, nconn: nconn, addr: nconn.RemoteAddr(), buff: make([]byte, BufferSize), gnmc: gnmc}
	}
	return
}

// Close closes the connection.
func (conn *Conn) Close() (err error) {
	if conn != nil && conn.nconn != nil {
		err = conn.nconn.Close()
		conn.closeChan <- conn.addr
		conn.nconn = nil
	}
	return
}

// loop loops, reading messages from the connection as long as it is open, and passing them on to msgChan.
func (conn *Conn) loop() {
	if conn != nil && conn.nconn != nil {
		defer conn.nconn.Close()
		for conn.nconn != nil {
			if n, err := conn.nconn.Read(conn.buff); err != nil {
				conn.Close()
				break
			} else if n > 4 {
				msgId := binary.BigEndian.Uint32(conn.buff)
				var data proto.Message = conn.gnmc(msgId)
				// If data is nil, the gnmc function must not have recognized the message ID. We'll ignore the message.
				if data != nil {
					proto.Unmarshal(conn.buff[1:n], data)
					msg := NewMessage(conn.nconn.RemoteAddr(), msgId, data)
					conn.msgChan <- msg
				}
			}
		}
	}
}

// Address returns the remote address of the connection, or nil if conn is nil or it isn't connected to anything.
func (conn *Conn) Address() (addr net.Addr) {
	if conn != nil && conn.nconn != nil {
		addr = conn.addr
	}
	return
}

// Send sends a message through the connection.
func (conn *Conn) Send(msgId uint32, msg proto.Message) (err error) {
	if conn != nil && conn.nconn != nil {
		var buff []byte
		buff, err = proto.Marshal(msg)
		if len(buff)+1 >= BufferSize {
			err = fmt.Errorf("Message with ID %v is too long!", msgId)
		} else {
			var msgIdBytes []byte = make([]byte, 4)
			binary.BigEndian.PutUint32(msgIdBytes, msgId)
			buff = append(msgIdBytes, buff...)
			conn.nconn.Write(buff)
		}
	}
	return
}
