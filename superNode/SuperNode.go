package node

import (
	"fmt"

	dht "../dht"
	dns "../dnsService"
	MP "../messagePasser"
	config "../config"

	joinElection "../supernodeLib/joinElection"
	streaming "../supernodeLib/streaming"
)

const (
	localname = "DS.supernodes.com"
)

var mp *MP.MessagePasser
var dHashtable *dht.DHT
var streamHandler *streaming.StreamingHandler
var jElection *joinElection.JoinElection

//var sElection	*streamElection.StreamElection

func Start() {
	// First register on the dnsService
	// In test stage, it's actually "ec2-54-86-213-108.compute-1.amazonaws.com"
	dns.RegisterSuperNode(config.BootstrapDomainName)
	fmt.Println("Message Passer To initialize!")
	// Initialize the message passer
	// Note: all the packages are using the same message passer!
	mp = MP.NewMessagePasser(localname)
	fmt.Println("Message Passer Initialized!")

	// Initialize all the package structs
	dHashtable = dht.NewDHT(mp)
	streamHandler = streaming.NewStreamingHandler(dHashtable, mp)
	jElection = joinElection.NewJoinElection(mp)
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

		"join":          jElection.Start,
		"join_election": jElection.Receive,

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

		//"stream_election":	sElection.Receive,
	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *MP.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}
}

func listenOnChannel(channelName string, handler func(*MP.Message)) {
	for {
		//
		msg := <-mp.Messages[channelName]
		go handler(msg)
	}
}
