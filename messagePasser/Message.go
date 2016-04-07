package mylib

import (
	"bytes"
	"encoding/gob"
	"fmt"
)

type Message struct {
	Dest     string
	Src      string
	SrcName  string
	DestName string
	Kind     string
	Data     [] byte
}

// If you don't know the destName, pass an empty string
func NewMessage(dest string, destName string, kind string, data [] byte) Message {
	msg := Message{Dest: dest, DestName: destName, Kind: kind, Data: data}

	return msg
}

// @override encoding
func (d *Message) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(d.Dest)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.Src)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.SrcName)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.Kind)
	if err != nil {
		return nil, err
	}
	err = encoder.Encode(d.Data)
	if err != nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// @override decoding
func (d *Message) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&d.Dest)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.Src)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.SrcName)
	if err != nil {
		return err
	}
	err = decoder.Decode(&d.Kind)
	if err != nil {
		return err
	}
	return decoder.Decode(&d.Data)
}

func (d *Message) Serialize() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(d)
        //return buffer.Bytes(),err
	// Use \xfe as the delimiter
	return append(buffer.Bytes(), 254), err
}

func (d *Message) Deserialize(buffer []byte) error {
	buf := bytes.NewBuffer(buffer)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(d)
	return err
}

func EncodeData(data interface {})([]byte){
	var buffer = new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(data)
	if (err != nil){
		panic(err)
	}
	return buffer.Bytes()
}

func DecodeData(decData interface {}, encData[] byte)(error){
	buf := bytes.NewBuffer(encData)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(decData)
	return err
}

func main() {

	// An example on how to use the serialize/deserialize feature
	msg := Message{Dest: "bob", Src: "alice", Data: EncodeData("hi!")}
	// Serialize msg into bytes.Buffer
	buffer, _ := msg.Serialize()
	fmt.Println(buffer)

	// Create a new struct and deserialize into it
	msg2 := new(Message)
	fmt.Println(msg2)
	msg2.Deserialize(buffer)
	fmt.Println(msg2)

	var strData string
	DecodeData(&strData,msg2.Data)
}
