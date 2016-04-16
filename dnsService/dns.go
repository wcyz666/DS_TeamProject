package dnsService

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
	"strconv"
)

/* decode only relevant portions which are important to us */
type DNSRecordEntry struct {
	record_type string
	name        string
	content     string
	id          float64
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
	recordEntry.id = recordMap["id"].(float64)
	return nil
}

type DNSResponseMessage struct {
	records []map[string]DNSRecordEntry
}

/*
 http://stackoverflow.com/questions/23558425/how-do-i-get-the-local-ip-address-in-go
*/
func ExternalIP() (string, error) {
	resp, err := http.Get("http://myexternalip.com/raw")
	if err != nil {
		os.Stderr.WriteString(err.Error())
		os.Stderr.WriteString("\n")
		os.Exit(1)
	}
	defer resp.Body.Close()
	var IP []byte
	IP = make([]byte, 100)
	n, err := resp.Body.Read(IP)
	return strings.TrimSpace(string(IP[:n])), err
}

func IsIpAlreadyRegistered(ipList []string, curIP string) bool {
	for _, ipAddr := range ipList {
		if ipAddr == curIP {
			return true
		}
	}
	return false
}

/* Just like Apocolyse who is believed to be the first mutant, you are the first node
 * in the network. So, you need to create the DHT. */
func AmIApocolypse(name string)(bool){
	curAddrList := GetAddr(name)
	extIP, _ := ExternalIP()
	return ((len(curAddrList) == 1) && IsIpAlreadyRegistered(curAddrList,extIP))
}
/* Name of the domain which is used to track super nodes.  Currently, we are using
   p2plive.phani.me as the name */
func RegisterSuperNode(name string) {
	curAddrList := GetAddr(name)
	extIP, _ := ExternalIP()

	if IsIpAlreadyRegistered(curAddrList, extIP) {
		fmt.Println("Node already registered as Super Node")
		return
	}

	fmt.Println("externalIP is " + extIP)
	AddAddr(name, extIP)
}

/*
NOTE: We receive a failure response with Bad Request as code when we attempt to add a new A record with an IP
address already present as an A record in the DNS server. Hence, before we call AddAddr, call GetAddr to check
if IP to be added is already present in one of the A records in the DNS system
*/
func AddAddr(name string, ipAddr string) error {
	url := "https://api.dnsimple.com/v1/domains/phani.me/records"
	//fmt.Println("URL:>", url)

	//var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
	var reqBody string = `{ "record": { "name": "` + name + `", "record_type": "A", "ttl": 3600, "prio": 10, "content": "` + ipAddr + `"}}`
	//fmt.Println("reqBody is " + reqBody)
	var jsonStr = []byte(reqBody)

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonStr))
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-DNSimple-Token", "phanishankarpr@gmail.com:oWxhZCENnNaLFq3WHyDEzpgYETguMyTC")
	req.Header.Set("Content-Type", "application/json")

	client := getHttpClient()
	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	//fmt.Println("response Status:", resp.Status)
	//fmt.Println("response Headers:", resp.Header)

	if resp.StatusCode != 201 {
		return fmt.Errorf("Request to add A record failed with code %d", resp.StatusCode)
	}
	return nil
}

func GetAddr(name string) []string {

	url := "https://api.dnsimple.com/v1/domains/phani.me/records"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-DNSimple-Token", "phanishankarpr@gmail.com:oWxhZCENnNaLFq3WHyDEzpgYETguMyTC")
	client := getHttpClient()

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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
	for i := 0; i < len(decData.records); i++ {
		entry = decData.records[i]["record"]
		if (entry.record_type == "A") && (entry.name == name) {
			//fmt.Println("record type is " + entry.record_type + " name is " + entry.name + " content is " +
			//	entry.content)
			addrList = append(addrList, entry.content)
		}

	}
	return addrList
}

func deleteAddrRecord(id float64) {
	client := getHttpClient()
	url := "https://api.dnsimple.com/v1/domains/phani.me/records/" + strconv.FormatFloat(id, 'f', -1, 64)
	//fmt.Println("URL:>", url)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-DNSimple-Token", "phanishankarpr@gmail.com:oWxhZCENnNaLFq3WHyDEzpgYETguMyTC")
	req.Header.Set("Content-Type", "application/json")

	resp, respErr := client.Do(req)
	if respErr != nil {
		panic(respErr)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		fmt.Println("response Status:", resp.Status)
		panic("Response code not OK")
	}
}

func ClearAddrRecords(name string, ipAddress string) {

	url := "https://api.dnsimple.com/v1/domains/phani.me/records"

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")
	req.Header.Set("X-DNSimple-Token", "phanishankarpr@gmail.com:oWxhZCENnNaLFq3WHyDEzpgYETguMyTC")
	client := getHttpClient()

	resp, err := client.Do(req)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
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
	for i := 0; i < len(decData.records); i++ {
		entry = decData.records[i]["record"]
		if (entry.record_type == "A") && (entry.name == name) {
			if (ipAddress == "" || ipAddress == entry.content){
				//fmt.Println( "name is " + entry.name + " id is " + strconv.FormatFloat(entry.id, 'f', -1, 64))
				deleteAddrRecord(entry.id)
			}
		}
	}
}

func getHttpClient() *http.Client {
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	return client
}
