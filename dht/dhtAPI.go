package dht

import (
	MP "../messagePasser"
	"time"
	"fmt"
)

const JOIN_MAX_REATTEMPTS = 5
const (
	NEW_DHT_CREATED = iota
	JOINING_EXISTING_DHT
)

/*
 * DHT APIs Implementation.
 */

/* API to start DHT service on a super node.
 * Params : Message passer reference
 * Return value : DHT service reference on success
 *                nil on failure
 */
func StartDHTService(mp *MP.MessagePasser) *DHTService {
	dhtNode,status := NewDHTNode(mp)
	var dhtService = DHTService{DhtNode: dhtNode}
	mp.AddMappings([]string{"join_dht_res"})

	if (status == NEW_DHT_CREATED){
		return &dhtService
	}

	numOfAttempts := JOIN_MAX_REATTEMPTS

	for {
		select {
			case joinRes := <-mp.Messages["join_dht_res"]:
				 status,successor := dhtService.DhtNode.HandleJoinRes(joinRes)
				 if (JOIN_IN_PROGRESS_RETRY_LATER == status){
					 numOfAttempts--
					 if (numOfAttempts <= 0 ){
						 return nil
					 }
					 /* Another instance of Join is in progress in successor Node
					  * Retry after 2 seconds
					  */
					 timer1 := time.NewTimer(time.Second * 2)
					 go func(){
						 <-timer1.C
						 fmt.Println("Retransmitting Join Request")
						 dhtService.DhtNode.sendJoinReq(successor)
					 }()
				 } else {
					/* Join completed with error or success. Return control to caller */
					 if (status != SUCCESS){
						 return  nil
					 }
					break;
				 }
		}
	}
	return &dhtService
}

func (dht *DHTService) Get(streamingGroupID string) ([]MemberShipInfo, int) {
	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		return dht.DhtNode.getData(streamingGroupID)
	} else {
		/* TODO fetch data from other node */
		return make([]MemberShipInfo, 0), SUCCESS
	}
}

func (dht *DHTService) Create(streamingGroupID string, data MemberShipInfo) (int){
	status:= SUCCESS
	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.createEntry(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
	return status
}

func (dht *DHTService) Delete(streamingGroupID string) (int) {
	status:= SUCCESS
	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.deleteEntry(streamingGroupID)
	} else {
		/* TODO send update to other node */
	}
	return status
}

func (dht *DHTService) Append(streamingGroupID string, data MemberShipInfo) (int) {
	status := SUCCESS
	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status =  dht.DhtNode.appendData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
	return status
}

func (dht *DHTService) Remove(streamingGroupID string, data MemberShipInfo) (int){
	status := SUCCESS
	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.removeData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
	return status
}
