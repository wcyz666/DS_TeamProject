package dht

import (
	messagePasser "../messagePasser"
)

/* Planning to use MD5 to generate the Hash. Hence 128 bit */
const HASH_SIZE = 128

/* Initial DHT version contains details of next and previous nodes */
type DHT struct {
	prefixForwardingTable map[int8][]string
	hashTable             map[string][]MemberShipInfo
	mp 		      *messagePasser.MessagePasser
}

/* Constructor */
func NewDHT(mp *messagePasser.MessagePasser) struct *DHT{
	return DHT{mp: mp}
}

/* TODO: Need to revisit data structures. Temporarily adding superNodeIp*/
type MemberShipInfo struct {
	SuperNodeIp string
}

/* Private Methods */

/* Appends data to entry associated with key "key" in the hash map */
func (dht *DHT) appendData(key string, data MemberShipInfo) {
	entry, isPresent := dht.hashTable[key]
	if false == isPresent {
		newEntry := make([]MemberShipInfo, 0)
		/* add a new entry to hash table */
		newEntry = append(newEntry, data)
		dht.hashTable[key] = newEntry
	} else {
		entry = append(entry, data)
		dht.hashTable[key] = entry
	}
}

/* Replaces entry associated with key "key" with given data */
func (dht *DHT) putData(key string, data []MemberShipInfo) {
	newEntry := make([]MemberShipInfo, len(data))
	copy(newEntry, data)
	dht.hashTable[key] = newEntry
}

/* Retrieves entry with given key */
func (dht *DHT) getData(key string) ([]MemberShipInfo, bool) {
	data, isPresent := dht.hashTable[key]
	return data, isPresent
}

/* delete entry corresponding to given key */
func (dht *DHT) deleteEntry(key string) {
	delete(dht.hashTable, key)
}

/* remove given membership data from entry corresponding to given key */
func (dht *DHT) removeData(key string, data MemberShipInfo) {
	var isMemberShipDataPresent bool = false
	/* Get the index of member info to be removed */
	index := 0
	var value MemberShipInfo

	for index, value = range dht.hashTable[key] {
		if data == value {
			isMemberShipDataPresent = true
			break
		}
	}
	/* Delete entry present at index */
	if isMemberShipDataPresent {
		dht.hashTable[key] = append(dht.hashTable[key][:index], dht.hashTable[key][(index+1):]...)
	}
}

/* Given a key, function will check whether key is within node's key space */
func (dht *DHT) isKeyPresentInKeySpace() bool {
	return true
}

/* Public Methods */

func (dht *DHT) Initialize() {
	dht.hashTable = make(map[string][]MemberShipInfo)
}

func (dht *DHT) Join() {

}

func (dht *DHT) Leave() {

}

func (dht *DHT) Get(streamingGroupID string) ([]MemberShipInfo, bool) {
	if dht.isKeyPresentInKeySpace() {
		return dht.getData(streamingGroupID)
	} else {
		/* TODO fetch data from other node */
		return make([]MemberShipInfo, 0), false
	}
}

func (dht *DHT) Append(streamingGroupID string, data MemberShipInfo) {
	if dht.isKeyPresentInKeySpace() {
		dht.appendData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
}

func (dht *DHT) Put(streamingGroupID string, data []MemberShipInfo) {
	if dht.isKeyPresentInKeySpace() {
		dht.putData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
}

func (dht *DHT) Delete(streamingGroupID string) {
	if dht.isKeyPresentInKeySpace() {
		dht.deleteEntry(streamingGroupID)
	} else {
		/* TODO send update to other node */
	}
}

func (dht *DHT) Remove(streamingGroupID string, data MemberShipInfo) {
	if dht.isKeyPresentInKeySpace() {
		dht.removeData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
}

func (dht *DHT) Refresh(StreamingGroupID string) {

}

/* handler responsible for processing messages received from other nodes
 * and updating the local hash table
 */
func (dht *DHT) HandleRequest() {

}
