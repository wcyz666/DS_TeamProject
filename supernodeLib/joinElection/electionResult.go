package joinElection

type ElectionResult struct {
    ParentIP string
    ParentName string
}

type ElectionBroadcastMessage struct {
    IP string
    Name string
    ChildCount int
}

func NewElectionBroadcastMessage(IP string, Name string, ChildCount int) ElectionBroadcastMessage {
    return ElectionBroadcastMessage{IP: IP, Name: Name, ChildCount: ChildCount}
}