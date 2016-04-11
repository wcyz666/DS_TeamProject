package dht

/* Implements functionality related to storing and accessing data in the DHT */

import (
	MP "../messagePasser"
)

/* Private Methods */

/* Creates a new (key,value) pair entry in DHT */
func (dhtNode *DHTNode) createEntry(key string, data MemberShipInfo) (int) {
	var status int
	_, isPresent := dhtNode.hashTable[key]

	status = SUCCESS
	if true == isPresent {
		/* releases the entry for garbage collection */
		dhtNode.hashTable[key] = nil
		status = SUCCESS_ENTRY_OVERWRITTEN
	}

	newEntry := make([]MemberShipInfo, 0)
	/* add a new entry to hash table */
	newEntry = append(newEntry, data)
	dhtNode.hashTable[key] = newEntry

	return status
}

/* deletes entry corresponding to given key */
func (dhtNode *DHTNode) deleteEntry(key string) (int) {
	delete(dhtNode.hashTable, key)
	return SUCCESS
}

/* Appends membership data to existing entry associated with key "key" in the hash map */
func (dhtNode *DHTNode) appendData(key string, data MemberShipInfo) (int){
	entry, isPresent := dhtNode.hashTable[key]
	if false == isPresent {
		return KEY_NOT_PRESENT
	} else {
		entry = append(entry, data)
		dhtNode.hashTable[key] = entry
	}
	return SUCCESS
}

/* remove given membership data from entry corresponding to given key */
func (dhtName *DHTNode) removeData(key string, data MemberShipInfo) (int) {
	var isMemberShipDataPresent bool = false
	/* Get the index of member info to be removed */
	index := 0
	var value MemberShipInfo

	entry, isPresent := dhtName.hashTable[key]
	if (false == isPresent){
		return KEY_NOT_PRESENT
	}

	for index, value = range entry {
		if data == value {
			isMemberShipDataPresent = true
			break
		}
	}
	/* Delete entry present at index */
	if isMemberShipDataPresent {
		dhtName.hashTable[key] = append(dhtName.hashTable[key][:index], dhtName.hashTable[key][(index+1):]...)
	}

	return  SUCCESS
}

/* Retrieves entry with given key */
func (dhtNode *DHTNode) getData(key string) ([]MemberShipInfo, int) {
	status := SUCCESS
	data, isPresent := dhtNode.hashTable[key]
	if (false == isPresent){
		status = KEY_NOT_PRESENT
	}
	return data, status
}


/* handler responsible for processing messages received from other nodes
 * and updating the local hash table
 */
func (dhtNode *DHTNode) HandleDataOperationRequest(msg *MP.Message){

	// decode message into proper structure
	var dataOperationReq DataOperationRequest
	var dataOperationRes DataOperationResponse
	MP.DecodeData(&dataOperationReq, msg.Data)

	var kind string

	if (false == dhtNode.isKeyPresentInMyKeyspaceRange(dataOperationReq.Key)){
		// forward request to node closer to the super node responsible for the key
		nextNode := dhtNode.GetNextNodeToForwardInRing(dataOperationReq.Key)
		msg := MP.NewMessage(nextNode.IpAddress, nextNode.Name, msg.Kind, MP.EncodeData(dataOperationReq))
		dhtNode.mp.Send(msg)
		return
	}

	/* I am responsible for this key. Do the necessary processing */
	msg_type := msg.Kind

	switch msg_type{

	/* handle CreateNewEntry request */
	case "create_new_entry_req":
		dataOperationRes.Status = dhtNode.createEntry(dataOperationReq.Key, dataOperationReq.Data)
		kind = "create_new_entry_res"

	/* handle UpdateEntry request */
	case "update_entry_req":
		// update entry in this node
		if(dataOperationReq.Add == true){
			// add data
			dataOperationRes.Status = dhtNode.appendData(dataOperationReq.Key, dataOperationReq.Data)
		} else if (dataOperationReq.Remove == true){
			// remove data
			dataOperationRes.Status = dhtNode.removeData(dataOperationReq.Key, dataOperationReq.Data)
		}
		kind = "update_entry_res"

	/* handle DeleteEntry request */
	case "delete_entry_req":
		// delete entry in this node
		dataOperationRes.Status = dhtNode.deleteEntry(dataOperationReq.Key)
		kind = "delete_entry_res"

	/*handle GetDate request */
	case "get_data_req":
		// get entry in this node
		dataOperationRes.Data, dataOperationRes.Status = dhtNode.getData(dataOperationReq.Key)
		kind = "get_data_res"

	/* handle unknown kind in message request*/
	default:
		panic("WARNING: Unknown kind in HandleRequest")
	}

	/* Send response to the origin node */
	ip := dataOperationReq.OriginIpAddress
	name := dataOperationReq.OriginName
	responseMsg := MP.NewMessage(ip, name, kind, MP.EncodeData(dataOperationRes))
	dhtNode.mp.Send(responseMsg)
}


func (dhtNode *DHTNode) HandleDataOperationResponse(msg *MP.Message) (int, []MemberShipInfo) {
	var msgDataRes DataOperationResponse
	MP.DecodeData(&msgDataRes, msg.Data)
	return msgDataRes.Status, msgDataRes.Data
}

