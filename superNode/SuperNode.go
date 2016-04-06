package node

import (
	"fmt"

	Dht "../dht"
	dns "../dnsService"
	MP "../messagePasser"

	Config "../config"
	SNC "./superNodeContext/"
	JoinElection "../supernodeLib/joinElection"
	Streaming "../streaming/supernodeStreamingHandler"
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
	// Initialize SuperNodeContext
	// Currently SuperNodeContext contains all info of the assigned child nodes
	superNodeContext = SNC.NewSuperNodeContext()
	// First register on the dnsService
	// In test stage, it's actually "ec2-54-86-213-108.compute-1.amazonaws.com"
	dns.RegisterSuperNode(Config.BootstrapDomainName)
	fmt.Println("Message Passer To initialize!")
	// Initialize the message passer
	// Note: all the packages are using the same message passer!
	mp = MP.NewMessagePasser(superNodeContext.LocalName)
	fmt.Println("Message Passer Initialized!")

	// Block supernode until receive exit msg
	mp.AddMappings([]string{"exit"})

	// Initialize all the package structs
	//dHashtable = Dht.NewDHT(mp)
	streamHandler = Streaming.NewStreamingHandler(dHashtable, mp, superNodeContext)
	jElection = JoinElection.NewJoinElection(mp)
	//sElection = streamElection.NewStreamElection(mp)

	// Define all the channel names and the binded functions
	// TODO: Register your channel name and binded eventhandlers here
	// The map goes as  map[channelName][eventHandler]
	// All the messages with type channelName will be put in this channel by messagePasser
	// Then the binded handler of this channel will be called with the argument (*Message)
	channelNames := map[string]func(*MP.Message){
		// "dht": dHashtable.msgHandler(messaage),

		"heartbeat": heartBeatHandler,
		"hello":          jElection.Start,
		"join": 			newChild,
		"join_election": jElection.Receive,
		"error": errorHandler,

		/* DHT call backs */
		"join_dht_req":            	dHashtable.HandleJoinReq,
		"join_dht_res":             dHashtable.HandleJoinRes,
		"join_dht_complete":        dHashtable.HandleJoinComplete,  // To indicate successor about completion of join
		"join_dht_notify":          dHashtable.HandleJoinNotify,    // To indicate predecessor about completion of join
		"leave_dht_req":            dHashtable.Leave,

		/* Having separate channels will allow concurrent access to hash map.
		 * Need to update hash table to be a concurrent map */
		// Creates new (key,value) pair in DHT. Used when creating a new streaming group
		"create_entry_req":         dHashtable.HandleCreateNewEntryReq,
		"create_entry_res":         dHashtable.HandleCreateNewEntryRes,
		// Update existing entry in DHT. Used when adding or removing a member to existing streaming group
		"update_entry_req":         dHashtable.HandleUpdateEntryReq,
		"update_entry_res":         dHashtable.HandleUpdateEntryRes,
		// Delete existing entry in DHT. Used when dissolving a streaming group
		"delete_entry_req":         dHashtable.HandleDeleteEntryReq,
		"delete_entry_res":         dHashtable.HandleDeleteEntryRes,
		// Query contents of existing entry using its key . Used to learn about members of an existing group
		"get_data_req":            dHashtable.HandleGetDataReq,
		"get_data_res":            dHashtable.HandleGetDataRes,

		/* Here goes the handlers related to streaming process */
		"stream_start": streamHandler.StreamStart,
		"stream_stop": streamHandler.StreamStop,
		"stream_join":     streamHandler.StreamJoin,
		"stream_new_program": streamHandler.StreamProgramStart,  // This is sent from other supernodes
		"stream_new_program": streamHandler.StreamProgramStop,  // This is sent from other supernodes


	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *MP.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}
	go nodeStateWatcher()

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

func heartBeatHandler(msg *MP.Message)  {
	superNodeContext.SetAlive(msg.SrcName)
}

func nodeStateWatcher() {
	for {
		time.Sleep(5 * time.Second)
		hasDead, deadNodes := superNodeContext.CheckDead()
		if hasDead {
			for _, nodeName := range deadNodes {
				mp.RemoveMapping(superNodeContext.GetIPByName(nodeName))
				mp.RemoveMapping(nodeName)
				superNodeContext.RemoveNodes(nodeName)
			}
		}
		fmt.Printf("SuperNode: check node state, Alive child count: [%d]\n", superNodeContext.GetNodeCount())
		superNodeContext.ResetState()
	}
}

func newChild(msg *MP.Message)  {
	fmt.Printf("SuperNode: receive new Node, IP [%s] Name [%s]\n", msg.Src, msg.SrcName)
	mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "ack", MP.EncodeData("this is an ACK message")))
	superNodeContext.AddNode(msg.SrcName, msg.Src)
}
