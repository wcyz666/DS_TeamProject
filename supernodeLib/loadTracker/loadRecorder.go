package loadTracker


type LoadBroadcastMessage struct {
    InitNodeIP string
    InitNodeName string
    SuperNodeUsages []SuperNodeUsage
}


type SuperNodeUsage struct {
    IP string
    Name string
    ChildCount int
}

func NewUsage(IP string, Name string, ChildCount int) SuperNodeUsage {
    return SuperNodeUsage{IP: IP, Name: Name, ChildCount: ChildCount};
}

func NewTracker(IP string, Name string, sNU SuperNodeUsage) *LoadBroadcastMessage {
    SuperNodeUsages := make([]SuperNodeUsage, 1);
    SuperNodeUsages[0] = sNU
    return &LoadBroadcastMessage{InitNodeIP: IP, InitNodeName: Name, SuperNodeUsages: SuperNodeUsages}
}

