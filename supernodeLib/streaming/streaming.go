package streaming

import (
	dht "../../dht"
	messagePasser "../../messagePasser"
)

/**
The stream handler class
This class takes care of
	 1. The related requests from the normal nodes
	 2. Sync related information in the DHT and with other supernodes
*/
type StreamingHandler struct {
	dHashtable *dht.DHTService
	mp         *messagePasser.MessagePasser
}

func NewStreamingHandler(dHashtable *dht.DHTService, mp *messagePasser.MessagePasser) *StreamingHandler {
	return &StreamingHandler{dHashtable: dHashtable, mp: mp}
}

/* A node starts to stream messages */
func (sHandler *StreamingHandler) StreamStart(msg *messagePasser.Message) {

}

/* A node asks the up-to-date stream list */
func (sHandler *StreamingHandler) StreamGetList(msg *messagePasser.Message) {

}

/* A child node asks to join a certain streaming group */
func (sHandler *StreamingHandler) StreamJoin(msg *messagePasser.Message) {

}
