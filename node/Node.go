package node

import (
    dns "../dnsService"
    //mp "../messagePasser/"
    "fmt"
)


func Start()  {
    dns.AddAddr("wtheproject.zone", "124.124.124.124")
    fmt.Print("wcyz666")
    fmt.Print(dns.GetAddr("wtheproject.zone"))
}
