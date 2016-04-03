package node

import (
	"fmt"

	Dht "../dht"
	dns "../dnsService"
	MP "../messagePasser"
	SNC "./superNodeContext/"
	JoinElection "../supernodeLib/joinElection"
	Streaming "../supernodeLib/streaming"
	"time"
)

const (
	localname = "DS.supernodes.com"
)

var mp *MP.MessagePasser
var dHashtable *Dht.DHT
var streamHandler *Streaming.StreamingHandler
var jElection *JoinElection.JoinElection
var superNodeContext *SNC.SuperNodeContext

//var sElection	*streamElection.StreamElection

func Start() {
	// First register on the dnsService
	// In test stage, it's actually "ec2-54-86-213-108.compute-1.amazonaws.com"
	dns.RegisterSuperNode(localname)
	fmt.Println("Message Passer To initialize!")
	// Initialize the message passer
	// Note: all the packages are using the same message passer!
	mp = MP.NewMessagePasser(localname)
	fmt.Println("Message Passer Initialized!")

	// Block supernode until receive exit msg
	mp.AddMappings([]string{"exit"})
	// Initialize SuperNodeContext
	// Currently SuperNodeContext contains all info of the assigned child nodes
	superNodeContext = SNC.NewSuperNodeContext()

	// Initialize all the package structs
	dHashtable = Dht.NewDHT(mp)
	streamHandler = Streaming.NewStreamingHandler(dHashtable, mp)
	jElection = JoinElection.NewJoinElection(mp)
	//sElection = streamElection.NewStreamElection(mp)

	// Define all the channel names and the binded functions
	// TODO: Register your channel name and binded eventhandlers here
	// The map goes as  map[channelName][eventHandler]
	// All the messages with type channelName will be put in this channel by messagePasser
	// Then the binded handler of this channel will be called with the argument (*Message)
	channelNames := map[string]func(*MP.Message){
		// "dht": dHashtable.msgHandler(messaage),

		"stream_start":    streamHandler.StreamStart,
		"stream_get_list": streamHandler.StreamGetList,
		"stream_join":     streamHandler.StreamJoin,
		"heartbeat": heartBeatHandler,
		"hello":          jElection.Start,
		"join": 			newChild,
		"join_election": jElection.Receive,
		"error": errorHandler,
		//"stream_election":	sElection.Receive,
	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *MP.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}

	exitMsg := <- mp.Messages["exit"]
	fmt.Println(exitMsg)
}

func listenOnChannel(channelName string, handler func(*MP.Message)) {
	for {
		//
		msg := <-mp.Messages[channelName]
		go handler(msg)
	}
}

func errorHandler(*MP.Message)  {

}

func heartBeatHandler(*MP.Message)  {
	time.Sleep(10 * time.Second)
	hasDead, deadNodes := superNodeContext.CheckDead()
	if hasDead {
		superNodeContext.RemoveNodes(deadNodes)
	}
	superNodeContext.ResetState()
}

func newChild(msg *MP.Message)  {
	superNodeContext.AddNode(msg.SrcName)
}