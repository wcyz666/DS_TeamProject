package node

import (
	//dns "../dnsService"
	nameService "../localNameService"
	MP "../messagePasser/"
	StreamElection "../streamElection"
	"fmt"
	nc "./nodeContext"
	"time"
)

const (
	bootstrap_dns = "DS.supernodes.com"
	HeartBeatPort = 8888
)


var mp *MP.MessagePasser
var sElection *StreamElection.StreamElection
var nodeContext *nc.NodeContext
var exitChannal chan int

/**
All internal helper functions
*/
func heartBeat() {
	for {
		time.Sleep(time.Second * 5)
		fmt.Println("Node: send out heart beat message")
		mp.Send(MP.NewMessage(nodeContext.ParentIP, nodeContext.ParentName, "heartbeat", "Hello, this is a heartbeat message."))
	}
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
	nodeContext.ParentName = msg.SrcName
	// Test
	fmt.Printf("Be assigned to parent! IP [%s], Name [%s]\n" + nodeContext.ParentIP, nodeContext.ParentName)
}

func streamAssign(msg *MP.Message, nodeContext *nc.NodeContext) {

}

func programListParser(msg *MP.Message, nodeContext *nc.NodeContext) {

}

func receiveReceive(msg *MP.Message, nodeContext *nc.NodeContext) {

}

func errorHandler(msg *MP.Message, nodeContext *nc.NodeContext) {
	switch nodeContext.State {
	// Re-throw it to init_fail channel
	case nc.NodeHello:
		msg.Kind = "init_fail"
	}

	mp.Messages[msg.Kind] <- msg
}


/**
Here goes all the apis to be called by the application
*/

func Start(IPs []string) {
	nodeContext = nc.NewNodeContext()
	nodeContext.SetLocalName(nameService.GetLocalName())
	mp = MP.NewMessagePasser(nodeContext.LocalName)

	// We use for loop to connect with all supernode one-by-one,
	// if a connection to one supernode fails, an error message
	// will be sent by messagePasser, and this message is further
	// processed in error handler.
	// init_fail: used in hello phase
	// exit: used when all supernode cannot be connected.
	mp.AddMappings([]string{"exit", "init_fail", "ack"})

	// Initialize all the package structs
	sElection = StreamElection.NewStreamElection(mp)

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
		"error" : errorHandler,
	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *MP.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}
	go nodeJoin(IPs)
	exitMsg := <- mp.Messages["exit"]
	fmt.Println(exitMsg)
}

/* Join the network */
func nodeJoin(IPs []string) {
	//Send hello messages until find out a working supernode
	i := 0
	helloMsg := MP.NewMessage(IPs[i], "", "join", "hello, my name is Bay Max, you personal healthcare companion")
	mp.Send(helloMsg)
	for {
		select {
		case err := <-mp.Messages["init_fail"]:
			// wait and retry the next
			fmt.Printf("Connetion to spernode failed: %s\n", err.Data)
			i += 1
			if (i == len(IPs)) {
				exitMsg := MP.NewMessage("self", nodeContext.LocalName, "exit", "All supernodes are down, exit")
				mp.Messages["exit"] <- &exitMsg
				break;
			}
			helloMsg := MP.NewMessage(IPs[i], "", "join", "hello, my name is Bay Max, you personal healthcare companion")
			mp.Send(helloMsg)
		case msg := <- mp.Messages["ack"]:
			fmt.Printf("Node: receiving ACK message [%s]\n", msg)
			nodeContext.ParentIP = msg.Src
			nodeContext.ParentName = msg.SrcName
			go heartBeat()
		}
	}

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
