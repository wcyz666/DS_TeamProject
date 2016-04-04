package dht

import (
	MP "../messagePasser"
	dns "../dnsService"
	config "../config"
	lns "../localNameService"
	"math/big"
)

/* Planning to use MD5 to generate the Hash. Hence 128 bit */
const HASH_KEY_SIZE = 128

type Node struct {
	nodeKey string /*TODO we may need it for optimizing the node to be stored in this slot */
	IpAddress string
	port int
}

/* Initial DHT version contains details of next and previous nodes */
type DHT struct {
	/* We interpret Hash value as a hexadecimal stream so each digit is 4 bit long */
	prefixForwardingTable [HASH_KEY_SIZE/4][HASH_KEY_SIZE/4]Node
	hashTable             map[string][]MemberShipInfo
	mp                    *MP.MessagePasser
	nodeKey               string
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
func (dht *DHT) isKeyPresentInKeySpace(key string) bool {
	return true
}

/* Public Methods */

/* Constructor */
func NewDHT(mp *MP.MessagePasser) *DHT {
	var dht = DHT{mp: mp}
	dht.hashTable = make(map[string][]MemberShipInfo)
	/* Use hash of mac address of the super node as the key for partitioning key space */
	dht.nodeKey = lns.GetLocalName()
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
	return ""
}

/*TODO Function responsible for updating leaf table and prefix table based on new information */
func (dht *DHT)updateLeafAndPrefixTablesWithNewNode(newNodeIpAddress string){

}

/*TODO*/
func (dht *DHT)getPredecessorFromLeafTable()(*Node)  {
	return nil
}

/*TODO*/
func (dht *DHT) findSuccessor(key string) (*Node){
	return nil
}

func (dht *DHT) createOrJoinRing(){
	ipAddr := getFirstNonSelfIpAddr()
	if ("" == ipAddr){
		/* No entries exist or your are the only one. This means you are like
		 * Apocalypse, the first mutant. Create a DHT*/

	} else {
		/* Send a message to one of the super nodes requesting to provide successor node's information
		 * based on key provided
		 */
		dht.mp.Send(MP.NewMessage(ipAddr, "", "join_dht_req", MP.EncodeData(JoinRequest{dht.nodeKey})))

	}
}

func (dht *DHT) HandleJoinReq(msg *MP.Message) {
	var joinReq JoinRequest
	MP.DecodeData(&joinReq,msg.Data)
	var joinRes JoinResponse

	/* Send failure status or redirect status if key is not managed by you */
	if (false == dht.isKeyPresentInKeySpace(joinReq.Key)){
		/* Find successor node and send it in the response */
		successor := dht.findSuccessor(joinReq.Key)
		if (nil == successor){
			joinRes.Status = FAILURE
		} else {
			/* indicate to the node about the actual successor */
			joinRes.Status = SUCCESSOR_REDIRECTION
			joinRes.NewSuccessorNode = *successor
		}
		dht.mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "join_dht_res", MP.EncodeData(joinRes)))
		return
	}

	/* Retrieve entries which are less than new node's key and create a map out of it.*/
	nodeKey := new(big.Int)
	_,status := nodeKey.SetString(joinReq.Key, 16)
	if (false == status){
		panic("WARNING: Unable to convert key to a valid value")
	}
	entryKey := new(big.Int)

	for k,v := range dht.hashTable {
		_,status = entryKey.SetString(k,16)
		if (false == status){
			panic("WARNING: Unable to convert key to a valid value")
		}
		/* If entry key is <= new node's key, transfer the data to new node */
		if (entryKey.Cmp(nodeKey) <= 0){
			joinRes.HashTable[k] = v
		}
	}

	/* Send the map in the response to caller */
	joinRes.Status = SUCCESS
	joinRes.Predecessor = *(dht.getPredecessorFromLeafTable())
	dht.mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "join_dht_res", MP.EncodeData(joinRes)))
}

func (dht *DHT) HandleJoinRes(msg *MP.Message) {
	var joinRes JoinResponse
	MP.DecodeData(&joinRes,msg.Data)

	if (joinRes.Status == SUCCESSOR_REDIRECTION){
		/* Send join request to new successor node */
		dht.mp.Send(MP.NewMessage(joinRes.NewSuccessorNode.IpAddress, "", "join_dht_req",
			                             MP.EncodeData(JoinRequest{dht.nodeKey})))

	} else if (joinRes.Status == FAILURE){
		panic ("Join procedure for DHT failed")

	} else {
		/* SUCCESS case */


		/* 1. Add received map to local DHT table */
		dht.hashTable = joinRes.HashTable

		/* 2. Send Join complete to successor */
		dht.mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "join_dht_complete", MP.EncodeData(JoinComplete{SUCCESS,dht.nodeKey})))

		/* 3. Send join notification to predecessor */
		dht.mp.Send(MP.NewMessage(joinRes.Predecessor.IpAddress, "", "join_dht_notify",
			                              MP.EncodeData(JoinNotify{dht.nodeKey})))
	}
}

func (dht *DHT) HandleJoinComplete(msg *MP.Message) {
	var joinComplete JoinComplete
	MP.DecodeData(&joinComplete,msg.Data)

	/* Update routing information to include this new node */
	dht.updateLeafAndPrefixTablesWithNewNode(msg.Src)

	/* Delete entries transferred to new node */
	/* TODO After replication, this needs to be done in farthest replica */
	nodeKey := new(big.Int)
	_,status := nodeKey.SetString(joinComplete.Key, 16)
	if (false == status){
		panic("WARNING: Unable to convert key to a valid value")
	}
	entryKey := new(big.Int)

	for k,_ := range dht.hashTable {
		_,status = entryKey.SetString(k,16)
		if (false == status){
			panic("WARNING: Unable to convert key to a valid value")
		}
		/* If entry key is <= new node's key, remove the entry as it is already transferred to new node */
		if (entryKey.Cmp(nodeKey) <= 0){
			delete(dht.hashTable,k)
		}
	}
}

func (dht *DHT) HandleJoinNotify(msg *MP.Message) {
	/* Update routing information to include this new node */
	dht.updateLeafAndPrefixTablesWithNewNode(msg.Src)
}

func (dht *DHT) Leave(msg *MP.Message) {

}

func (dht *DHT) HandleCreateLSGroupReq(msg *MP.Message) {

}

func (dht *DHT) HandleCreateLSGroupRes(msg *MP.Message) {

}

func (dht *DHT) AddStreamer(msg *MP.Message){

}

func (dht *DHT) RemoveStreamer(msg *MP.Message){

}

func (dht *DHT) DeleteLSGroup(msg *MP.Message){

}


/*
	if (nil == successor){
		successorInfoRes.status = FAILURE
	}

	dht.mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "successor_info_res", MP.EncodeData(successorInfoRes)))
}

func (dht *DHT) HandleSuccessorInfoRes(msg *MP.Message){
	var successorInfoRes SuccessorInfoRes
	MP.DecodeData(&successorInfoRes,msg.Data)

	if (successorInfoRes.status == FAILURE){
		panic("Successor Information Request Failed")
	}

	dht.mp.Send(MP.NewMessage(successorInfoRes.node.IpAddress, "", "join_dht_req", MP.EncodeData(JoinRequest{dht.nodeKey})))
}
*/

func (dht *DHT) Get(streamingGroupID string) ([]MemberShipInfo, bool) {
	if dht.isKeyPresentInKeySpace(streamingGroupID) {
		return dht.getData(streamingGroupID)
	} else {
		/* TODO fetch data from other node */
		return make([]MemberShipInfo, 0), false
	}
}

func (dht *DHT) Append(streamingGroupID string, data MemberShipInfo) {
	if dht.isKeyPresentInKeySpace(streamingGroupID) {
		dht.appendData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
}

func (dht *DHT) Put(streamingGroupID string, data []MemberShipInfo) {
	if dht.isKeyPresentInKeySpace(streamingGroupID) {
		dht.putData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
}

func (dht *DHT) Delete(streamingGroupID string) {
	if dht.isKeyPresentInKeySpace(streamingGroupID) {
		dht.deleteEntry(streamingGroupID)
	} else {
		/* TODO send update to other node */
	}
}

func (dht *DHT) Remove(streamingGroupID string, data MemberShipInfo) {
	if dht.isKeyPresentInKeySpace(streamingGroupID) {
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
