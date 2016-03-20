package dnsService

/**
This is the fake dns map service
Just hardcode all the dns/ip in a map now
To be modified
 */

var dnsMap = make(map[string]string)

func init(){
	dnsMap["alice"] = "127.0.0.1"
	dnsMap["bob"] = "127.0.0.1"
}

func GetAddr(name string) (string){
	return dnsMap[name]
}