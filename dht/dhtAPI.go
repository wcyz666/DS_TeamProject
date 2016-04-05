package dht

/*
 * DHT APIs Implementation.
*/

func (dht *DHT) Get(streamingGroupID string) ([]MemberShipInfo, int) {
	if dht.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		return dht.getData(streamingGroupID)
	} else {
		/* TODO fetch data from other node */
		return make([]MemberShipInfo, 0), SUCCESS
	}
}

func (dht *DHT) Create(streamingGroupID string, data MemberShipInfo) (int){
	status:= SUCCESS
	if dht.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.createEntry(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
	return status
}

func (dht *DHT) Delete(streamingGroupID string) (int) {
	status:= SUCCESS
	if dht.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.deleteEntry(streamingGroupID)
	} else {
		/* TODO send update to other node */
	}
	return status
}

func (dht *DHT) Append(streamingGroupID string, data MemberShipInfo) (int) {
	status := SUCCESS
	if dht.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status =  dht.appendData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
	return status
}

func (dht *DHT) Remove(streamingGroupID string, data MemberShipInfo) (int){
	status := SUCCESS
	if dht.isKeyPresentInMyKeyspaceRange(streamingGroupID) {
		status = dht.removeData(streamingGroupID, data)
	} else {
		/* TODO send update to other node */
	}
	return status
}
