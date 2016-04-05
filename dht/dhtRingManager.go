package dht

/* Implements functionality related to creation and management of underlying Ring structure for Pastry DHT */

import (
	MP "../messagePasser"
	dns "../dnsService"
	config "../config"
	lns "../localNameService"
	"math/big"
)

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

/* Given a key, function will check whether key is within key space managed by this node
 * KeyspaceRange is from (previous node's key + 1) to current node's key
*/
func (dht *DHT) isKeyPresentInMyKeyspaceRange(key string) bool {
	return true
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
	if (false == dht.isKeyPresentInMyKeyspaceRange(joinReq.Key)){
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

func (dht *DHT) Refresh(StreamingGroupID string) {

}

/* handler responsible for processing messages received from other nodes
 * and updating the local hash table
 */
func (dht *DHT) HandleRequest() {

}


