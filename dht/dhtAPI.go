package dht

import (
	MP "../messagePasser"
	"time"
	"fmt"
	"math/big"
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
	mp.AddMappings([]string{"join_dht_res", "join_dht_conn_failed"})
	/*TODO check if adding a global handler for receving data operation response is fine */
	mp.AddMappings([]string{"dht_data_operation_res", "get_data_res", "delete_entry_res", "create_new_entry_res",
							"update_entry_res"})
	return &dhtService
}

func (dhtService *DHTService)Start() int{
	status := dhtService.DhtNode.CreateOrJoinRing()
	if (status == NEW_DHT_CREATED){
		/* Unit testing the ring*/
		//dhtService.DhtNode.PerformPeriodicBroadcast()
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
				return DHT_API_SUCCESS;
			}
		case msg := <-dhtService.DhtNode.mp.Messages["join_dht_conn_failed"]:
			var failClientInfo MP.FailClientInfo
			MP.DecodeData(&failClientInfo,msg.Data)
			fmt.Println("Join Attempt with " + failClientInfo.IP + " failed. Removing the entry and moving to next IP")
			/* Remove failed node from DNS */
			dhtService.DhtNode.RemoveFailedSuperNode(failClientInfo.IP)
			status := dhtService.DhtNode.CreateOrJoinRing()
			if (status == NEW_DHT_CREATED){
				return DHT_API_SUCCESS
			}
		}
	}
}

func IsStreamingGroupIdValid(streamingGroupID string) (bool){
	numericKey := new(big.Int)
	_,status := numericKey.SetString(streamingGroupID, 16)
	return status
}

func (dht *DHTService) Get(streamingGroupID string) ([]MemberShipInfo, int) {
	if (false == IsStreamingGroupIdValid(streamingGroupID)) {
		return make([]MemberShipInfo, 0),INVALID_INPUT_PARAMS
	}

	var  dataOperationReq = DataOperationRequest{OriginIpAddress: dht.DhtNode.IpAddress,
		                                         OriginName : dht.DhtNode.NodeName}

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		return dht.DhtNode.getData(streamingGroupID)
	} else {
		dataOperationReq.Key = streamingGroupID
		nextNode := dht.DhtNode.GetNextNodeToForwardInRing(streamingGroupID)
		msg := MP.NewMessage(nextNode.IpAddress, nextNode.Name, "get_data_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <-dht.DhtNode.mp.Messages["get_data_res"]:
			status, data := dht.DhtNode.HandleDataOperationResponse(getDataResMsg)
			return data, status
		}
	}
}

func (dht *DHTService) Create(streamingGroupID string, data MemberShipInfo) (int){
	if (false == IsStreamingGroupIdValid(streamingGroupID)) {
		return INVALID_INPUT_PARAMS
	}

	status:= SUCCESS
	var  dataOperationReq = DataOperationRequest{OriginIpAddress: dht.DhtNode.IpAddress,
		                                         OriginName : dht.DhtNode.NodeName}

	// add entry to this node
	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.createEntry(streamingGroupID, data)

	// send the entry to the next node
	} else {
		dataOperationReq.Key = streamingGroupID
		dataOperationReq.Data = data
		nextNode := dht.DhtNode.GetNextNodeToForwardInRing(streamingGroupID)
		msg := MP.NewMessage(nextNode.IpAddress, nextNode.Name, "create_new_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["create_new_entry_res"]:
			status,_ = dht.DhtNode.HandleDataOperationResponse(getDataResMsg)
		}
	}
	return status
}

func (dht *DHTService) Delete(streamingGroupID string) (int) {
	if (false == IsStreamingGroupIdValid(streamingGroupID)) {
		return INVALID_INPUT_PARAMS
	}

	status:= SUCCESS
	var dataOperationReq = DataOperationRequest{OriginIpAddress: dht.DhtNode.IpAddress,
		                                        OriginName : dht.DhtNode.NodeName}

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.deleteEntry(streamingGroupID)
	} else {
		dataOperationReq.Key = streamingGroupID
		nextNode := dht.DhtNode.GetNextNodeToForwardInRing(streamingGroupID)
		msg := MP.NewMessage(nextNode.IpAddress, nextNode.Name, "delete_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["delete_entry_res"]:
			status,_ = dht.DhtNode.HandleDataOperationResponse(getDataResMsg)
		}
	}
	return status
}

func (dht *DHTService) Append(streamingGroupID string, data MemberShipInfo) (int) {
	if (false == IsStreamingGroupIdValid(streamingGroupID)) {
		return INVALID_INPUT_PARAMS
	}
	status := SUCCESS
	var  dataOperationReq = DataOperationRequest{OriginIpAddress: dht.DhtNode.IpAddress,
		                                         OriginName : dht.DhtNode.NodeName}

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status =  dht.DhtNode.appendData(streamingGroupID, data)
	} else {
		dataOperationReq.Key = streamingGroupID
		dataOperationReq.Add = true
		dataOperationReq.Remove = false
		dataOperationReq.Data = data
		nextNode := dht.DhtNode.GetNextNodeToForwardInRing(streamingGroupID)
		msg := MP.NewMessage(nextNode.IpAddress, nextNode.Name, "update_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["update_entry_res"]:
			status,_ = dht.DhtNode.HandleDataOperationResponse(getDataResMsg)
		}
	}
	return status
}

func (dht *DHTService) Remove(streamingGroupID string, data MemberShipInfo) (int){
	if (false == IsStreamingGroupIdValid(streamingGroupID)) {
		return INVALID_INPUT_PARAMS
	}
	status := SUCCESS
	var  dataOperationReq = DataOperationRequest{OriginIpAddress: dht.DhtNode.IpAddress,
		                                         OriginName : dht.DhtNode.NodeName}

	if dht.DhtNode.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.DhtNode.removeData(streamingGroupID, data)
	} else {
		dataOperationReq.Key = streamingGroupID
		dataOperationReq.Add = false
		dataOperationReq.Remove = true
		dataOperationReq.Data = data
		nextNode := dht.DhtNode.GetNextNodeToForwardInRing(streamingGroupID)
		msg := MP.NewMessage(nextNode.IpAddress, nextNode.Name, "update_entry_req", MP.EncodeData(dataOperationReq))
		dht.DhtNode.mp.Send(msg)

		select {
		case getDataResMsg := <- dht.DhtNode.mp.Messages["update_entry_res"]:
			status,_ = dht.DhtNode.HandleDataOperationResponse(getDataResMsg)
		}
	}
	return status
}

func (dht *DHTService)  TriggerBroadcastMessage(){
	dht.DhtNode.CreateBroadcastMessage()
}
