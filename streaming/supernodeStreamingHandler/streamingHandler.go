package streaming

import (
	dht "../../dht"
	MP "../../messagePasser"
	SNC "../../superNode/superNodeContext"
	SDataType "../"
	"fmt"
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

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStart(msg *MP.Message) {
	//TODO: Notify all the supernodes

	// TODO: NOTIFY all children

	//TODO: Update DHT table


}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStop(msg *MP.Message) {
	//TODO: Notify all the supernodes

	//TODO: Update DHT table

}

/* Update broadcasted from other supernodes */
func (sHandler *StreamingHandler) StreamProgramStart(msg *MP.Message) {
	//TODO: Notify the children the new program
}

/* Update broadcasted from other supernodes */
func (sHandler *StreamingHandler) StreamProgramStop(msg *MP.Message) {
	//TODO: Notify the children the new program
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