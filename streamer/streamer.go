package streamer

import (
	MP "../messagePasser"
	"strconv"
)

const(
	IDEAL = iota
	JOINING
	STREAMING
)

//streamId := 0

type Streamer struct{
	STATE int
	StreamingParent string
	Streamingchildren []string
	mp *MP.MessagePasser
	ProgramList map[string]string
	// TODO: Change to []byte and handle video data
	StreamingData chan string
	ReceivingData chan string
}

/*
Initialization
 */
func NewStreamer(mp *MP.MessagePasser) *Streamer{
	streamer := Streamer{STATE:IDEAL, mp: mp}
	streamer.StreamingData = make(chan string)
	streamer.ReceivingData = make(chan string)
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
func (streamer *Streamer) Start(){
	if streamer.STATE != IDEAL{
		return
	}
	streamer.STATE = STREAMING
}

/* The streamer quit the streaming process */
func (streamer *Streamer) Stop(){
	if streamer.STATE != STREAMING{
		return
	}
	streamer.STATE = IDEAL
}

/* A node request to join a certain program */
func (streamer *Streamer) Join(programId string){
	if streamer.STATE != IDEAL{
		return
	}
	streamer.STATE = JOINING
}

/* Was assign to parentId*/
func (streamer *Streamer) Assign(parentId string){
	streamer.StreamingParent = parentId
	streamer.STATE = STREAMING
}

/* Called to stream the data */
func (streamer *Streamer) Stream(data string){
	streamer.StreamingData <- data
}

/********************************************************************************/
/*   When in the STATE of streaming*/
/********************************************************************************/
/* Being selected in the DHT and handle the join request*/
func (streamer *Streamer) HandleJoin(msg *MP.Message){
	var controlData StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)

	// TODO: Actually start the election below

	// Fake here

}

/* Election related messgaes*/
func (streamer *Streamer) HandleElection(msg *MP.Message){
	// TODO: Election
}

/* A New Program is added */
func (streamer *Streamer) HandleNewProgram(msg *MP.Message){
	var controlData StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	streamer.ProgramList[controlData.SrcName + "[" + strconv.Itoa(controlData.StreamID) + "]"] = controlData.Title
}

/* A program is stoped */
func (streamer *Streamer) HandleStopProgram(msg *MP.Message) {
	var controlData StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	delete(streamer.ProgramList, controlData.SrcName + "[" + strconv.Itoa(controlData.StreamID) + "]")
}

/* Handle the receiving streaming data*/
func (streamer *Streamer) HandleStreamerData(msg *MP.Message) {
	var data string
	MP.DecodeData(&data, msg.Data)
	streamer.ReceivingData <- data
}


