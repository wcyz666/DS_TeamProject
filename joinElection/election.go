package joinElection

import (
	"fmt"

	messagePasser "../messagePasser"
	//dns "../dnsService"
)

var mp *messagePasser.MessagePasser

func Start(_mp *messagePasser.MessagePasser, msg *messagePasser.Message){
	mp = _mp
	go receive(msg.SrcName)
	
}


/**
srcName: The name identifier for the client
 */
func receive(srcName string){
	identifier := "join_" + srcName
	channel, ok := mp.Messages[identifier]
	if(ok == false){
		mp.Messages[identifier] = make(chan *messagePasser.Message)
		channel = mp.Messages[identifier]
	}
	msg := <- channel
	fmt.Println(msg)
}
