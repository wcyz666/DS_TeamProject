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