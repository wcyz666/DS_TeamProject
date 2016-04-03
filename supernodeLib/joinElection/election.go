package joinElection

import (
	//"fmt"

	MP "../../messagePasser"
	//dns "../dnsService"

)

/**
The package takes care of all the conditions a new node join the network
*/

type JoinElection struct {
	mp *MP.MessagePasser
}

/* Constructor */
func NewJoinElection(mp *MP.MessagePasser) *JoinElection {
	j := JoinElection{mp: mp}
	return &j
}

func (j *JoinElection) Start(msg *MP.Message) {
	// Start the election process below
	// TODO: Actually implement the election algorithm


	// Current directly take over the child node
	// Send assign message
	childNodeAddr := msg.Src
	childName := msg.SrcName
	kind := "join_assign"
	j.mp.Send(MP.NewMessage(childNodeAddr, childName, kind, "hehehe"))
}

func (j *JoinElection) Receive(msg *MP.Message) {
	// Deal with the received messages
}
