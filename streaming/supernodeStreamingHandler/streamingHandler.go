package streaming

import (
	dht "../../dht"
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
	dHashtable *dht.DHTService
	mp         *MP.MessagePasser
	superNodeContext	*SNC.SuperNodeContext
}

func NewStreamingHandler(dHashtable *dht.DHTService, mp *MP.MessagePasser, superNodeContext *SNC.SuperNodeContext) *StreamingHandler {
	return &StreamingHandler{dHashtable: dHashtable, mp: mp, superNodeContext:superNodeContext}
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


}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStop(msg *MP.Message) {
	//Notify all the supernodes
	sHandler.broadcast("stream_program_stop", msg.Data)

	//TODO: Update DHT table

}

/* Update broadcasted from other supernodes */
func (sHandler *StreamingHandler) StreamProgramStart(msg *MP.Message) {
	//Notify the children the new program
	childrenNames := sHandler.superNodeContext.GetAllChildrenName()
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
	// TODO: Send "streaming_join" to one of the streamers
}