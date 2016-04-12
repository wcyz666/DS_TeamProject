package joinElection

import (
	//"fmt"

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
	fmt.Println("Supernode: election start for node [%s]\n", msg.SrcName)
	// TODO: Actually implement the election algorithm

	//Generate payload. This will be transmitted over the DHT ring
	childNodeAddr := msg.Src
	childName := msg.SrcName
	kind := "election_assign"
	myIP, _ := DNS.ExternalIP()
	eBMsg := ElectionBroadcastMessage{IP: myIP, Name: LNS.GetLocalName(), ChildCount: j.superNodeContext.GetNodeCount()}
	payload := MP.NewMessage(childNodeAddr, childName, kind, MP.EncodeData(eBMsg))

	//If only me, then election is completed: send my info back
	if j.dht.AmITheOnlyNodeInDHT() {
		j.mp.Messages["election_complete"] <- &payload
	}

}

func (j *JoinElection) ForwardElection(msg *MP.Message) {
	// Deal with the received messages

}

func (j *JoinElection) CompleteElection(msg *MP.Message) {
	// Deal with the received messages
	//result := ElectionResult{ParentIP: myIP, ParentName: LNS.GetLocalName()}
	j.mp.Send(*msg)
}
