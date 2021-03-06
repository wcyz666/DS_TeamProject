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

	if streamer.STATE != IDEAL{
		return
	}
	fmt.Println("Start to streaming a program!")
	// Update local program id
	streamer.streamID += 1

	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		SrcIp: streamer.nodeContext.LocalIp,
		StreamID: streamer.streamID,
		Title: title,
		RootStreamer: streamer.nodeContext.LocalName,
	}

	fmt.Print("Sending confirming message ")
	fmt.Print(data)
	fmt.Println(" to the supernode " + streamer.nodeContext.ParentName)
	// Notify the parent
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_start", MP.EncodeData(data)))
	streamer.CurrentProgram = streamer.nodeContext.LocalName
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
		SrcIp: streamer.nodeContext.LocalIp,
		StreamID: streamer.streamID,
		RootStreamer: streamer.nodeContext.LocalName,
	}

	// Notify the parent
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_stop", MP.EncodeData(data)))

	// Notify Stream Parent
	if streamer.StreamingParent != ""{
		streamer.mp.Send(MP.NewMessage("", streamer.StreamingParent, "streaming_quit", MP.EncodeData(data)))
	}

	// Notify Stream Children
	msg := MP.NewMessage(streamer.nodeContext.LocalIp, streamer.nodeContext.LocalName, "streaming_stop", MP.EncodeData(data))
	//streamer.mp.Send(msg)
	streamer.HandleStop(&msg)
	streamer.STATE = IDEAL
}

/* A node request to join a certain program */
func (streamer *Streamer) Join(root string){
	fmt.Println("Request to join a program! " + root)


	if streamer.STATE != IDEAL {
		return
	}

	// If program is not in the list
	if _,ok := streamer.ProgramList[root]; ok == false{
		return
	}

	// Construct data, to be used by updating the program
	data := SDataType.StreamControlMsg{
		SrcName: streamer.nodeContext.LocalName,
		SrcIp: streamer.nodeContext.LocalIp,
		RootStreamer: root,
		Title: streamer.ProgramList[root],
	}


	// Notify the parent
	fmt.Print("Sending confirming message ")
	fmt.Print(data)
	fmt.Println(" to the supernode " + streamer.nodeContext.ParentName)
	streamer.mp.Send(MP.NewMessage(streamer.nodeContext.ParentIP, streamer.nodeContext.ParentName, "stream_join", MP.EncodeData(data)))


	streamer.STATE = JOINING
}


/* Called to stream the data */
func (streamer *Streamer) Stream(data string){
	if streamer.STATE != STREAMING{
		return
	}
	streamer.StreamingData <- data
}

/* Called to stream the data */
func (streamer *Streamer) Receive() (data string){
	select {
	case data := <-streamer.ReceivingData:
		return data
	default:
		return ""
	}
}


/* Log the essential information */
func (streamer *Streamer) Log(){
	fmt.Println("#################################")
	fmt.Println("Local name: " + streamer.nodeContext.LocalName)
	fmt.Println("Parent supernode: " + streamer.nodeContext.ParentName + " IP: " + streamer.nodeContext.ParentIP)
	fmt.Print("Streaming State: ")
	fmt.Println(streamer.STATE)
	fmt.Println("Streaming parent: " + streamer.StreamingParent)
	fmt.Print("Streaming children: ")
	fmt.Println(streamer.Streamingchildren)
	fmt.Print("Current program list: ")
	fmt.Println(streamer.ProgramList)
	fmt.Println("#################################")
}