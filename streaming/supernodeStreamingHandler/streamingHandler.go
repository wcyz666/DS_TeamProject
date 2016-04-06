package streaming

import (
	dht "../../dht"
	MP "../../messagePasser"
	SNC "../../superNode/superNodeContext"
)

/**
The stream handler class
This class takes care of
	 1. The related requests from the normal nodes
	 2. Sync related information in the DHT and with other supernodes
*/
type StreamingHandler struct {
	dHashtable *dht.DHT
	mp         *MP.MessagePasser
	superNodeContext	*SNC.SuperNodeContext
}

func NewStreamingHandler(dHashtable *dht.DHT, mp *MP.MessagePasser, superNodeContext *SNC.SuperNodeContext) *StreamingHandler {
	return &StreamingHandler{dHashtable: dHashtable, mp: mp, superNodeContext:superNodeContext}
}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStart(msg *MP.Message) {
	//TODO: Notify all the supernodes

	//TODO: Update DHT table


}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStop(msg *MP.Message) {
	//TODO: Notify all the supernodes

	//TODO: Update DHT table

}

/* A node asks the up-to-date stream list */
func (sHandler *StreamingHandler) StreamNewProgram(msg *MP.Message) {
	//TODO: Notify the children the new program
}

/* A child node asks to join a certain streaming group */
func (sHandler *StreamingHandler) StreamJoin(msg *MP.Message) {
	//TODO: Notify one of the streamer in the DHT
}
