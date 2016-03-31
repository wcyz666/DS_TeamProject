package node

import (
	"fmt"

	dns "../dnsService"
	joinElection "../joinElection"
	messagePasser "../messagePasser"
)

const (
	localname = "DS.supernodes.com"
)

var mp *messagePasser.MessagePasser

func Start() {
	// First register on the dnsService
	// In test stage, it's actually "ec2-54-175-192-219.compute-1.amazonaws.com"
	dns.RegisterSuperNode(localname)
	fmt.Println("Message Passer To initialize!")
	// Initialize the message passer
	// Note: all the packages are using the same message passer!
	mp = messagePasser.NewMessagePasser(localname)
	fmt.Println("Message Passer Initialized!")

	// Define all the channel names and the binded functions
	channelNames := map[string]func(*messagePasser.Message, *messagePasser.MessagePasser){
		"join":          joinElection.Start,
		"election_join": joinElection.Receive,
		// "dht": dhtHandler
	}
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *messagePasser.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}
}

func listenOnChannel(channelName string, handler func(*messagePasser.Message, *messagePasser.MessagePasser)) {
	for {
		//
		msg := <-mp.Messages[channelName]
		go handler(msg, mp)
	}
}
