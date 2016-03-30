package mylib

import (
	"encoding/gob"
	"bytes"
	"fmt"
)

type Message struct {
	dest string
	src string
	kind string
	data string
}

func NewMessage(dest string, data string) Message {
	msg := Message{dest:dest, data:data}
	return msg
}

// @override encoding
func (d *Message) GobEncode() ([]byte, error) {
	w := new(bytes.Buffer)
	encoder := gob.NewEncoder(w)
	err := encoder.Encode(d.dest)
	if err!=nil {
		return nil, err
	}
	err = encoder.Encode(d.src)
	if err!=nil {
		return nil, err
	}
	err = encoder.Encode(d.data)
	if err!=nil {
		return nil, err
	}
	return w.Bytes(), nil
}

// @override decoding
func (d *Message) GobDecode(buf []byte) error {
	r := bytes.NewBuffer(buf)
	decoder := gob.NewDecoder(r)
	err := decoder.Decode(&d.dest)
	if err!=nil {
		return err
	}
	err = decoder.Decode(&d.src)
	if err!=nil {
		return err
	}
	return decoder.Decode(&d.data)
}

func (d *Message) Serialize() ([]byte, error) {
	var buffer = new(bytes.Buffer)
	enc := gob.NewEncoder(buffer)
	err := enc.Encode(d)
	// Use \xfe as the delimiter
	return append(buffer.Bytes(), 254), err
}


func (d *Message) Deserialize(buffer []byte) (error) {
	buf := bytes.NewBuffer(buffer)
	dec := gob.NewDecoder(buf)
	err := dec.Decode(d)
	return err
}

/*
Setters and getters
 */
func (d *Message) GetDest() (string) {
	return d.dest
}

func (d *Message) GetSrc() (string) {
	return d.src
}

func (d *Message) GetData() (string) {
	return d.data
}

func (d *Message) SetDest(dest string) {
	d.dest = dest
}

func (d *Message) SetSrc(src string) {
	d.src = src
}

func (d *Message) SetData(data string) {
	d.data = data
}



func main() {

	// An example on how to use the serialize/deserialize feature
	msg := Message{dest:"bob", src:"alice", data:"hi!"}
	// Serialize msg into bytes.Buffer
	buffer, _ := msg.Serialize()
	fmt.Println(buffer)

	// Create a new struct and deserialize into it
	msg2 := new(Message)
	fmt.Println(msg2)
	msg2.Deserialize(buffer)
	fmt.Println(msg2)
}
