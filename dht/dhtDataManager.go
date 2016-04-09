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
	var msgDataReq DataOperationRequest
	var msgDataRes DataOperationResponse
	MP.DecodeData(&msgDataReq, msg.Data)


	msg_type := msg.Kind

	switch msg_type{

	/* handle CreateNewEntry request */
	case "create_entry_req":

		// create entry in this node
		if (dhtNode.isKeyPresentInMyKeyspaceRange(msgDataReq.Key)) {
			msgDataRes.Status = dhtNode.createEntry(msgDataReq.Key, msgDataReq.Data)

		// forward entry to next node
		} else {
			ip, name := dhtNode.GetNextNodeIPAndNameInRing()
			msg := MP.NewMessage(ip, name, "create_new_entry_req", MP.EncodeData(msgDataReq))
			dhtNode.mp.Send(msg)
		}

		// send response if entry was successful
		if (msgDataRes.Status == SUCCESS) {
			ip := msgDataReq.OriginIpAddress
			name := msgDataReq.OriginName
			msg := MP.NewMessage(ip, name, "create_new_entry_res", MP.EncodeData(msgDataRes))
			dhtNode.mp.Send(msg)
		}


	/* handle UpdateEntry request */
	case "update_entry_req":

		// update entry in this node
		if (dhtNode.isKeyPresentInMyKeyspaceRange(msgDataReq.Key)) {

			// add data
			if(msgDataReq.Add == true){
				msgDataRes.Status = dhtNode.appendData(msgDataReq.Key, msgDataReq.Data)

			// remove data
			} else if (msgDataReq.Remove == true){
				dhtNode.removeData(msgDataReq.Key, msgDataReq.Data)
			}

		// forward entry to next node
		} else {
			ip, name := dhtNode.GetNextNodeIPAndNameInRing()
			msg := MP.NewMessage(ip, name, "update_entry_req", MP.EncodeData(msgDataReq))
			dhtNode.mp.Send(msg)
		}

		// send response if entry was successful
		if (msgDataRes.Status == SUCCESS){
			ip := msgDataReq.OriginIpAddress
			name := msgDataReq.OriginName
			msg := MP.NewMessage(ip, name, "update_entry_res", MP.EncodeData(msgDataRes))
			dhtNode.mp.Send(msg)
		}


	/* handle DeleteEntry request */
	case "delete_entry_req":

		// delete entry in this node
		if (dhtNode.isKeyPresentInMyKeyspaceRange(msgDataReq.Key)){
			msgDataRes.Status = dhtNode.deleteEntry(msgDataReq.Key)

		// send entry to next node
		} else {
			ip := msgDataReq.OriginIpAddress
			name := msgDataReq.OriginName
			msg := MP.NewMessage(ip, name, "delete_entry_req", MP.EncodeData(msgDataRes))
			dhtNode.mp.Send(msg)
		}

		// send response if deletion was successful
		if (msgDataRes.Status == SUCCESS){
			ip := msgDataReq.OriginIpAddress
			name := msgDataReq.OriginName
			msg := MP.NewMessage(ip, name, "delete_entry_res", MP.EncodeData(msgDataRes))
			dhtNode.mp.Send(msg)
		}


	/*handle GetDate request */
	case "get_data_req":

		// get entry in this node
		if (dhtNode.isKeyPresentInMyKeyspaceRange(msgDataReq.Key)){
			msgDataRes.Data, msgDataRes.Status = dhtNode.getData(msgDataReq.Key)

		// send entry to next node
		} else {
			ip := msgDataReq.OriginIpAddress
			name := msgDataReq.OriginName
			msg := MP.NewMessage(ip, name, "get_data_req", MP.EncodeData(msgDataRes))
			dhtNode.mp.Send(msg)
		}

		// send response if retrieval was successful
		if (msgDataRes.Status == SUCCESS){
			ip := msgDataReq.OriginIpAddress
			name := msgDataReq.OriginName
			msg := MP.NewMessage(ip, name, "get_data_res", MP.EncodeData(msgDataRes))
			dhtNode.mp.Send(msg)
		}

	/* handle unknown kind in message request*/
	default:
		panic("WARNING: Unknown kind in HandleRequest")
	}
}


func (dhtNode *DHTNode) HandleDataOperationResponse(msg *MP.Message) (int, []MemberShipInfo) {

	msg_type := msg.Kind

	var msgDataRes DataOperationResponse
	MP.DecodeData(&msgDataRes, msg.Data)

	switch msg_type{

	/* handle CreateEntry response */
	case "create_entry_res":


	/* handle UpdateEntry */
	case "update_entry_res":


	/* handle DeleteEntry */
	case "delete_entry_res":


	/* handle GetDate */
	case "get_data_res":


	default:
		panic("WARNING: Unknown kind in HandleResponse")
	}
	return FAILURE, make([]MemberShipInfo, 0)
}

