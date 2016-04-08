package dht

/* Implements functionality related to creation and management of underlying Ring structure for Pastry DHT */

import (
	MP "../messagePasser"
	dns "../dnsService"
	config "../config"
	lns "../localNameService"
	"math/big"
	"fmt"
	"time"
)

const (
	TRAVERSE_CLOCK_WISE   = iota // Traverse in the direction of next node
	TRAVERSE_ANTI_CLOCK_WISE // Traverse in direction of prev node
)

/* Constructor */
func NewDHTNode(mp *MP.MessagePasser) (*DHTNode) {
	var dhtNode = DHTNode{mp: mp}
	dhtNode.hashTable = make(map[string][]MemberShipInfo)
	/* Use hash of mac address of the super node as the key for partitioning key space */
	dhtNode.nodeKey = lns.GetLocalName()
	dhtNode.ipAddress, _ = dns.ExternalIP()
	dhtNode.curNodeNumericKey =  getBigIntFromString(dhtNode.nodeKey)
	return &dhtNode
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


/* Computes the direction in which we need to make the traversal. First argument is
 * curNodeKey and second argument is the key towards which we need to traversal */
func computeTraversalDirection(curNodeKey *big.Int, newKey *big.Int ) int{
	status := TRAVERSE_CLOCK_WISE

	zero := getBigIntFromString("0")
	one  := getBigIntFromString("1")
	maxKey := getBigIntFromString(MAX_KEY)

	/* compute the distance in clock wise and anti clock wise direction and travel in the direction
	*  which is shortest */
	k := new(big.Int)
	/* k = curNodeKey - newKey */
	k.Sub(curNodeKey,newKey)

	clockWiseDistance :=  new(big.Int)
	antiClockWiseDistance := new(big.Int)

	if (k.Cmp(zero) > 0) {
		antiClockWiseDistance = k
		clockWiseDistance.Sub(maxKey,antiClockWiseDistance)
		clockWiseDistance.Add(clockWiseDistance,one)
	} else {
		clockWiseDistance.Sub(zero,k)
		antiClockWiseDistance.Sub(maxKey,clockWiseDistance)
		antiClockWiseDistance.Add(antiClockWiseDistance,one)
	}

	if (clockWiseDistance.Cmp(antiClockWiseDistance) > 0){
		status = TRAVERSE_ANTI_CLOCK_WISE
	} else {
		status = TRAVERSE_CLOCK_WISE
	}
	return status
}

func getBigIntFromString(key string) *big.Int{
	numericKey := new(big.Int)
	_,status := numericKey.SetString(key, 16)
	if (false == status){
		panic("WARNING: Unable to convert newNodeKey to a valid value")
	}
	return numericKey
}

/* Given a key, function will check whether key is within key space managed by this node
 * KeyspaceRange is from (previous node's key + 1) to current node's key
*/
func (dhtNode *DHTNode) isKeyPresentInMyKeyspaceRange(key string) bool {
	numericKey := getBigIntFromString(key)
	if  ((dhtNode.leafTable.nextNode == nil) &&(dhtNode.leafTable.prevNode == nil)){
		return true
	}

	zero := getBigIntFromString("0")
	maxKey := getBigIntFromString(MAX_KEY)

	/* If curNodeKey > prevNodeKey, check if new key in (prevNodeKey, curNodeKey]
	 * If not, check if new key is in (prevNodeKey, Maxkey) or [0, curNodeKey]
	*/
	if (dhtNode.curNodeNumericKey.Cmp(dhtNode.prevNodeNumericKey) > 0){
		if ((numericKey.Cmp(dhtNode.prevNodeNumericKey) > 0) &&
		    (numericKey.Cmp(dhtNode.curNodeNumericKey) <= 0)){
			return true
		} else {
			return false
		}
	} else {
		if (((numericKey.Cmp(dhtNode.prevNodeNumericKey) > 0) && (numericKey.Cmp(maxKey) <=0)) ||
			((numericKey.Cmp(zero)>=0) && (numericKey.Cmp(dhtNode.curNodeNumericKey) <=0))) {
			return true
		} else {
			return false
		}
	}
}

/*TODO Function responsible for updating leaf table and prefix table based on new information */
func (dhtNode *DHTNode)updateLeafAndPrefixTablesWithNewNode(newNodeIpAddress string, newNodeName string,
																	 newNodeKey string, isPrevNode bool){

	/* Update prev and next node information for now */
	newNodeNumericKey := getBigIntFromString(newNodeKey)

	var node = Node{newNodeIpAddress,newNodeName}
	if (true == isPrevNode){
		dhtNode.leafTable.prevNode = &node
		dhtNode.prevNodeNumericKey = newNodeNumericKey
	} else{
		dhtNode.leafTable.nextNode = &node
	}
}

func (dhtNode *DHTNode)getPredecessorFromLeafTable()(*Node)  {
	return dhtNode.leafTable.prevNode
}

/*TODO currently we are implementing a simple find successor scheme */
func (dhtNode *DHTNode) findSuccessor(key string) (*Node){
	if (computeTraversalDirection(dhtNode.curNodeNumericKey,getBigIntFromString(key)) == TRAVERSE_ANTI_CLOCK_WISE){
		return dhtNode.leafTable.prevNode
	} else {
		return dhtNode.leafTable.nextNode
	}
}

func (dhtNode *DHTNode) CreateOrJoinRing()int{
	ipAddr := getFirstNonSelfIpAddr()
	if ("" == ipAddr){
		/* No entries exist or your are the only one. This means you are like
		 * Apocalypse, the first mutant. Create a DHT*/
		fmt.Println("[DHT]	Creating New DHT")
		dhtNode.leafTable.nextNode = nil
		dhtNode.leafTable.prevNode = nil
		return NEW_DHT_CREATED
	} else {
		/* Send a message to one of the super nodes requesting to provide successor node's information
		 * based on key provided
		 */
		fmt.Println("[DHT]	Joining Existing DHT. Sending Request to " + ipAddr)
		ip,name := dhtNode.mp.GetNodeIpAndName()
		dhtNode.mp.Send(MP.NewMessage(ipAddr, "", "join_dht_req", MP.EncodeData(JoinRequest{dhtNode.nodeKey,ip,name})))
		return JOINING_EXISTING_DHT
	}
}

func (dhtNode *DHTNode) sendJoinReq(node *Node){
	ip,name := dhtNode.mp.GetNodeIpAndName()
	dhtNode.mp.Send(MP.NewMessage(node.IpAddress, "", "join_dht_req", MP.EncodeData(JoinRequest{dhtNode.nodeKey,ip,name})))
}

func (dhtNode *DHTNode) AmITheOnlyNodeInDHT()(bool){
	if ((nil == dhtNode.leafTable.prevNode) &&
	    (nil == dhtNode.leafTable.nextNode)){
		return true
	}
	return false
}

func (dhtNode *DHTNode) HandleJoinReq(msg *MP.Message) {
	var joinReq JoinRequest
	MP.DecodeData(&joinReq,msg.Data)
	var joinRes JoinResponse
	fmt.Println("[DHT] HandleJoinReq")
        
	if (true == dhtNode.AmITheOnlyNodeInDHT()){
		/* Me apocolyse, got my first disciple. Join request received for a DHT ring of one node */
		/* Add new node as both prev and next node of current node */
		node := Node{joinReq.OriginIpAddress,joinReq.OriginName}
		joinRes.Predecessor = node
		fmt.Println("[DHT] Adding my first disciple (i.e.) second node in DHT.")
	} else {
		/* Forward the message if key is not managed by you */
		if (false == dhtNode.isKeyPresentInMyKeyspaceRange(joinReq.Key)){
			/* Find successor node and send it in the response */
			successor := dhtNode.findSuccessor(joinReq.Key)
			if (nil == successor){
				joinRes.Status = FAILURE
				/* Send failure message to Join Request originator */
				dhtNode.mp.Send(MP.NewMessage(joinReq.OriginIpAddress,
					joinReq.OriginName, "join_dht_res", MP.EncodeData(joinRes)))
			} else {
				/* Forward the message towards successor */
				dhtNode.mp.Send(MP.NewMessage(successor.IpAddress, "", "join_dht_req", MP.EncodeData(joinReq)))
			}
			return
		}
		joinRes.Predecessor = *(dhtNode.getPredecessorFromLeafTable())
	}


	if (true == dhtNode.isRingUpdateInProgress){
		fmt.Println("[DHT] Join In Progress. Retry later")
		joinRes.Status = JOIN_IN_PROGRESS_RETRY_LATER
		dhtNode.mp.Send(MP.NewMessage(joinReq.OriginIpAddress,
			joinReq.OriginName, "join_dht_res", MP.EncodeData(joinRes)))
		return
	}

	/* Since we are transferring a portion of our hashtable to new node and the process is still in progress
	 * set this flag */
	dhtNode.isRingUpdateInProgress = true
	/* Retrieve entries which are less than new node's key and create a map out of it.*/
	nodeKey := getBigIntFromString(joinReq.Key)
	var entryKey *big.Int

	for k,v := range dhtNode.hashTable {
		entryKey = getBigIntFromString(k)
		/* If entry key is <= new node's key, transfer the data to new node */
		if (entryKey.Cmp(nodeKey) <= 0){
			joinRes.HashTable[k] = v
		}
	}

	/* Send the map in the response to Join Request originator */
	joinRes.Status = SUCCESS

	fmt.Println("[DHT] Sending Successful Join Response to " + joinReq.OriginIpAddress)
	dhtNode.mp.Send(MP.NewMessage(joinReq.OriginIpAddress, "" , "join_dht_res", MP.EncodeData(joinRes)))
}

func (dhtNode *DHTNode) HandleJoinRes(msg *MP.Message) (int,*Node) {
	var joinRes JoinResponse
	MP.DecodeData(&joinRes,msg.Data)
	var node *Node = nil

	if joinRes.Status == FAILURE {
		panic ("Join procedure for DHT failed")
	} else if (joinRes.Status == JOIN_IN_PROGRESS_RETRY_LATER) {
		node = &(Node{msg.Src,msg.SrcName})
	} else {
		/* SUCCESS case */
		fmt.Println("[DHT] Join Response with Success received")
		/* 1. Add received map to local DHT table */
		dhtNode.hashTable = joinRes.HashTable

		/* Update prev and next nodes */
		dhtNode.leafTable.nextNode = &(Node{msg.Src,msg.SrcName})

		/* If my successor node indicates that I am its predecessor, then it is a loop and
		 * we are the only 2 nodes in the DHT*/
		if (joinRes.Predecessor.IpAddress == dhtNode.ipAddress){
			dhtNode.leafTable.prevNode = dhtNode.leafTable.nextNode
		} else {
			dhtNode.leafTable.prevNode =&(Node{joinRes.Predecessor.IpAddress,""})
	}

		/* 2. Send Join complete to successor */
		dhtNode.mp.Send(MP.NewMessage(msg.Src, msg.SrcName, "join_dht_complete", MP.EncodeData(JoinComplete{SUCCESS, dhtNode.nodeKey})))

		/* 3. Send join notification to predecessor */
		dhtNode.mp.Send(MP.NewMessage(dhtNode.leafTable.prevNode.IpAddress, "", "join_dht_notify",
			                              MP.EncodeData(JoinNotify{dhtNode.nodeKey})))
	}

	return joinRes.Status,node
}

func (dhtNode *DHTNode) HandleJoinComplete(msg *MP.Message) {
	var joinComplete JoinComplete
	MP.DecodeData(&joinComplete,msg.Data)
	fmt.Println("[DHT] Join Complete received")

	/* Update routing information to include this new node */
	dhtNode.updateLeafAndPrefixTablesWithNewNode(msg.Src, msg.SrcName, joinComplete.Key,true)

	/* Delete entries transferred to new node */
	/* TODO After replication, this needs to be done in farthest replica */
	newNodeKey := getBigIntFromString(joinComplete.Key)
	var entryKey *big.Int

	for k,_ := range dhtNode.hashTable {
		entryKey = getBigIntFromString(k)
		/* If entry key is <= new node's key, remove the entry as it is already transferred to new node */
		if (entryKey.Cmp(newNodeKey) <= 0){
			delete(dhtNode.hashTable,k)
		}
	}

	dhtNode.isRingUpdateInProgress = false
}

func (dhtNode *DHTNode) HandleJoinNotify(msg *MP.Message) {
	var joinNotify JoinNotify
	MP.DecodeData(&joinNotify,msg.Data)
	fmt.Println("[DHT] Join Notify received")

	/* Update routing information to include this new node */
	dhtNode.updateLeafAndPrefixTablesWithNewNode(msg.Src, msg.SrcName, joinNotify.Key,false)
}

func (dhtNode *DHTNode) HandleBroadcastMessage(msg *MP.Message) {
	var broadcastMsg BroadcastMessage
	MP.DecodeData(&broadcastMsg,msg.Data)

	fmt.Println("Processing Handle Broadcast Request in "+dhtNode.nodeKey)
	if (broadcastMsg.OriginIpAddress == dhtNode.ipAddress) {
		/* Token returned back to us. Don't forward */
		fmt.Println("Nodes in the ring are ")
		for _, val := range broadcastMsg.TraversedNodesList {
			fmt.Println(val.IpAddress)
		}
	} else {
		/* Add current node details into the list. Currently we use this for debugging
		 * to understand the structure of the ring */
		node := Node{msg.Dest,msg.DestName}
		broadcastMsg.TraversedNodesList = append(broadcastMsg.TraversedNodesList,node)

		nextNode := dhtNode.leafTable.nextNode
		dhtNode.mp.Send(MP.NewMessage(nextNode.IpAddress, "", "dht_broadcast_msg",
			MP.EncodeData(broadcastMsg)))
	}
}
/*TODO add a parameter to take suitable payload for broadcast. For e.g. we can have type which
  describes about streaming group being newly launched */
func (dhtNode *DHTNode) CreateBroadcastMessage(){
	var broadcastMsg BroadcastMessage
	broadcastMsg.OriginIpAddress, broadcastMsg.OriginName= dhtNode.mp.GetNodeIpAndName()
	node:= Node{broadcastMsg.OriginIpAddress,broadcastMsg.OriginName}
	broadcastMsg.TraversedNodesList = append(broadcastMsg.TraversedNodesList,node)

	nextNode := dhtNode.leafTable.nextNode
	if (nextNode == nil){
		if (dhtNode.leafTable.prevNode == nil){
			fmt.Println("DHT with only one node")
		} else {
			panic ("Broken Ring. Next node cannot be nil if there are more than 1 node in DHT")
		}
		return
	}
	fmt.Println("Sending initial broad cast message with node key" + broadcastMsg.OriginName)
	dhtNode.mp.Send(MP.NewMessage(nextNode.IpAddress, "", "dht_broadcast_msg",
		MP.EncodeData(broadcastMsg)))
}

func (dhtNode *DHTNode) PerformPeriodicBroadcast(){
	ticker := time.NewTicker(time.Second * 25)
	go func() {
		for _ = range ticker.C {
			dhtNode.CreateBroadcastMessage()
		}
	}()
}

func (dhtNode *DHTNode) Leave(msg *MP.Message) {

}

func (dhtNode *DHTNode) Refresh(StreamingGroupID string) {

}



/* handler responsible for processing messages received from other nodes
 * and updating the local hash table
 */
func (dhtNode *DHTNode) HandleRequest() {

}

func (dhtNOde *DHTNode) GetNextNodeToForwardInRing(key string) (*Node){
	return dhtNOde.findSuccessor(key)
}


