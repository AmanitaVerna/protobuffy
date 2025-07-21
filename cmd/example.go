package cmd

import (
	"fmt"

	"github.com/amanitaverna/protobuffy"
	"github.com/amanitaverna/protobuffy/cmd/pb"
	"google.golang.org/protobuf/proto"
)

const (
	ChatMessageID uint32 = iota
	ShutdownMessageID
)

func main() {
	ServerStuff()
}

// ServerStuff runs a simple example server.
func ServerStuff() (err error) {
	server, err := protobuffy.NewServer(protobuffy.TCP, 12345, GetNewMessage)
	go server.Start()
	connChan := server.GetNewConnectionChannel()
	closedChan := server.GetConnectionClosedChannel()
	recvChan := server.GetReceiveChannel()
	sendChan := server.GetSendChannel()
	for server != nil {
		select {
		case addr := <-connChan:
			fmt.Printf("New connection from address: %v\n", addr)
			sendChan <- protobuffy.NewMessage(addr, ChatMessageID, &pb.ChatMessage{Text: "Hello there!"})
		case addr := <-closedChan:
			fmt.Printf("Disonnection by address: %v\n", addr)
		case msg := <-recvChan:
			fmt.Printf("Received message from address %v, with message ID %v. Message contents: %v\n", msg.Address(), msg.MsgId(), msg.Data())
			switch msg.MsgId() {
			case ChatMessageID:
				if chat, ok := msg.Data().(*pb.ChatMessage); ok {
					fmt.Printf("Client at address %v sent us a message: %v\n", msg.Address(), chat.Text)
				}
			case ShutdownMessageID:
				if shutdownMsg, ok := msg.Data().(*pb.ShutdownMessage); ok {
					fmt.Printf("Server shutdown requested by client at address %v. Reason: %v\n", msg.Address(), shutdownMsg.Text)
					server.Shutdown()
					server = nil
				}
			}
		}
	}
	return
}

func GetNewMessage(messageType uint32) (msg proto.Message) {
	switch messageType {
	case ChatMessageID:
		msg = &pb.ChatMessage{}
	case ShutdownMessageID:
		msg = &pb.ShutdownMessage{}
	}
	return
}
