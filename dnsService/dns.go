package dnsService

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
	"bytes"
	"net"
	"errors"
)

/* decode only relevant portions which are important to us */
type DNSRecordEntry struct{
	record_type string
	name string
	content string
}

func (recordEntry *DNSRecordEntry) UnmarshalJSON(b []byte) (err error) {
	var recordMap map[string]interface{}
	err = json.Unmarshal(b, &recordMap)
	if err != nil {
		return err
	}
	recordEntry.content = recordMap["content"].(string)
	recordEntry.record_type = recordMap["record_type"].(string)
	recordEntry.name = recordMap["name"].(string)
	return nil
}

type DNSResponseMessage struct {
	records[]  map[string]DNSRecordEntry
}

/*
 http://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
*/
func ExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Unable to get a valid IP address")
}

func isIpAlreadyRegistered(ipList []string, curIP string) bool {
	for _,ipAddr := range ipList{
		if ipAddr == curIP{
			return true
		}
	}
	return false
}

/* Name of the domain which is used to track super nodes.  Currently, we are using
   p2plive.phani.me as the name */
func RegisterSuperNode(name string){
	curAddrList := GetAddr(name)
	extIP,_ := ExternalIP()

	if (isIpAlreadyRegistered(curAddrList, extIP)) {
		fmt.Println("Node already registered as Super Node")
		return
	}

	fmt.Println("externalIP is "+ extIP)
	AddAddr(name, extIP)
}

/*
NOTE: We receive a failure response with Bad Request as code when we attempt to add a new A record with an IP
address already present as an A record in the DNS server. Hence, before we call AddAddr, call GetAddr to check
if IP to be added is already present in one of the A records in the DNS system
*/
func AddAddr(name string, ipAddr string) (error) {
	url := "https://api.dnsimple.com/v1/domains/phani.me/records"
	fmt.Println("URL:>", url)

	//var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	var reqBody string = `{ "record": { "name": "` + name + `", "record_type": "A", "ttl": 3600, "prio": 10, "content": "` + ipAddr + `"}}`
	fmt.Println("reqBody is "+ reqBody)
	var jsonStr = []byte(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-DNSimple-Token", "phanishankarpr@gmail.com:oWxhZCENnNaLFq3WHyDEzpgYETguMyTC")
	req.Header.Set("Content-Type","application/json")



	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	fmt.Println("response Status:", resp.Status)
	fmt.Println("response Headers:", resp.Header)

	if resp.StatusCode != 201 {
		return fmt.Errorf("Request to add A record failed with code %d",resp.StatusCode)
	}
	return nil
}

func GetAddr(name string) ([] string){

	url := "https://api.dnsimple.com/v1/domains/phani.me/records"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-DNSimple-Token", "phanishankarpr@gmail.com:oWxhZCENnNaLFq3WHyDEzpgYETguMyTC")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if (resp.StatusCode != 200){
		fmt.Println("response Status:", resp.Status)
		panic("Response code not OK")
	}

	body, _ := ioutil.ReadAll(resp.Body)
	var decData DNSResponseMessage
	err = json.Unmarshal(body, &(decData.records))
	if err != nil {
		panic(err)
	}

	var entry DNSRecordEntry
	var addrList []string
	for i :=0; i < len(decData.records);i++ {
		entry = decData.records[i]["record"]
		if ((entry.record_type =="A") && (entry.name == name)){
			fmt.Println("record type is "+entry.record_type + " name is "+ entry.name + " content is "+
			entry.content)
			addrList = append(addrList,entry.content)
		}

	}
	return addrList
}

