package node

import (
	"fmt"

	dns 		"../dnsService"
	dht 		"../dht"
	messagePasser 	"../messagePasser"

	joinElection 	"../supernodeLib/joinElection"
	streamElection  "../streamElection"
	streaming	"../supernodeLib/streaming"

)


const (
	localname = "DS.supernodes.com"
)


var mp 	*messagePasser.MessagePasser
var dHashtable 	*dht.DHT
var streamHandler 	*streaming.StreamingHandler
var jElection 	*joinElection.JoinElection
var sElection	*streamElection.StreamElection


func Start(){
	// First register on the dnsService
	// In test stage, it's actually "ec2-54-86-213-108.compute-1.amazonaws.com"
	dns.RegisterSuperNode(localname)
	fmt.Println("Message Passer To initialize!")
	// Initialize the message passer
	// Note: all the packages are using the same message passer!
	mp = messagePasser.NewMessagePasser(localname)
	fmt.Println("Message Passer Initialized!")

	// Initialize all the package structs
	dHashtable = dht.NewDHT(mp)
	streamHandler = streaming.NewStreamingHandler(dHashtable, mp)
	jElection = joinElection.NewJoinElection(mp)
	sElection = streamElection.NewStreamElection(mp)


	// Define all the channel names and the binded functions
	// TODO: Register your channel name and binded eventhandlers here
	// The map goes as  map[channelName][eventHandler]
	// All the messages with type channelName will be put in this channel by messagePasser
	// Then the binded handler of this channel will be called with the argument (*Message)
	channelNames := map[string]func(*messagePasser.Message){
		// "dht": dHashtable.msgHandler(messaage),

		"stream_start":	streamHandler.StreamStart,
		"stream_get_list"	: streamHandler.StreamGetList,
		"stream_join":	streamHandler.StreamJoint,

		"join":          jElection.Start,
		"join_election": jElection.Receive,
		
		"stream_election":	sElection.Receive,
	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *messagePasser.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}
}

func listenOnChannel(channelName string, handler func(*messagePasser.Message)) {
	for {
		//
		msg := <-mp.Messages[channelName]
		go handler(msg, mp)
	}
}