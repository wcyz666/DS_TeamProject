package dht

import (
	"fmt"
	"math/big"
	"strconv"
	MP "../messagePasser"
)

/* Replication related functions */

func (dhtNode *DHTNode) HandleDeleteReplicaRequest(msg *MP.Message) {
	var deleteReplicaRequest DeleteReplicaRequest
	var deleteReplicaResponse DeleteReplicaResponse
	MP.DecodeData(&deleteReplicaRequest,msg.Data)
	fmt.Println("[DHT] Delete Replica Request received")

	startNumericKey := getBigIntFromString(deleteReplicaRequest.StartKey)
	endNumericKey :=  getBigIntFromString(deleteReplicaRequest.EndKey)
	var entryKey *big.Int

	deleteReplicaResponse.Status = KEY_NOT_PRESENT
	for k,_ := range dhtNode.hashTable {
		deleteReplicaResponse.Status = SUCCESS
		entryKey = getBigIntFromString(k)
		/* If entry key is in new node's key space, remove the entry as it is already transferred to new node */
		if (false == isKeyPresentInKeyspaceRange(entryKey, startNumericKey, endNumericKey)){
			delete(dhtNode.hashTable,k)
		}
	}
	/* send delete replica response */
	dhtNode.mp.Send(MP.NewMessage(msg.Src, msg.SrcName , "dht_delete_replica_res", MP.EncodeData(deleteReplicaResponse)))
}

func (dhtNode *DHTNode) HandleDeleteReplicaResponse(msg *MP.Message) {
	var deleteReplicaResponse DeleteReplicaResponse
	MP.DecodeData(&deleteReplicaResponse,msg.Data)
	fmt.Println("[DHT] Delete Replica Response received with status "+ strconv.Itoa(deleteReplicaResponse.Status))
	/* TODO DO we need to synchronize access to curReplicaCount since both CreateReplicaResponse and DeleteReplicaResponse can update it */
	dhtNode.curReplicaCount--
}

func (dhtNode *DHTNode) StartReplicaSync(){

	if (dhtNode.AmITheOnlyNodeInDHT()){
		/* Replication. Really ? huh */
		return
	}

	var replicaSyncMsg ReplicaSyncMessage
	/* Since I have a copy of data too, need to create only (Replica Factor -1) replicas */
	replicaSyncMsg.ResidualHopCount = (REPLICATION_FACTOR - 1)
	replicaSyncMsg.OriginIpAddress = dhtNode.IpAddress
	replicaSyncMsg.OriginName = dhtNode.NodeName
	replicaSyncMsg.HashTable = make(map[string][]MemberShipInfo)

	var entryKey *big.Int
	for k,v := range dhtNode.hashTable {
		entryKey = getBigIntFromString(k)
		/* If entry key is within my key space, add it to the hash table */
		if (false == isKeyPresentInKeyspaceRange(entryKey, dhtNode.prevNodeNumericKey, dhtNode.curNodeNumericKey)){
			replicaSyncMsg.HashTable[k] = v
		}
	}

	nodeToForward := dhtNode.leafTable.nextNode
	dhtNode.mp.Send(MP.NewMessage(nodeToForward.IpAddress, nodeToForward.Name , "dht_replica_sync",
		MP.EncodeData(replicaSyncMsg)))
}

func (dhtNode *DHTNode) HandleReplicaSyncMsg(msg *MP.Message){
	fmt.Println("[DHT] Handle Replica Synchronization message ")
	var replicaSyncMsg ReplicaSyncMessage
	MP.DecodeData(&replicaSyncMsg,msg.Data)

	//fmt.Println("HandleReplicaSyncMsg message from "+ msg.Src + "with hop "+ strconv.Itoa(discoveryMsg.ResidualHopCount))

	if (replicaSyncMsg.OriginIpAddress == dhtNode.IpAddress){
		/* If hop count is zero, it there are as many replicas as required by replication factor.
		 * Otherwise, residual hop count will indicate how much we are are short of replication factor */
		dhtNode.curReplicaCount = REPLICATION_FACTOR - replicaSyncMsg.ResidualHopCount

	} else{

		/* Update the local hash table with received values */
		for k,v := range replicaSyncMsg.HashTable {
			dhtNode.hashTable[k] = v
		}

		replicaSyncMsg.ResidualHopCount--
		if (replicaSyncMsg.ResidualHopCount == 0){
			//fmt.Println("Forwarded message to origin: "+ discoveryMsg.OriginIpAddress)
			dhtNode.mp.Send(MP.NewMessage(replicaSyncMsg.OriginIpAddress, replicaSyncMsg.OriginName,
				"dht_replica_sync", MP.EncodeData(replicaSyncMsg)))
		} else {
			nodeToForward := dhtNode.leafTable.nextNode
			//fmt.Println("Forwarded message to "+ nodeToForward.IpAddress)
			//logNodeList(discoveryMsg.NodeList)
			dhtNode.mp.Send(MP.NewMessage(nodeToForward.IpAddress, nodeToForward.Name,
				"dht_replica_sync", MP.EncodeData(replicaSyncMsg)))
		}
	}
}

