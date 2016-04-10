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
	streamer.StreamingData = make(chan string)
	streamer.ReceivingData = make(chan string)
	go streamer.backgroundStreaming()
	go streamer.testReceive()
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
			msg := MP.NewMessage(destName, "", "streaming_data", MP.EncodeData(data))
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
}

/* Being selected in the DHT and handle the join request*/
func (streamer *Streamer) HandleJoin(msg *MP.Message){
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)

	// TODO: Actually start the election below

	// Fake now
	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		RootStreamer: controlData.RootStreamer,
	}

	// Store this child
	streamer.Streamingchildren = append(streamer.Streamingchildren, controlData.SrcName)
	// Notify the src (the one join the network first)
	streamer.mp.Send(MP.NewMessage("", controlData.SrcName, "streaming_assign", MP.EncodeData(data)))
}

/* Election related messgaes*/
func (streamer *Streamer) HandleElection(msg *MP.Message){
	// TODO: Election
}

/* Program stopped */
func (streamer *Streamer) HandleStop(msg *MP.Message){
	// Notify Stream Children
	for _, destName := range(streamer.Streamingchildren){
		msg := MP.NewMessage(destName, "", "streaming_stop", MP.EncodeData(""))
		go streamer.mp.Send(msg)
	}

	// Clear
	streamer.StreamingParent = streamer.nodeContext.LocalName
	streamer.Streamingchildren = []string{}
}

/* A New Program is added */
func (streamer *Streamer) HandleNewProgram(msg *MP.Message){
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	streamer.ProgramList[controlData.SrcName] = controlData.Title
	fmt.Println("New program detected! ")
	fmt.Println("Current program list:" + streamer.ProgramList)
}

/* A program is stoped */
func (streamer *Streamer) HandleStopProgram(msg *MP.Message) {
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	streamer.ProgramList = delete(streamer.ProgramList, controlData.SrcName)
	fmt.Println("Program deleted! ")
	fmt.Println("Current program list:" + streamer.ProgramList)
}

/* Handle the receiving streaming data*/
func (streamer *Streamer) HandleStreamerData(msg *MP.Message) {
	var data string
	MP.DecodeData(&data, msg.Data)
	streamer.ReceivingData <- data
	streamer.StreamingData <- data
}