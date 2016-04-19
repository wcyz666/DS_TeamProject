package streamer

import (
	MP "../../messagePasser"
	SDataType "../"
	NC "../../node/nodeContext"
	"fmt"
)

const(
	IDEAL = iota
	JOINING
	STREAMING
)


type Streamer struct{
	nodeContext *NC.NodeContext
	mp *MP.MessagePasser
	STATE int
	StreamingParent string
	Streamingchildren []string
	CurrentProgram string
	ProgramList map[string]string
	// TODO: Change to []byte and handle video data
	StreamingData chan string
	ReceivingData chan string
	streamID int
}

/*
Initialization
 */
func NewStreamer(mp *MP.MessagePasser, nodeContext *NC.NodeContext) *Streamer{
	streamer := Streamer{STATE:IDEAL, mp: mp, streamID:0, nodeContext:nodeContext}
	streamer.StreamingData = make(chan string, 1000)
	streamer.ReceivingData = make(chan string, 1000)
	streamer.ProgramList = make(map[string]string)
	go streamer.backgroundStreaming()
	//go streamer.testReceive()
	return &streamer
}

/* Temp function to show the receiving data*/
func (streamer *Streamer)testReceive(){
	for{
		data := <-streamer.ReceivingData
		fmt.Println("Received: " + data)
	}
}


/* This function keep distributing the data in the channel to all the recipients*/
func (streamer *Streamer) backgroundStreaming(){
	for{
		data := <- streamer.StreamingData
		for _, destName := range(streamer.Streamingchildren){
			msg := MP.NewMessage("", destName, "streaming_data", MP.EncodeData(data))
			//fmt.Print("Sending streaming data:" )
			//fmt.Println(msg)
			go streamer.mp.Send(msg)
		}
	}
}




/********************************************************************************/
/*   When in the STATE of streaming*/
/********************************************************************************/

/* Be assigned to a node */
func (streamer *Streamer) HandleAssign(msg *MP.Message){
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	fmt.Println("Handling assign! As child of " + controlData.SrcName)
	streamer.StreamingParent = controlData.SrcName
	streamer.STATE = STREAMING
	streamer.CurrentProgram = controlData.RootStreamer
}

/* Being selected in the DHT and handle the join request*/
func (streamer *Streamer) HandleJoin(msg *MP.Message){
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)

	fmt.Println("Handling join request of " + controlData.SrcName)

	// TODO: Actually start the election below

	// Fake now
	// Construct data, to be used by updating the program


	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		SrcIp:  streamer.nodeContext.LocalIp,
		RootStreamer: controlData.RootStreamer,
	}

	// Store this child
	streamer.Streamingchildren = append(streamer.Streamingchildren, controlData.SrcName)
	// Notify the src (the one join the network first)

	assignMsg := MP.NewMessage(controlData.SrcIp, controlData.SrcName, "streaming_assign", MP.EncodeData(data))
	//fmt.Println(assignMsg)
	streamer.mp.Send(assignMsg)
}

/* Election related messgaes*/
func (streamer *Streamer) HandleElection(msg *MP.Message){
	// TODO: Election
}

/* Program stopped */
func (streamer *Streamer) HandleStop(msg *MP.Message){
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)

	// Clear
	streamer.StreamingParent = ""
	streamer.Streamingchildren = []string{}
	streamer.STATE = IDEAL

	// Notify Stream Children
	for _, destName := range(streamer.Streamingchildren){
		msg := MP.NewMessage("", destName, "streaming_stop", msg.Data)
		go streamer.mp.Send(msg)
	}

	if controlData.RootStreamer != streamer.CurrentProgram &&
		controlData.SrcName != streamer.nodeContext.LocalName{
		streamer.Join(streamer.CurrentProgram)
	}
}

/* A New Program is added */
func (streamer *Streamer) HandleNewProgram(msg *MP.Message) {
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	// If this program is not started by myself
	if controlData.SrcName != streamer.nodeContext.LocalName {
		streamer.ProgramList[controlData.SrcName] = controlData.Title
		fmt.Println("New program detected! ")
		fmt.Print(" Current program list:")
		fmt.Println(streamer.ProgramList)
	}
}

/* A program is stoped */
func (streamer *Streamer) HandleStopProgram(msg *MP.Message) {
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	delete(streamer.ProgramList, controlData.SrcName)
	fmt.Println("Program deleted! ")
	fmt.Print(" Current program list:")
	fmt.Println(streamer.ProgramList)
}


func (streamer *Streamer) HandleChildQuit(msg *MP.Message) {
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	streamer.deleteStreamingChild(controlData.SrcIp, controlData.SrcName)
}


/* Handle the receiving streaming data*/
func (streamer *Streamer) HandleStreamerData(msg *MP.Message) {
	var data string
	MP.DecodeData(&data, msg.Data)
	streamer.ReceivingData <- data
	streamer.StreamingData <- data
}




/********************************************************************************/
/*  Error handling related functions:
/* 	1. Parent/child quit streaming
/*	2. Parent/child accidentally failed
/********************************************************************************/

/* Take care of the tcp connection error message */
func (streamer *Streamer) HandleErrorMsg(msg *MP.Message) {
	failNode := MP.FailClientInfo{}
	MP.DecodeData(&failNode, msg.Data)

	// We use name as identifier for streaming parents/ children
	// Streaming parent fails
	if failNode.Name == streamer.StreamingParent{
		// Rejoin the streaming process
		fmt.Println("[streaming] Parent node " + failNode.Name + " failed!")
		streamer.StreamingParent = ""
		streamer.Streamingchildren = []string{}
		streamer.STATE = IDEAL
		if failNode.Name != streamer.CurrentProgram{
			streamer.Join(streamer.CurrentProgram)
		}

	}else{
		streamer.deleteStreamingChild(failNode.IP, failNode.Name)
	}
}

func (streamer *Streamer) deleteStreamingChild(childIP string, childName string){
	for index, child := range(streamer.Streamingchildren){
		// If is a children in the streaming group
		if child == childName{
			fmt.Println("[streaming] Children node " + childName + " quit!")
			// Delete this node in the children array
			length := len(streamer.Streamingchildren)
			streamer.Streamingchildren[index] = streamer.Streamingchildren[length-1]
			streamer.Streamingchildren[length-1] = ""
			streamer.Streamingchildren = streamer.Streamingchildren[:length-1]

			// Notify supernode to delete this child in DHT
			removeData := SDataType.RemoveFromDht{
				RootStreamer:streamer.CurrentProgram,
				FailNodeIp: childIP,
				FailNodeName: childName,
			}
			removeMsg := MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName,
				"stream_delete_from_dht", MP.EncodeData(removeData))
			streamer.mp.Send(removeMsg)
			break
		}
	}
}

func (streamer *Streamer) rejoin(root string){
	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		SrcIp: streamer.nodeContext.LocalIp,
		RootStreamer: root,
		Type: "rejoin",
	}
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_join", MP.EncodeData(data)))
	streamer.STATE = JOINING
}
