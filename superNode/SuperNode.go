package node

import (
	"fmt"

	messagePasser "../messagePasser"
	dns "../dnsService"
	joinElection "../joinElection"
)

const (
	localname = "DS.supernodes.com"
)

var mp *messagePasser.MessagePasser

func Start(){
	// First register on the dnsService
	// In test stage, it's actually "ec2-54-175-192-219.compute-1.amazonaws.com"
	dns.RegisterSuperNode(localname)
	fmt.Println("Message Passer To initialize!")
	mp = messagePasser.NewMessagePasser(localname);
	fmt.Println("Message Passer Initialized!")
	go Join()
	_ <- nil
}


/**
A sample handler: To react to all join messages (a node requests to join the network)
 */
func Join(){
	for {
		//fmt.Println(mp.Messages["join"])
		channel, ok := mp.Messages["join"]
		if (ok == false) {
			mp.Messages["join"] = make(chan *messagePasser.Message)
			channel = mp.Messages["join"]
		}
		msg := <-channel
		//msg := <- mp.Incoming;
		fmt.Println(msg)
		fmt.Println("New message received! " + msg.Src + " Joined!")

		go joinElection.Start(mp, msg)

		// TODO: Remove
		// Test Code
		replyMsg := messagePasser.NewMessage(msg.Src, "ack", msg.Data)
		mp.Send(replyMsg)
	}
}
