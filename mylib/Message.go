package main

type Message struct {
	dest string
	src string
	data string
}

func NewMessage(data string, dest string) *Message {
	msg := &Message{
		dest: dest,
		data: data
	}

	return msg
}