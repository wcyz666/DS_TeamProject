package dnsService

import (
	"fmt"
	"net/http"
	"io/ioutil"
	"encoding/json"
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

