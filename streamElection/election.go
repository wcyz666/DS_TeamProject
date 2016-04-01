package streamElection

import (
//"fmt"

	messagePasser "../messagePasser"
//dns "../dnsService"
)

/**
The package takes care of all the conditions a new node join the network
 */

type StreamElection struct {
	mp *messagePasser.MessagePasser
}

/* Constructor */
func NewStreamElection(mp *messagePasser.MessagePasser) *StreamElection{
	j := StreamElection{mp:mp}
	return j
}

/* Deal with the election messages of streaming */
func (*StreamElection) Receive(mp *messagePasser.Message) {

}
