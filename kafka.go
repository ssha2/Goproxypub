package main

import (
	"bytes"
	"context"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

type LogElem struct {
	ltype string    // response or request or error
	u_id  string    // session id
	head  []byte    // headers
	body  []byte    // body
	t     time.Time //timestamp
}

type Kafkaparam struct {
	topicname string
	brokeradr string
	partition uint
}

var kafkaparam Kafkaparam

func configkafka(topicname string, brokeradr string) {
	kafkaparam = Kafkaparam{topicname, brokeradr, 0}
}

func bytestoKafka(elem LogElem) {

	var buffer bytes.Buffer
	buffer.WriteString("$$meta$$")
	buffer.WriteString(elem.ltype)
	buffer.WriteString("\n$$id$$")
	buffer.WriteString(elem.u_id)
	buffer.WriteString("\n$$time$$")
	buffer.WriteString(elem.t.String())
	buffer.WriteString("\n$$head$$")
	buffer.Write(elem.head)
	buffer.WriteString("\n$$body$$")
	buffer.Write(elem.body)
	senttoKofkaRecover(buffer.Bytes(), []byte(elem.ltype))
}

func senttoKofkaRecover(b []byte, k []byte) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered panic:", r)
		}
	}()
	senttoKofka(b, k)
}

func senttoKofka(b []byte, k []byte) {

	conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaparam.brokeradr, kafkaparam.topicname, int(kafkaparam.partition))
	if err != nil {
		log.Println("failed to dial leader:", err)
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
	_, err = conn.WriteMessages(
		kafka.Message{Key: k, Value: b})
	if err != nil {
		log.Println("failed to write messages:", err)
	}

	if err := conn.Close(); err != nil {
		log.Println("failed to close writer:", err)
	}
}
