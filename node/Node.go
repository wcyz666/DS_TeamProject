package node

import (
	DNS "../dnsService"
	nameService "../localNameService"
	MP "../messagePasser/"
	"fmt"
	JE "../supernodeLib/joinElection"
	NC "./nodeContext"
	"time"
	Streamer "../streaming/streamer"
	"bufio"
	"os"
	"strings"
)

const (
	bootstrap_dns = "p2plive.supernodes.com"
)


var mp *MP.MessagePasser
var nodeContext *NC.NodeContext
var streamer *Streamer.Streamer
var isSendHeartBeat bool

/**
All internal helper functions
*/
func heartBeat() {
	for {
		time.Sleep(time.Second * 5)
		if isSendHeartBeat {
			mp.Send(MP.NewMessage(nodeContext.ParentIP, nodeContext.ParentName, "node_heartbeat", MP.EncodeData("Hello, this is a heartbeat message.")))
		}
		//fmt.Println("Node: send out heart beat message")
	}
}


/* Event handler distributer*/
func listenOnChannel(channelName string, handler func(*MP.Message)) {
	for {
		//
		msg := <- mp.Messages[channelName]
		go handler(msg)
	}
}

/**
Here goes all the internal event handlers
*/


func joinAssign(msg *MP.Message) {

	// Store the parentIP
	result := JE.ElectionResult{}
	MP.DecodeData(&result, msg.Data)
	nodeContext.ParentIP = result.ParentIP
	nodeContext.ParentName = result.ParentName
	fmt.Println(result)
	isSendHeartBeat = true
	go heartBeat()
	nodeContext.State = NC.Joined
	fmt.Printf("Be assigned to parent! IP [%s], Name [%s]\n", result.ParentIP, result.ParentName)
	joinMsg := MP.NewMessage(nodeContext.ParentIP, nodeContext.ParentName, "election_join", MP.EncodeData("hello, my name is Bay Max, you personal healthcare companion"))
	mp.Send(joinMsg)
}


func errorHandler(msg *MP.Message) {
	switch nodeContext.State {
	// Re-throw it to init_fail channel
	case NC.NodeHello:
		msg.Kind = "init_fail"
		mp.Messages[msg.Kind] <- msg
	case NC.Joined:
		failNode := MP.FailClientInfo{}
		MP.DecodeData(&failNode, msg.Data)
		if (failNode.IP == nodeContext.ParentIP) {
			fmt.Println("Node: detect Supernode failure, try another supernode...")
			msg.Kind = "super_fail"
			mp.Messages[msg.Kind] <- msg
		}else{
			streamer.HandleErrorMsg(msg)
		}

	}

}





/**
Here goes all the apis to be called by the application
*/

func Start() {
	IPs := DNS.GetAddr(bootstrap_dns)
	nodeContext = NC.NewNodeContext()
	nodeContext.SetLocalName(nameService.GetLocalName())
	nodeContext.LocalIp, _ = DNS.ExternalIP()
	mp = MP.NewMessagePasser(nodeContext.LocalName)
	streamer = Streamer.NewStreamer(mp, nodeContext)


	// We use for loop to connect with all supernode one-by-one,
	// if a connection to one supernode fails, an error message
	// will be sent by messagePasser, and this message is further
	// processed in error handler.
	// init_fail: used in hello phase
	// exit: used when all supernode cannot be connected.
	mp.AddMappings([]string{"exit", "init_fail", "super_fail", "ack"})

	// Initialize all the package structs

	// Define all the channel names and the binded functions
	// TODO: Register your channel name and binded eventhandlers here
	// The map goes as  map[channelName][eventHandler]
	// All the messages with type channelName will be put in this channel by messagePasser
	// Then the binded handler of this channel will be called with the argument (*Message)

	channelNames := map[string]func(*MP.Message){
		"election_assign":     joinAssign,
		"error" : errorHandler,

		// The streaming related handlers goes here
		"streaming_election": streamer.HandleElection,
		"streaming_join": streamer.HandleJoin,
		"streaming_data": streamer.HandleStreamerData,
		"streaming_stop": streamer.HandleStop,
		"streaming_assign": streamer.HandleAssign,
		"streaming_new_program": streamer.HandleNewProgram,
		"streaming_stop_program": streamer.HandleStopProgram,
		"streaming_quit": streamer.HandleChildQuit,
	}

	// Init and listen
	for channelName, handler := range channelNames {
		// Init all the channels listening on
		mp.Messages[channelName] = make(chan *MP.Message)
		// Bind all the functions listening on the channel
		go listenOnChannel(channelName, handler)
	}
	go nodeJoin(IPs)
	go NodeCLIInterface(streamer)
	webInterface(streamer)
	exitMsg := <- mp.Messages["exit"]
	var exitData string
	MP.DecodeData(&exitData, exitMsg.Data)
	fmt.Printf("Node: receiving force exit message [%s], node exit\n", exitData);
}

/* Join the network */
func nodeJoin(IPs []string) {
	/* If there are no running super nodes, exit program */
	if (len(IPs) == 0){
		exitMsg := MP.NewMessage("self", nodeContext.LocalName, "exit", MP.EncodeData("All supernodes are down, exit"))
		mp.Messages["exit"] <- &exitMsg
		return
	}

	//Send hello messages until find out a working supernode
	i := 0
	helloMsg := MP.NewMessage(IPs[i], "", "election_hello", MP.EncodeData("hello, my name is Bay Max, you personal healthcare companion"))
	mp.Send(helloMsg)
	fmt.Printf("Node: send hello message to SuperNode [%s]\n", IPs[i])
	for {
		select {
		case msg := <-mp.Messages["init_fail"]:
			// wait and retry the next
			errInfo := MP.FailClientInfo{}
			MP.DecodeData(&errInfo, msg.Data)
			fmt.Printf("Connetion to spernode failed: %s\n", errInfo.ErrMsg)
			i += 1
			if (i == len(IPs)) {
				exitMsg := MP.NewMessage("self", nodeContext.LocalName, "exit", MP.EncodeData("All supernodes are down, exit"))
				mp.Messages["exit"] <- &exitMsg
				break;
			}
			helloMsg := MP.NewMessage(IPs[i], "", "election_hello", MP.EncodeData("hello, my name is Bay Max, you personal healthcare companion"))
			mp.Send(helloMsg)
		case <- mp.Messages["super_fail"]:
			i += 1
			if (i == len(IPs)) {
				exitMsg := MP.NewMessage("self", nodeContext.LocalName, "exit", MP.EncodeData("All supernodes are down, exit"))
				mp.Messages["exit"] <- &exitMsg
				break;
			}
			helloMsg := MP.NewMessage(IPs[i], "", "election_hello", MP.EncodeData("hello, my name is Bay Max, you personal healthcare companion"))
			mp.Send(helloMsg)
		}
	}

}




func printHelp(){
	fmt.Println("Enter P Key to retrive Parent Info")
	fmt.Println("	   C Key to leave from parent node")
	fmt.Println("	   R Key to reconnect parent node")
	fmt.Println("      S Key to start a Streaming")
	fmt.Println("      T Key to stop a Streaming")
	fmt.Println("      J Key to join a Streaming")
	fmt.Println("      D Key to send streaming data")
	fmt.Println("      L Key to print log")
	fmt.Println("      H for help")
	fmt.Println("      Q to quit")
}

func NodeCLIInterface(streamer *Streamer.Streamer){
	printHelp()
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		line := scanner.Text()
		if ((line == "Q") || (line == "q")) {
			os.Exit(0)
		} else if ((line == "H") || (line == "h")){
			printHelp()
		} else {
			inputs := strings.Split(strings.TrimSpace(line), " ")
			switch inputs[0] {
			case "P", "p", "Parent", "parent":
				fmt.Printf("Node: print parent info IP: [%s], name [%s]\n", nodeContext.ParentIP, nodeContext.ParentName)
			case "C", "c", "Leave", "leave":
				isSendHeartBeat = false
			case "R", "r", "Reconnect", "reconnect":
				isSendHeartBeat = true
			case "S","s", "start", "Start":
				if len(inputs) > 1 {
					streamer.Start(inputs[1])
				}
			case "T","t", "Stop", "stop":
				streamer.Stop()
			case "J","j", "Join", "join":
				if len(inputs) > 1 {
					streamer.Join(inputs[1])
				}
			case "D","d", "Stream", "stream":
				if len(inputs) > 1 {
					streamer.Stream(inputs[1])
				}
			case "L","l", "Log", "log":
				streamer.Log()
			case "Receive", "receive":
				fmt.Println("Received :" + streamer.Receive())
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