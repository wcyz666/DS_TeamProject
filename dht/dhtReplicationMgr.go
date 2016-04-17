package dht

import (
	"fmt"
	"math/big"
	"strconv"
	"time"
	MP "../messagePasser"
)

const (
	REPLICA_CREATION_PENDING   = iota
	REPLICA_CREATION_IN_PROGRESS
	REPLICA_DELETION_PENDING
	REPLICA_DELETION_IN_PROGRESS
	REPLICA_PROCEDURE_NONE
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
	dhtNode.ReplicationState = REPLICA_PROCEDURE_NONE
}

func (dhtNode *DHTNode) CreateRequiredReplicas(failedNode *Node, depth int){
	/* Depth is used when a sequence of nodes fail together, thereby causing replication factor to reduce by more than 1 */
	if (depth > REPLICATION_FACTOR){
		fmt.Println("Since failed node is more than REPLICATION_FACTOR away from us, we don't have a replica of their key space. :( ")
		return
	}
	i := (REPLICATION_FACTOR -1)
	var nodeToCreateReplica *Node
	for (i >= 0 && depth > 0){
//		nodeToCreateReplica = dhtNode.leafTable.NextNodeList[i]
		var createReplicaReq CreateReplicaRequest
		/* Retrieve entries which are less than new node's key and create a map out of it.*/
		prevNodeNumericKey := getBigIntFromString(failedNode.Key)
		var entryKey *big.Int
		createReplicaReq.HashTable = make(map[string][]MemberShipInfo)

		for k,v := range dhtNode.hashTable {
			entryKey = getBigIntFromString(k)
			/* If entry key is within failed node's key space, transfer the data to new node */
			if (false == isKeyPresentInKeyspaceRange(entryKey, prevNodeNumericKey, dhtNode.curNodeNumericKey)){
				createReplicaReq.HashTable[k] = v
			}
		}
		dhtNode.mp.Send(MP.NewMessage(nodeToCreateReplica.IpAddress, nodeToCreateReplica.Name , "dht_create_replica_req",
			MP.EncodeData(createReplicaReq)))
		i--
		depth--
	}

	if (depth > len(dhtNode.leafTable.NextNodeList)){

	}

}

/* After NEIGHBOURHOOD_DISTANCE seconds (allowing for ring to be stabilized by assuming 1 second for message to travel
 * between nodes), create a new replica to bring back currentReplica count to desired value */
func (dhtNode *DHTNode)ScheduleReplicaCreation(prevNode *Node, nodeToCreateReplica *Node){
	if (len(dhtNode.leafTable.NextNodeList) < REPLICATION_FACTOR){
		fmt.Println("Create more super nodes to meet the desired replication factor")
	} else {
		timer1 := time.NewTimer(time.Second * NEIGHBOURHOOD_DISTANCE)
		go func(){
			<-timer1.C
			if (nil == nodeToCreateReplica){
				//nodeToCreateReplica = dhtNode.leafTable.NextNodeList[REPLICATION_FACTOR - 1]
			}

			var createReplicaReq CreateReplicaRequest
			/* Retrieve entries which are less than new node's key and create a map out of it.*/
			prevNodeNumericKey := getBigIntFromString(prevNode.Key)
			var entryKey *big.Int
			createReplicaReq.HashTable = make(map[string][]MemberShipInfo)

			for k,v := range dhtNode.hashTable {
				entryKey = getBigIntFromString(k)
				/* If entry key is within failed node's key space, transfer the data to new node */
				if (false == isKeyPresentInKeyspaceRange(entryKey, prevNodeNumericKey, dhtNode.curNodeNumericKey)){
					createReplicaReq.HashTable[k] = v
				}
			}
			dhtNode.mp.Send(MP.NewMessage(nodeToCreateReplica.IpAddress, nodeToCreateReplica.Name , "dht_create_replica_req",
				MP.EncodeData(createReplicaReq)))
		}()
	}
}

func (dhtNode *DHTNode) HandleCreateReplicaRequest(msg *MP.Message) {
	var createReplicaRequest CreateReplicaRequest
	MP.DecodeData(&createReplicaRequest,msg.Data)
	fmt.Println("[DHT] Create Replica Request received")

	/* Update the local hash table with received values */
	for k,v := range createReplicaRequest.HashTable {
		dhtNode.hashTable[k] = v
	}

	/* send create replica response */
	dhtNode.mp.Send(MP.NewMessage(msg.Src, msg.SrcName , "dht_create_replica_res",
		MP.EncodeData(CreateReplicaResponse{SUCCESS})))
}

func (dhtNode *DHTNode) HandleCreateReplicaResponse(msg *MP.Message) {
	var createReplicaResponse CreateReplicaResponse
	MP.DecodeData(&createReplicaResponse,msg.Data)
	fmt.Println("[DHT] Create Replica Response received with status "+ strconv.Itoa(createReplicaResponse.Status))
	dhtNode.curReplicaCount++
}

