package node

import (
	//dns "../dnsService"
	nameService "../localNameService"
	MP "../messagePasser/"
	streamElection "../streamElection"
	"fmt"
	nc "./nodeContext"
)

const (
	bootstrap_dns = "DS.supernodes.com"
	HeartBeatPort = 8888
)


var mp *MP.MessagePasser
var sElection *streamElection.StreamElection
var nodeContext *nc.NodeContext

/**
All internal helper functions
*/
func heartBeat() {

}


/* Event handler distributer*/
func listenOnChannel(channelName string, handler func(*MP.Message, *nc.NodeContext)) {
	for {
		//
		msg := <- mp.Messages[channelName]
		go handler(msg, nodeContext)
	}
}

/**
Here goes all the internal event handlers
*/

func joinAssign(msg *MP.Message, nodeContext *nc.NodeContext) {
	// Store the parentIP
	nodeContext.ParentIP = msg.Src
	// Test
	fmt.Println("Be assigned to parent! " + nodeContext.ParentIP)
}

func streamAssign(msg *MP.Message, nodeContext *nc.NodeContext) {

}

func programListParser(msg *MP.Message, nodeContext *nc.NodeContext) {

}

func receiveReceive(msg *MP.Message, nodeContext *nc.NodeContext) {

}

/**
Here goes all the apis to be called by the application
*/

func Start() {
	nodeContext = new(nc.NodeContext)
	nodeContext.SetLocalName(nameService.GetLocalName())
	mp = MP.NewMessagePasser(nodeContext.LocalName)
	go heartBeat()

	// Initialize all the package structs
	sElection = streamElection.NewStreamElection(mp)

	// Define all the channel names and the binded functions
	// TODO: Register your channel name and binded eventhandlers here
	// The map goes as  map[channelName][eventHandler]
	// All the messages with type channelName will be put in this channel by messagePasser
	// Then the binded handler of this channel will be called with the argument (*Message)
	channelNames := map[string]func(*MP.Message, *nc.NodeContext){
		"join_assign":     joinAssign,
		"stream_assign":   streamAssign,
		"program_list":    programListParser,
		"election_stream": receiveReceive,
	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *MP.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}

}

/* Join the network */
func NodeJoin(IP string) {
	helloMsg := MP.NewMessage(IP, "join", "hello, my name is Bay Max, you personal healthcare companion")
	mp.Send(helloMsg)
	echoMsg := <- mp.Messages["ack"]
	fmt.Println(echoMsg)
}

/* Start Streaming */
func StreamStart() {

}

/* Stop Streaming */
func StreamStop() {

}

/* Join a streaming group */
func StreamJoin(programId string) {

}

/* Stream Quit */
func StreamQuit() {

}

/* Get the list of programs */
// TODO: Add a return type
func StreamGetList() {

}

/* Produce the stream the data */
func StreamData(data string) {

}

/* Get the data from other streamers */
func StreamReadData() string {
	return ""
}
