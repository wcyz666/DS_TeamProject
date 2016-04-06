package dht

import (
	MP "../messagePasser"
)

/*
 * DHT APIs Implementation.
 */

/* Constructor */
func StartDHTService(mp *MP.MessagePasser) *DHTService {
	var dhtService = DHTService{DhtNode:NewDHTNode(mp)}
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
