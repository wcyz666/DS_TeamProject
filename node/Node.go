package node

import (
	//dns "../dnsService"
	nameService "../localNameService"
	MP "../messagePasser/"
	streamElection "../streamElection"
	"fmt"
)

const (
	bootstrap_dns = "DS.supernodes.com"
	HeartBeatPort = 8888
)


type NodeContext struct {
	mp *MP.MessagePasser
	localName string
	parentIP string
	sElection *streamElection.StreamElection
}

var nodeContext *NodeContext

/**
All internal helper functions
*/
func heartBeat() {

}

func (nodeContext *NodeContext) setLocalName(name string) {
	nodeContext.localName = name
}

/* Event handler distributer*/
func listenOnChannel(channelName string, handler func(*MP.Message, nodeContext *NodeContext)) {
	for {
		//
		msg := <- nodeContext.mp.Messages[channelName]
		go handler(msg)
	}
}

/**
Here goes all the internal event handlers
*/

func joinAssign(msg *MP.Message, nodeContext *NodeContext) {
	// Store the parentIP
	nodeContext.parentIP = msg.Src
	// Test
	fmt.Println("Be assigned to parent! " + nodeContext.parentIP)
}

func streamAssign(msg *MP.Message, nodeContext *NodeContext) {

}

func programListParser(msg *MP.Message, nodeContext *NodeContext) {

}

/**
Here goes all the apis to be called by the application
*/

func Start() {
	nodeContext = new(NodeContext)
	nodeContext.setLocalName(nameService.GetLocalName())
	nodeContext.mp = MP.NewMessagePasser(nodeContext.localName)
	go heartBeat()

	// Initialize all the package structs
	nodeContext.sElection = streamElection.NewStreamElection(nodeContext.mp)

	// Define all the channel names and the binded functions
	// TODO: Register your channel name and binded eventhandlers here
	// The map goes as  map[channelName][eventHandler]
	// All the messages with type channelName will be put in this channel by messagePasser
	// Then the binded handler of this channel will be called with the argument (*Message)
	channelNames := map[string]func(*MP.Message, *NodeContext){
		"join_assign":     joinAssign,
		"stream_assign":   streamAssign,
		"program_list":    programListParser,
		"election_stream": nodeContext.sElection.Receive,
	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		nodeContext.mp.Messages[channelName] = make(chan *MP.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}

}

/* Join the network */
func NodeJoin(IP string) {
	helloMsg := MP.NewMessage(IP, "join", "hello, my name is Bay Max, you personal healthcare companion")
	nodeContext.mp.Send(helloMsg)
	echoMsg := <- nodeContext.mp.Messages["ack"]
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
