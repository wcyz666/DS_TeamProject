package joinElection

import (
	MP "../../messagePasser"
	"fmt"
	DHT "../../dht"
	SC "../../superNode/superNodeContext"
)

/**
The package takes care of all the conditions a new node join the network
*/
const E_KIND  = "election"


type JoinElection struct {
	mp *MP.MessagePasser
	dht *DHT.DHTNode
	superNodeContext *SC.SuperNodeContext
}

/* Constructor */
func NewJoinElection(mp *MP.MessagePasser, dht *DHT.DHTNode, snContext *SC.SuperNodeContext) *JoinElection {
	j := JoinElection{mp: mp, dht: dht, superNodeContext: snContext}
	return &j
}

func (j *JoinElection) StartElection(msg *MP.Message) {
	// Start the election process below
	fmt.Printf("Supernode: election start for node [%s]\n", msg.SrcName)
	// TODO: Actually implement the election algorithm

	//Generate payload. This will be transmitted over the DHT ring
	childNodeAddr := msg.Src
	childName := msg.SrcName

	eBMsgPayload := NewElectionBroadcastMessage(j.superNodeContext.IP, j.superNodeContext.LocalName, j.superNodeContext.GetNodeCount())
	payload := MP.NewMessage(childNodeAddr, childName, E_KIND, MP.EncodeData(eBMsgPayload))

	//If only me, then election is completed: send my info back
	if j.dht.AmITheOnlyNodeInDHT() {
		j.mp.Messages["election_complete"] <- &payload
	} else {
		eBMsg := j.dht.NewBroadcastMessage()
		j.dht.PassBroadcastMessage(&eBMsg, &payload)
	}

}

func (j *JoinElection) ForwardElection(msg *MP.Message) {
	// Deal with the received messages
	bMsg, payloadMsg, eBMsg := j.getPrevElectionMessage(msg)

	//evict the cache

	if (j.dht.IsBroadcastOver(bMsg)) {
		fmt.Print("Election: election over, result: ")
		fmt.Println(eBMsg)
		j.mp.Messages["election_complete"] <- payloadMsg
	} else {
		payloadMsg.Data = MP.EncodeData(j.compareAndUpdatePayload(eBMsg))
		j.dht.AppendSelfToBroadcastTrack(bMsg)
		j.dht.PassBroadcastMessage(bMsg, payloadMsg)
		fmt.Println("Election: forward election message to the next supernode")
	}
}



func (j *JoinElection) compareAndUpdatePayload(eBMsg *ElectionBroadcastMessage) *ElectionBroadcastMessage {
	if (eBMsg.ChildCount > j.superNodeContext.GetNodeCount()) {
		eBMsg.Name = j.superNodeContext.LocalName
		eBMsg.IP = j.superNodeContext.IP
		eBMsg.ChildCount = j.superNodeContext.GetNodeCount()
		fmt.Print("Election: Better option for a parent, new parent")
		fmt.Println(eBMsg)
	}
	return eBMsg
}

// decapsulate and get ElectionBroadcastMessage back
func (j *JoinElection) getPrevElectionMessage(msg *MP.Message) (*DHT.BroadcastMessage, *MP.Message, *ElectionBroadcastMessage) {
	var payloadMsg MP.Message
	var eBMsg ElectionBroadcastMessage
	bMsg := j.dht.GetBroadcastMessage(msg)
	MP.DecodeData(&payloadMsg, bMsg.Payload)
	MP.DecodeData(&eBMsg, payloadMsg.Data)

	return bMsg, &payloadMsg, &eBMsg
}

func (j *JoinElection) CompleteElection(msg *MP.Message) {

	// Deal with the received messages
	result := transferEbmToResult(msg)
	msg.Data = MP.EncodeData(result)
	msg.Kind = "election_assign"
	j.mp.Send(*msg)
}

// Transform an ElectionBroadcastingMessage to ElectionResult
func transferEbmToResult(msg *MP.Message) *ElectionResult {
	var eBMsg ElectionBroadcastMessage
	MP.DecodeData(&eBMsg, msg.Data)
	return &ElectionResult{ParentName: eBMsg.Name, ParentIP: eBMsg.IP}
}