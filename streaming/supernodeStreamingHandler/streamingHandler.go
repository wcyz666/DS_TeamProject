package streaming

import (
	DHT "../../dht"
	MP "../../messagePasser"
	SNC "../../superNode/superNodeContext"
	SDataType "../"
	"fmt"
	"../../config"
	DNS "../../dnsService"
)

/**
The stream handler class
This class takes care of
	 1. The related requests from the normal nodes
	 2. Sync related information in the DHT and with other supernodes
*/
type StreamingHandler struct {
	dht *DHT.DHTService
	mp         *MP.MessagePasser
	superNodeContext	*SNC.SuperNodeContext
}

func NewStreamingHandler(dHashtable *DHT.DHTService, mp *MP.MessagePasser, superNodeContext *SNC.SuperNodeContext) *StreamingHandler {
	return &StreamingHandler{dht: dHashtable, mp: mp, superNodeContext:superNodeContext}
}

/* Broadcast service */
func (sHandler *StreamingHandler) broadcast(channelName string, data []byte){
	// Get all the supernodes
	IPs := DNS.GetAddr(config.BootstrapDomainName)
	for _, ip := range(IPs) {
		sHandler.mp.Send(MP.NewMessage(ip, "", channelName, data))
	}
}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStart(msg *MP.Message) {
	//Notify all the supernodes
	//Including itself (Since one supernode may have multipy children)
	sHandler.broadcast("stream_program_start", msg.Data)

	//TODO: Update DHT table
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	sHandler.dht.Create(controlData.RootStreamer, DHT.MemberShipInfo{SuperNodeIp:controlData.SrcName})

}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStop(msg *MP.Message) {
	//Notify all the supernodes
	sHandler.broadcast("stream_program_stop", msg.Data)

	//TODO: Update DHT table
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)
	sHandler.dht.Delete(controlData.RootStreamer)
}

/* Update broadcasted from other supernodes */
func (sHandler *StreamingHandler) StreamProgramStart(msg *MP.Message) {
	//Notify the children the new program
	childrenNames := sHandler.superNodeContext.GetAllChildrenName()
	fmt.Println("New programs detected! Sending to all children")
	fmt.Println(childrenNames)
	for _, child := range(childrenNames){
		sHandler.mp.Send(MP.NewMessage("", child, "streaming_new_program", msg.Data))
	}
}

/* Update broadcasted from other supernodes */
func (sHandler *StreamingHandler) StreamProgramStop(msg *MP.Message) {
	//Notify the children the new program
	childrenNames := sHandler.superNodeContext.GetAllChildrenName()
	for _, child := range(childrenNames){
		sHandler.mp.Send(MP.NewMessage("", child, "streaming_stop_program", msg.Data))
	}
}

/* A child node asks to join a certain streaming group */
func (sHandler *StreamingHandler) StreamJoin(msg *MP.Message) {
	// Notify one of the streamers in the DHT
	var controlData SDataType.StreamControlMsg
	MP.DecodeData(&controlData, msg.Data)

	root := controlData.RootStreamer
	fmt.Println(root)

	// TODO: find the streaming group with root in the DHT and update it
	streamers, _ := sHandler.dht.Get(root)
	// Choose the last streamer to start the election
	parentName := streamers[len(streamers)-1].SuperNodeIp
	// TODO: Send "streaming_join" to one of the streamers
	sHandler.mp.Send(MP.NewMessage("", parentName, "streaming_join", msg.Data))
}