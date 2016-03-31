package node

import (
	"fmt"

	messagePasser "../messagePasser"
	dns "../dnsService"
)

const (
	localname = "DS.supernodes.com"
)

var mp *messagePasser.MessagePasser

func Start(){
	// First register on the dnsService
	// In test stage, it's actually "ec2-54-175-192-219.compute-1.amazonaws.com"
	dns.RegisterSuperNode(localname)
	mp = messagePasser.NewMessagePasser(localname);
}

/**
A sample handler: To react to all join messages (a node requests to join the network)
 */
func Join(){
	var msg messagePasser.Message
	msg <- mp.Messages["hello"]
	fmt.Println(msg)
	fmt.Println("New message received! " + msg.GetSrc() + " Joined!")

	// TODO: Remove
	// Test Code
	replyMsg := messagePasser.NewMessage(msg.GetSrc(), "reply", msg.GetData())
	mp.Send(replyMsg)
}

