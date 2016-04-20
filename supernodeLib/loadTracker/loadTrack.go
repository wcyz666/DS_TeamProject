package loadTracker

import (
	DHT "../../dht"
	MP "../../messagePasser"
	SC "../../superNode/superNodeContext"
	"fmt"
)

/**
The package takes care of all the conditions a new node join the network
*/

const (
	LT_KIND = "loadtrack"
)

type LoadTracker struct {
	mp               *MP.MessagePasser
	dht              *DHT.DHTNode
	superNodeContext *SC.SuperNodeContext
    cache            *LoadBroadcastMessage
}

/* Constructor */
func NewLoadTracker(mp *MP.MessagePasser, dht *DHT.DHTNode, snContext *SC.SuperNodeContext) *LoadTracker {
	l := LoadTracker{mp: mp, dht: dht, superNodeContext: snContext}
	return &l
}

func (l *LoadTracker) StartLoadTrack(msg *MP.Message) {
	// Start the election process below
	fmt.Printf("Supernode: election start for node [%s]\n", msg.SrcName)
	// TODO: Actually implement the election algorithm

	//Generate payload. This will be transmitted over the DHT ring
	childNodeAddr := msg.Src
	childName := msg.SrcName

	myLoad := l.getCurLoad()
	uTBMsgPayload := NewTracker(childName, childNodeAddr, myLoad)
	payload := MP.NewMessage(childNodeAddr, childName, LT_KIND, MP.EncodeData(uTBMsgPayload))

	//If only me, then Load tracking is completed: send my info back
	if l.dht.AmITheOnlyNodeInDHT() {
		l.cache = uTBMsgPayload
		l.mp.Messages["loadtrack_complete"] <- &payload
		fmt.Println("Load Track: Single SuperNode mode, Load track end.")
	//If the cache is not evicted, return cache.
	} else if l.superNodeContext.IsLoadCacheEffective {
		payload.Data = MP.EncodeData(l.cache)
		l.mp.Messages["loadtrack_complete"] <- &payload
		fmt.Println("Load Track: cache hit, Load track end.")
	} else {
		uTBMsg := l.dht.NewBroadcastMessage()
		l.dht.PassBroadcastMessage(&uTBMsg, &payload)
	}
}

func (l *LoadTracker) ForwardTrack(msg *MP.Message) {
	// Deal with the received messages
	bMsg, payloadMsg, eBMsg := l.getPrevElectionMessage(msg)

	if l.dht.IsBroadcastOver(bMsg) {
		fmt.Print("Load Track: tracking over, result: ")
		fmt.Println(eBMsg)
		l.mp.Messages["loadtrack_complete"] <- payloadMsg
	} else {
		payloadMsg.Data = MP.EncodeData(l.UpdatePayload(eBMsg))
		l.dht.AppendSelfToBroadcastTrack(bMsg)
		l.dht.PassBroadcastMessage(bMsg, payloadMsg)
		fmt.Println("Load Track: forward tracking message to the next supernode")
	}
}

func (l *LoadTracker) UpdatePayload(uTBMsg *LoadBroadcastMessage) *LoadBroadcastMessage {
	uTBMsg.SuperNodeUsages = append(uTBMsg.SuperNodeUsages, l.getCurLoad())
	return uTBMsg
}

func (l *LoadTracker) getCurLoad() SuperNodeUsage {
	return NewUsage(l.superNodeContext.IP, l.superNodeContext.LocalName, l.superNodeContext.GetNodeCount())
}

// de-capsulate and get ElectionBroadcastMessage back
func (j *LoadTracker) getPrevElectionMessage(msg *MP.Message) (*DHT.BroadcastMessage, *MP.Message, *LoadBroadcastMessage) {
	var payloadMsg MP.Message
	var uTBMsg LoadBroadcastMessage
	bMsg := j.dht.GetBroadcastMessage(msg)
	MP.DecodeData(&payloadMsg, bMsg.Payload)
	MP.DecodeData(&uTBMsg, payloadMsg.Data)

	return bMsg, &payloadMsg, &uTBMsg
}

func (l *LoadTracker) CompleteTracking(msg *MP.Message) {
	// Deal with the received messages
	l.superNodeContext.IsLoadCacheEffective = true
	result := transferUtbmToResult(msg)
	l.cache = result
	msg.Data = MP.EncodeData(result)
	msg.Kind = "loadtrack_result"
	l.mp.Send(*msg)
}

// Transform an ElectionBroadcastingMessage to ElectionResult
func transferUtbmToResult(msg *MP.Message) *LoadBroadcastMessage {
	var uTBMsg LoadBroadcastMessage
	MP.DecodeData(&uTBMsg, msg.Data)
	return &uTBMsg
}
