package joinElection

import (
	//"fmt"

	messagePasser "../messagePasser"
	//dns "../dnsService"
)

var mp *messagePasser.MessagePasser

func Start(msg *messagePasser.Message, _mp *messagePasser.MessagePasser) {
	mp = _mp
	// Start the election process below
}

func Receive(msg *messagePasser.Message, _mp *messagePasser.MessagePasser) {
	// Deal with the received messages
}
