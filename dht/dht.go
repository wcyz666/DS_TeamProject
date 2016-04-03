package dht

import (
	MP "../messagePasser"
	dns "../dnsService"
	config "../config"
	lns "../localNameService"
)

/* Planning to use MD5 to generate the Hash. Hence 128 bit */
const HASH_KEY_SIZE = 128

type NodeEntry struct {
	nodeKey string /*TODO we may need it for optimizing the node to be stored in this slot */
	IpAddress string
	port int
}

/* Initial DHT version contains details of next and previous nodes */
type DHT struct {
	/* We interpret Hash value as a hexadecimal stream so each digit is 4 bit long */
	prefixForwardingTable [HASH_KEY_SIZE/4][HASH_KEY_SIZE/4]NodeEntry
	hashTable             map[string][]MemberShipInfo
	mp                    *MP.MessagePasser
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

/* Constructor */
func NewDHT(mp *MP.MessagePasser) *DHT {
	var dht = DHT{mp: mp}
	dht.hashTable = make(map[string][]MemberShipInfo)
	dht.createOrJoinRing()
	return &dht
}

func getFirstNonSelfIpAddr() (string){
	curAddrList := dns.GetAddr(config.BootstrapDomainName)
	extIP, _ := dns.ExternalIP()

	for _, ipAddr := range curAddrList {
		if ipAddr == extIP {
			continue
		} else{
			return ipAddr
		}
	}
	return nil
}

func (dht *DHT) findSuccessor(key string) (*NodeEntry){
	return nil
}

func (dht *DHT) createOrJoinRing(){
	ipAddr := getFirstNonSelfIpAddr()
	if (nil == ipAddr){
		/* No entries exist or your are the only one. This means you are like
		 * Apocalypse, the first mutant. Create a DHT*/

	} else {
		/* Use hash of mac address of the super node as the key for partitioning key space */
		key := lns.GetLocalName()

		/* Send a message to one of the super nodes requesting to provide successor node's information
		 * based on key provided
		 */
		kind := "successor_info_req"
		dht.mp.Send(MP.NewMessage(ipAddr, kind, key))
	}
}

func (dht *DHT) Join(msg *MP.Message) {

	/* give */
}

func (dht *DHT) Leave(msg *MP.Message) {

}

func (dht *DHT) CreateLSGroup(msg *MP.Message) {

}

func (dht *DHT) AddStreamer(msg *MP.Message){

}

func (dht *DHT) RemoveStreamer(msg *MP.Message){

}

func (dht *DHT) DeleteLSGroup(msg *MP.Message){

}

func (dht *DHT) HandleSuccessorInfoReq(msg *MP.Message){
	key := msg.Data
	/* TODO Do we need to trigger findSuccessor in a separate thread ? */
	successor := dht.findSuccessor(key)

	if (nil == successor){
		panic("Successor cannot be null")
	}

	kind := "successor_info_req"
	dht.mp.Send(MP.NewMessage(msg.Src, kind, key))

}

func (dht *DHT) HandleSuccessorInfoRes(msg *MP.Message){

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
