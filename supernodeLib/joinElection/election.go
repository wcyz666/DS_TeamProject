package joinElection

import (
	MP "../../messagePasser"
	DNS "../../dnsService"
	LNS "../../localNameService"
	"fmt"
	DHT "../../dht"
	SC "../../superNode/superNodeContext"
)

/**
The package takes care of all the conditions a new node join the network
*/


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
	kind := "election"
	myIP, _ := DNS.ExternalIP()
	eBMsgPayload := ElectionBroadcastMessage{IP: myIP, Name: LNS.GetLocalName(), ChildCount: j.superNodeContext.GetNodeCount()}
	payload := MP.NewMessage(childNodeAddr, childName, kind, MP.EncodeData(eBMsgPayload))

	//If only me, then election is completed: send my info back
	if j.dht.AmITheOnlyNodeInDHT() {
		j.mp.Messages["election_complete"] <- &payload
	} else {
		eBMsg := j.dht.NewBroadcastMessage()
		j.dht.PassBroadcastMessage(eBMsg, &payload)
	}

}

func (j *JoinElection) ForwardElection(msg *MP.Message) {
	// Deal with the received messages
	fmt.Println(j.getPrevElectionMessage(msg))
}

// de-capsulate and get ElectionBroadcastMessage back
func (j *JoinElection) getPrevElectionMessage(msg *MP.Message) *ElectionBroadcastMessage {
	var payloadMsg MP.Message
	var eBMsg ElectionBroadcastMessage
	MP.DecodeData(&payloadMsg, j.dht.GetPayload(msg))
	MP.DecodeData(&eBMsg, payloadMsg.Data)

	return &eBMsg
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