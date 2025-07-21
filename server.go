package protobuffy

import (
	"fmt"
	"net"

	"google.golang.org/protobuf/proto"
)

type ConnectionType string

const (
	TCP ConnectionType = "tcp"
	UDP ConnectionType = "udp"
)

// Server listens for connections and manages them.
type Server struct {
	listener    net.Listener
	conns       map[net.Addr]*Conn
	connChan    chan net.Addr
	closeChan   chan net.Addr
	outMsgChan  chan *Message
	inMsgChan   chan *Message
	inBytesChan chan *ByteMessage
	gnmc        func(messageType uint32) proto.Message
}

// NewServer creates a new Server.
// getNewMessage should be a function which can be called to get a proto.Message which proto.Unmarshal can write to.
// We call it when we receive a packet from a client.
// To start accepting connections, call `go server.Start()`, then call
// server.GetNewConnectionChannel(), server.GetReceiveChannel(), and server.GetSendChannel() to get the channels. See their documentation for what they are for.
// If net.Listen returns an error, it will be returned and no Server will be created.
// Messages sent to and from the server should have a 4 byte prefix which holds the message ID in big-endian orientation.
// This is needed to know what protobuf message is contained in the packet. The server will include this automatically when sending messages, and will use it
// when receiving them, but clients will have to be sure to include it in packets they are sending and look for it in packets they are receiving.
func NewServer(connectionType ConnectionType, port int, getNewMessage func(messageType uint32) proto.Message) (server *Server, err error) {
	var listener net.Listener
	if listener, err = net.Listen(string(connectionType), fmt.Sprintf(":%v", port)); err == nil {
		server = &Server{listener: listener, conns: make(map[net.Addr]*Conn), gnmc: getNewMessage}
	}
	return
}

// Start creates the channels and starts listening for connections. Call it with `go server.Start()` if you don't want it to block.
// After calling it, call server.GetNewConnectionChannel(), server.GetConnectionClosedChannel(), server.GetReceiveChannel(), and server.GetSendChannel().
func (server *Server) Start() {
	server.connChan = make(chan net.Addr, 100)
	server.closeChan = make(chan net.Addr, 100)
	server.outMsgChan = make(chan *Message, 100)
	server.inMsgChan = make(chan *Message, 100)
	go server.listen()
	for {
		msg := <-server.outMsgChan
		if msg != nil {
			addr := msg.Address()
			if addr == nil {
				// Send the message to all connections.
				for _, conn := range server.conns {
					conn.Send(msg.msgId, msg.Data())
				}
			} else if conn := server.conns[addr]; conn != nil {
				conn.Send(msg.msgId, msg.Data())
			}
		}
	}
}

// GetNewConnectionChannel returns the channel where new connections are announced. You can receive from it to see the address of newly accepted connections.
func (server *Server) GetNewConnectionChannel() (ch <-chan net.Addr) {
	if server != nil {
		ch = server.connChan
	}
	return
}

// GetConnectionClosedChannel returns the channel where closed connections are announced. You can receive from it to see the address of newly closed connections.
func (server *Server) GetConnectionClosedChannel() (ch <-chan net.Addr) {
	if server != nil {
		ch = server.closeChan
	}
	return
}

// GetReceiveChannel returns the channel where messages received from connections are sent. You can receive from it to read incoming messages.
func (server *Server) GetReceiveChannel() (inMsgChan <-chan *Message) {
	if server != nil {
		inMsgChan = server.inMsgChan
	}
	return
}

// GetSendChannel returns the channel where you can send messages you want to send to a connection, or to all connections.
// To send to all connections, specify a nil address in your Message.
func (server *Server) GetSendChannel() (outMsgChan chan<- *Message) {
	if server != nil {
		outMsgChan = server.outMsgChan
	}
	return
}

// CloseConnection tells the server to close the connection to a particular address.
func (server *Server) CloseConnection(addr net.Addr) {
	if server != nil {
		if conn := server.conns[addr]; conn != nil {
			conn.Close()
			delete(server.conns, addr)
		}
	}
}

// listen listens for new connections. It exits when server.listener is nil, which happens when the server shuts down.
func (server *Server) listen() {
	for server != nil && server.listener != nil {
		var nconn net.Conn
		nconn, _ = server.listener.Accept()
		if nconn != nil {
			var conn *Conn = newConn(server.inMsgChan, server.closeChan, nconn, server.gnmc)
			server.conns[conn.Address()] = conn
			server.connChan <- conn.Address()
			go conn.loop()
		}
	}
}

// Shutdown shuts down the listener, closes all connections, closes all channels, and sets everything to nil.
func (server *Server) Shutdown() {
	// We've been asked to shut down the server.
	server.listener.Close()
	for _, c := range server.conns {
		c.Close()
	}
	server.conns = nil
	close(server.connChan)
	close(server.closeChan)
	close(server.inMsgChan)
	close(server.outMsgChan)
	server.connChan = nil
	server.closeChan = nil
	server.inMsgChan = nil
	server.outMsgChan = nil
	server.listener = nil
}
