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


func (dhtNode *DHTNode) HandleCreateNewEntryReq(msg *MP.Message) {

	var createNewEntryReq CreateNewEntryRequest
	MP.DecodeData(&createNewEntryReq, msg.Data)
	//var createNewEntryRes CreateNewEntryResponse


	// put entry in this node
	if (dhtNode.isKeyPresentInMyKeyspaceRange(createNewEntryReq.Key)) {
		dhtNode.createEntry(createNewEntryReq.Key, createNewEntryReq.Data)

	// send entry to next node
	} else {
		ip, name := dhtNode.GetNextNodeIPAndNameInRing()
		msg := MP.NewMessage(ip, name, "create_new_entry_req", MP.EncodeData(createNewEntryReq))
		dhtNode.mp.Send(msg)
	}

	// send response back

}

func (dhtNode *DHTNode) HandleCreateNewEntryRes(msg *MP.Message) {

}

func (dhtNode *DHTNode) HandleUpdateEntryReq(msg *MP.Message){

}

func (dhtNode *DHTNode) HandleUpdateEntryRes(msg *MP.Message){

}

func (dhtNode *DHTNode) HandleDeleteEntryReq(msg *MP.Message){

}

func (dhtNode *DHTNode) HandleDeleteEntryRes(msg *MP.Message){

}

func (dhtNode *DHTNode) HandleGetDataReq(msg *MP.Message){

}

func (dhtNode *DHTNode) HandleGetDataRes(msg *MP.Message){

}




