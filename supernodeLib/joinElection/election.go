package joinElection

import (
	//"fmt"

	messagePasser "../../messagePasser"
	//dns "../dnsService"
)

/**
The package takes care of all the conditions a new node join the network
 */

type JoinElection struct {
	mp *messagePasser.MessagePasser
}

/* Constructor */
func NewJoinElection(mp *messagePasser.MessagePasser) *JoinElection{
	j := JoinElection{mp:mp}
	return j
}

func (j *JoinElection) Start(msg *messagePasser.Message) {
	// Start the election process below
	// j.mp.Send()
}

func (j *JoinElection) Receive(msg *messagePasser.Message) {
	// Deal with the received messages
}