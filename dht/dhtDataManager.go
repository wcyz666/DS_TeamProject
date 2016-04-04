package dht

/* Implements functionality related to storing and accessing data in the DHT */

import (
	MP "../messagePasser"
)

/* Private Methods */

/* Creates a new (key,value) pair entry in DHT */
func (dht *DHT) createEntry(key string, data MemberShipInfo) (int) {
	var status int
	_, isPresent := dht.hashTable[key]

	status = SUCCESS
	if true == isPresent {
		/* releases the entry for garbage collection */
		dht.hashTable[key] = nil
		status = SUCCESS_ENTRY_OVERWRITTEN
	}

	newEntry := make([]MemberShipInfo, 0)
	/* add a new entry to hash table */
	newEntry = append(newEntry, data)
	dht.hashTable[key] = newEntry

	return status
}

/* deletes entry corresponding to given key */
func (dht *DHT) deleteEntry(key string) (int) {
	delete(dht.hashTable, key)
	return SUCCESS
}

/* Appends membership data to existing entry associated with key "key" in the hash map */
func (dht *DHT) appendData(key string, data MemberShipInfo) (int){
	entry, isPresent := dht.hashTable[key]
	if false == isPresent {
		return KEY_NOT_PRESENT
	} else {
		entry = append(entry, data)
		dht.hashTable[key] = entry
	}
	return SUCCESS
}

/* remove given membership data from entry corresponding to given key */
func (dht *DHT) removeData(key string, data MemberShipInfo) (int) {
	var isMemberShipDataPresent bool = false
	/* Get the index of member info to be removed */
	index := 0
	var value MemberShipInfo

	entry, isPresent := dht.hashTable[key]
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
		dht.hashTable[key] = append(dht.hashTable[key][:index], dht.hashTable[key][(index+1):]...)
	}

	return  SUCCESS
}

/* Retrieves entry with given key */
func (dht *DHT) getData(key string) ([]MemberShipInfo, int) {
	status := SUCCESS
	data, isPresent := dht.hashTable[key]
	if (false == isPresent){
		status = KEY_NOT_PRESENT
	}
	return data, status
}


func (dht *DHT) HandleCreateNewEntryReq(msg *MP.Message) {

}

func (dht *DHT) HandleCreateNewEntryRes(msg *MP.Message) {

}

func (dht *DHT) HandleUpdateEntryReq(msg *MP.Message){

}

func (dht *DHT) HandleUpdateEntryRes(msg *MP.Message){

}

func (dht *DHT) HandleDeleteEntryReq(msg *MP.Message){

}

func (dht *DHT) HandleDeleteEntryRes(msg *MP.Message){

}

func (dht *DHT) HandleGetDataReq(msg *MP.Message){

}

func (dht *DHT) HandleGetDataRes(msg *MP.Message){

}


