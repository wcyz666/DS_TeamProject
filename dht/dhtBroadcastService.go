package dht

import (
	"fmt"
	"time"
	MP "../messagePasser"
)

/* Broadcast related functions */

func (dhtNode *DHTNode) HandleBroadcastMessage(msg *MP.Message) {
	var broadcastMsg BroadcastMessage
	MP.DecodeData(&broadcastMsg,msg.Data)

	fmt.Println("Received broadcast message from " + msg.Src)
	if (dhtNode.IsBroadcastOver(&broadcastMsg)) {
		/* Token returned back to us. Don't forward */
		fmt.Println("Nodes in the ring are ")
		for _, val := range broadcastMsg.TraversedNodesList {
			fmt.Println("IP: "+ val.IpAddress +" Node key: " + val.Key)
		}
	} else {
		/* Add current node details into the list. Currently we use this for debugging
		 * to understand the structure of the ring */
		dhtNode.AppendSelfToBroadcastTrack(&broadcastMsg)
		dhtNode.PassBroadcastMessage(&broadcastMsg, nil)
	}

	//fmt.Println("[DHT] Lead Table contents")
	//fmt.Println("[DHT]	Previous Node List")
	//logNodeList(dhtNode.leafTable.PrevNodeList)
	//fmt.Println("[DHT]	Next Node List")
	//logNodeList(dhtNode.leafTable.NextNodeList)
}

func (dhtNode *DHTNode) GetBroadcastMessage(msg *MP.Message) *BroadcastMessage {
	var broadcastMsg BroadcastMessage
	MP.DecodeData(&broadcastMsg,msg.Data)
	return &broadcastMsg
}

func (dhtNode *DHTNode) IsBroadcastOver(broadcastMsg *BroadcastMessage) bool {
	return broadcastMsg.OriginIpAddress == dhtNode.IpAddress
}

func (dhtNode *DHTNode) AppendSelfToBroadcastTrack(broadcastMsg *BroadcastMessage)  {
	node := Node{dhtNode.IpAddress, dhtNode.NodeName, dhtNode.NodeKey}
	broadcastMsg.TraversedNodesList = append(broadcastMsg.TraversedNodesList,node)
}


func (dhtNode *DHTNode) PassBroadcastMessage(broadcastMsg *BroadcastMessage, payload *MP.Message)  {

	nextNode := dhtNode.leafTable.nextNode
	fmt.Println("Forwarding Broadcast message to " + nextNode.IpAddress)
	if (payload == nil) {
		dhtNode.mp.Send(MP.NewMessage(nextNode.IpAddress, "", "dht_broadcast_msg",
			MP.EncodeData(broadcastMsg)))
	} else {
		broadcastMsg.Payload = MP.EncodeData(payload)
		dhtNode.mp.Send(MP.NewMessage(nextNode.IpAddress, "", "dht_broadcast_msg_" + payload.Kind,
			MP.EncodeData(broadcastMsg)))
	}

}

/*TODO add a parameter to take suitable payload for broadcast. For e.g. we can have type which
  describes about streaming group being newly launched */
func (dhtNode *DHTNode) CreateBroadcastMessage(){
	broadcastMsg := dhtNode.NewBroadcastMessage()

	nextNode := dhtNode.leafTable.nextNode
	if (nextNode == nil){
		if (dhtNode.leafTable.prevNode == nil){
			fmt.Println("DHT with only one node")
		} else {
			panic ("Broken Ring. Next node cannot be nil if there are more than 1 node in DHT")
		}
		return
	}
	fmt.Println("Sending initial broad cast message with node key" + broadcastMsg.OriginName)
	dhtNode.mp.Send(MP.NewMessage(nextNode.IpAddress, "", "dht_broadcast_msg",
		MP.EncodeData(broadcastMsg)))
}

func (dhtNode *DHTNode) NewBroadcastMessage() BroadcastMessage {
	var broadcastMsg BroadcastMessage
	broadcastMsg.OriginIpAddress = dhtNode.IpAddress
	broadcastMsg.OriginName = dhtNode.NodeName
	node:= Node{broadcastMsg.OriginIpAddress,broadcastMsg.OriginName,dhtNode.NodeKey}
	broadcastMsg.TraversedNodesList = append(broadcastMsg.TraversedNodesList,node)

	return broadcastMsg
}

func (dhtNode *DHTNode) PerformPeriodicBroadcast(){
	ticker := time.NewTicker(time.Second * 25)
	go func() {
		for _ = range ticker.C {
			dhtNode.CreateBroadcastMessage()
		}
	}()
}

