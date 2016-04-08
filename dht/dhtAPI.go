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

const (
	DHT_API_SUCCESS = iota
	DHT_API_FAILURE_MAX_ATTEMPTS_REACHED
	DHT_API_FAILURE
)

/*
 * DHT APIs Implementation.
 */

/* API to start DHT service on a super node.
 * Params : Message passer reference
 * Return value : DHT service reference on success
 *                nil on failure
 */
func NewDHTService(mp *MP.MessagePasser) *DHTService {
	dhtNode := NewDHTNode(mp)
	var dhtService = DHTService{DhtNode: dhtNode}
	mp.AddMappings([]string{"join_dht_res"})
	return &dhtService
}

func (dhtService *DHTService)Start() int{
	status := dhtService.DhtNode.CreateOrJoinRing()
	if (status == NEW_DHT_CREATED){
		return DHT_API_SUCCESS
	}

	numOfAttempts := JOIN_MAX_REATTEMPTS

	for {
		select {
		case joinRes := <-dhtService.DhtNode.mp.Messages["join_dht_res"]:
			status,successor := dhtService.DhtNode.HandleJoinRes(joinRes)
			if (JOIN_IN_PROGRESS_RETRY_LATER == status){
				numOfAttempts--
				if (numOfAttempts <= 0 ){
					return DHT_API_FAILURE_MAX_ATTEMPTS_REACHED
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
					return  DHT_API_FAILURE
				}
				break;
			}
		}
	}
}

func (dht *DHTService) Get(streamingGroupID string) ([]MemberShipInfo, int) {
	var dataOperationReq DataOperationRequest

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		return dht.DhtNode.getData(streamingGroupID)
	} else {
		dataOperationReq.Key = streamingGroupID
		ip, name := dht.DhtNode.GetNextNodeIPAndNameInRing()
		msg := MP.NewMessage(ip, name, "get_data_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <-dht.DhtNode.mp.Messages["get_data_res"]:
			status, data := dht.DhtNode.HandleResponse(getDataResMsg)
			return data, status
		}

		return make([]MemberShipInfo, 0), FAILURE
	}
}

func (dht *DHTService) Create(streamingGroupID string, data MemberShipInfo) (int){
	status:= SUCCESS
	var dataOperationReq DataOperationRequest

	// add entry to this node
	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.createEntry(streamingGroupID, data)

	// send the entry to the next node
	} else {
		dataOperationReq.Key = streamingGroupID
		dataOperationReq.Data = data
		ip, name := dht.DhtNode.GetNextNodeIPAndNameInRing()
		msg := MP.NewMessage(ip, name, "create_new_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["create_new_entry_res"]:
			status,_ = dht.DhtNode.HandleResponse(getDataResMsg)
		}
	}
	return status
}

func (dht *DHTService) Delete(streamingGroupID string) (int) {
	status:= SUCCESS
	var dataOperationReq DataOperationRequest

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.deleteEntry(streamingGroupID)
	} else {
		dataOperationReq.Key = streamingGroupID
		ip, name := dht.DhtNode.GetNextNodeIPAndNameInRing()
		msg := MP.NewMessage(ip, name, "delete_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["delete_entry_res"]:
			status,_ = dht.DhtNode.HandleResponse(getDataResMsg)
		}
	}
	return status
}

func (dht *DHTService) Append(streamingGroupID string, data MemberShipInfo) (int) {
	status := SUCCESS
	var dataOperationReq DataOperationRequest

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status =  dht.DhtNode.appendData(streamingGroupID, data)
	} else {
		dataOperationReq.Key = streamingGroupID
		dataOperationReq.Add = true
		dataOperationReq.Remove = false
		ip, name := dht.DhtNode.GetNextNodeIPAndNameInRing()
		msg := MP.NewMessage(ip, name, "update_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["update_entry_res"]:
			status,_ = dht.DhtNode.HandleResponse(getDataResMsg)
		}
	}
	return status
}

func (dht *DHTService) Remove(streamingGroupID string, data MemberShipInfo) (int){
	status := SUCCESS
	var dataOperationReq DataOperationRequest

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.removeData(streamingGroupID, data)
	} else {
		dataOperationReq.Key = streamingGroupID
		dataOperationReq.Add = false
		dataOperationReq.Remove = true
		ip, name := dht.DhtNode.GetNextNodeIPAndNameInRing()
		msg := MP.NewMessage(ip, name, "update_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["update_entry_res"]:
			status,_ = dht.DhtNode.HandleResponse(getDataResMsg)
		}
	}
	return status
}

