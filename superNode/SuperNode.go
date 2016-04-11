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

	"strconv"
	"bufio"
	"os"
	"strings"
)

const (
	localname = "DS.supernodes.com"
)

var mp *MP.MessagePasser
var dhtService *Dht.DHTService
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

	dhtService = Dht.NewDHTService(mp)
	streamHandler = Streaming.NewStreamingHandler(dhtService, mp, superNodeContext)
	jElection = JoinElection.NewJoinElection(mp)

	dhtNode := dhtService.DhtNode
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
		"join_dht_req":            		dhtNode.HandleJoinReq,
		"join_dht_complete":        	dhtNode.HandleJoinComplete,  // To indicate successor about completion of join
		"join_dht_notify":          	dhtNode.HandleJoinNotify,    // To indicate predecessor about completion of join
		"leave_dht_req":            	dhtNode.Leave,
		"dht_broadcast_msg":        	dhtNode.HandleBroadcastMessage,
		"dht_neighbourhood_discovery":	dhtNode.HandleNeighbourhoodDiscovery,
		"dht_ring_repair_req":			dhtNode.HandleRingRepairRequest,
		"dht_ring_repair_res":			dhtNode.HandleRingRepairResponse,

		/* DHT Data operation handlers */
		/* Having separate channels will allow concurrent access to hash map.
		 * Need to update hash table to be a concurrent map */
		"create_entry_req":		dhtNode.HandleDataOperationRequest,
		"update_entry_req":		dhtNode.HandleDataOperationRequest,
		"delete_entry_req":		dhtNode.HandleDataOperationRequest,
		"get_data_req":			dhtNode.HandleDataOperationRequest,

		/* Here goes the handlers related to streaming process */
		"stream_start": streamHandler.StreamStart,
		"stream_stop": streamHandler.StreamStop,
		"stream_join":     streamHandler.StreamJoin,
		"stream_program_start": streamHandler.StreamProgramStart,  // This is sent from other supernodes
		"stream_program_stop": streamHandler.StreamProgramStop,  // This is sent from other supernodes


	}

	// Init and listen
	for channelName, _ := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *MP.Message)

	}

	for channelName, handler := range channelNames {
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}
	//stgo nodeStateWatcher()

	status := dhtService.Start()
	if (Dht.DHT_API_SUCCESS != status){
		panic ("DHT service start failed. Error is " + strconv.Itoa(status))
	}

	/* Start a CLI to handle user interaction */
	go DhtCLIInterface(dhtService)
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
		//fmt.Printf("SuperNode: check node state, Alive child count: [%d]\n", superNodeContext.GetNodeCount())
		superNodeContext.ResetState()
	}
}

func newChild(msg *MP.Message)  {
	fmt.Printf("SuperNode: receive new Node, IP [%s] Name [%s]\n", msg.Src, msg.SrcName)
	mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "ack", MP.EncodeData("this is an ACK message")))
	superNodeContext.AddNode(msg.SrcName, msg.Src)
}

func printHelp(){
	fmt.Println("Enter C Key MemberShipInfo to create an Streaming Group")
	fmt.Println("      D Key to delete a Streaming Group")
	fmt.Println("      A Key MemberShipInfo to add a member")
	fmt.Println("      R Key MemberShipInfo to delete a member")
	fmt.Println("      G Key to retrieve contents of a streaming group")
	fmt.Println("      H for help")
	fmt.Println("      Q to quit")
	fmt.Println(" For membership info, please pass the IP address (of parent super node)")
}

func DhtCLIInterface(dhtService *Dht.DHTService){
	printHelp()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if ((line == "Q") || (line == "q")) {
			os.Exit(0)
		} else if ((line == "H") || (line == "h")){
			printHelp()
		} else {
			inputList := strings.Split(line," ")
			switch inputList[0] {
			case "C", "c":
				if (len(inputList) !=3){
					fmt.Println("Invalid format")
					printHelp()
				} else {
					status := dhtService.Create(inputList[1], Dht.MemberShipInfo{inputList[2]})
					fmt.Println("Create API called and return status is "+strconv.Itoa(status))
				}
			case "D","d":
				if (len(inputList) !=2){
					fmt.Println("Invalid format")
					printHelp()
				} else {
					status := dhtService.Delete(inputList[1])
					fmt.Println("Delete API called and return status is "+ strconv.Itoa(status))
				}
			case "A","a":
				if (len(inputList) !=3){
					fmt.Println("Invalid format")
					printHelp()
				} else {
					status := dhtService.Append(inputList[1], Dht.MemberShipInfo{inputList[2]})
					fmt.Println("Append API called and return status is "+strconv.Itoa(status))
				}
			case "R","r":
				if (len(inputList) !=3){
					fmt.Println("Invalid format")
					printHelp()
				} else {
					status := dhtService.Remove(inputList[1], Dht.MemberShipInfo{inputList[2]})
					fmt.Println("Remove API called and return status is "+strconv.Itoa(status))
				}
			case "G","g":
				if (len(inputList) !=2){
					fmt.Println("Invalid format")
					printHelp()
				} else {
					memberShipInfo, status := dhtService.Get(inputList[1])
					fmt.Println("Get API called and return status is "+strconv.Itoa(status))
					fmt.Println("Members of streaming group are ")
					for _,member := range memberShipInfo{
						fmt.Println("	"+member.SuperNodeIp)
					}
				}
			default:
				fmt.Println("Unexpected option")
				printHelp()
			}
		}
	}
	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}