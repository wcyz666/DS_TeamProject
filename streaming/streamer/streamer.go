package streamer

import (
	MP "../../messagePasser"
	"strconv"
	SDataType "../"
	NC "../../node/nodeContext"
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
func NewStreamer(mp *MP.MessagePasser, ) *Streamer{
	streamer := Streamer{STATE:IDEAL, mp: mp, streamID:0}
	streamer.StreamingData = make(chan string)
	streamer.ReceivingData = make(chan string)
	go streamer.backgroundStreaming()
	return &streamer
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


/*
The control flow related functions
*/

/* The request to be the first to start a streming */
func (streamer *Streamer) Start(title string){
	if streamer.STATE != IDEAL{
		return
	}

	// Update local program id
	streamer.streamID += 1

	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		StreamID: streamer.streamID,
		Title: title,
	}

	// Notify the parent
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_start", MP.EncodeData(data)))
	streamer.STATE = STREAMING
}

/* The streamer quit the streaming process */
func (streamer *Streamer) Stop(){
	if streamer.STATE != STREAMING{
		return
	}

	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		StreamID: streamer.streamID,
	}

	// Notify the parent
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_stop", MP.EncodeData(data)))

	// Notify Stream Children
	streamer.HandleStop(MP.NewMessage("", "", "streaming_stop", MP.EncodeData("")))


	streamer.STATE = IDEAL
}

/* A node request to join a certain program */
func (streamer *Streamer) Join(root string){
	if streamer.STATE != IDEAL {
		return
	}

	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		RootStreamer: root,
	}

	// Notify the parent
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_join", MP.EncodeData(data)))


	streamer.STATE = JOINING
}


/* Called to stream the data */
func (streamer *Streamer) Stream(data string){
	streamer.StreamingData <- data
}

/********************************************************************************/
/*   When in the STATE of streaming*/
/********************************************************************************/

/* Be assigned to a node */
func (streamer *Streamer) HandleAssign(msg *MP.Message){
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)

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
	streamer.ProgramList[controlData.SrcName + "[" + strconv.Itoa(controlData.StreamID) + "]"] = controlData.Title
}

/* A program is stoped */
func (streamer *Streamer) HandleStopProgram(msg *MP.Message) {
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	delete(streamer.ProgramList, controlData.SrcName + "[" + strconv.Itoa(controlData.StreamID) + "]")
}

/* Handle the receiving streaming data*/
func (streamer *Streamer) HandleStreamerData(msg *MP.Message) {
	var data string
	MP.DecodeData(&data, msg.Data)
	streamer.ReceivingData <- data
}


