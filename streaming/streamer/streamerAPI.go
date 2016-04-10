package streamer

import (
	MP "../../messagePasser"
	SDataType "../"
	"fmt"
)

/*
 This file includes all the apis provided by the streamer to the client app
 */

/* The request to be the first to start a streming */
func (streamer *Streamer) Start(title string){
	fmt.Println("Start to sttreaming a program!")
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

	fmt.Println("Sending confirming message " + data + " to the supernode " + streamer.nodeContext.ParentName)
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
		RootStreamer: streamer.nodeContext.LocalName,
	}

	// Notify the parent
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_stop", MP.EncodeData(data)))

	// Notify Stream Children
	streamer.HandleStop(nil)


	streamer.STATE = IDEAL
}

/* A node request to join a certain program */
func (streamer *Streamer) Join(root string){
	fmt.Println("Request to join a program! " + root)
	if streamer.STATE != IDEAL {
		return
	}

	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		RootStreamer: root,
	}


	// Notify the parent
	fmt.Println("Sending confirming message " + data + " to the supernode " + streamer.nodeContext.ParentName)
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_join", MP.EncodeData(data)))


	streamer.STATE = JOINING
}


/* Called to stream the data */
func (streamer *Streamer) Stream(data string){
	streamer.StreamingData <- data
}

