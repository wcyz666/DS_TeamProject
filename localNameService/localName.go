package localNameService

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"fmt"
)

func getFirstMac() string {
	interfaces, err := net.Interfaces()
	if err != nil {
		panic("Poor soul, here is what you got: " + err.Error())
	}
	inter := interfaces[0]
	fmt.Println("HW address is "+ inter.HardwareAddr.String())
	return inter.HardwareAddr.String()
}

func hash(text string) string {
	hasher := md5.New()
	hasher.Write([]byte(text))
	return hex.EncodeToString(hasher.Sum(nil))
}

/*
   GetLocalName - return md5(first Mac address of the node) as the
       local name of the node
*/
func GetLocalName() string {
	return hash(getFirstMac())
}
