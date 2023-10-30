// to be remove just for test
package main

import (
	"context"
	"io"
	"log"
	"time"

	"github.com/segmentio/kafka-go"
)

func toberevome_consumer() {

	defer func() {
		if r := recover(); r != nil {
			log.Println("recovered consumer:", r)
		}
	}()
	toberevome_consumer_()

}

func toberevome_consumer_() {

	var keepOffset int64 = 0
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   []string{kafkaparam.brokeradr},
		Topic:     kafkaparam.topicname,
		Partition: int(kafkaparam.partition),
		MaxBytes:  20e6, // 20MB
	})

	var force bool = true
	for force {
		reader.SetOffset(keepOffset)
		for {
			message, err := reader.ReadMessage(context.Background())
			if err != nil {
				if err == io.EOF {
					log.Println("consume eof", err)
					time.Sleep(1000)
					break
				} else {
					log.Println("consume error", err)
					force = false
					break
				}
			}
			keepOffset = message.Offset
			log.Println("THE MESSAGE", string(string(message.Value)))
		}
	}

	if erre := reader.Close(); erre != nil {
		log.Println("failed to close reader:", erre)
	}
	// conn, err := kafka.DialLeader(context.Background(), "tcp", kafkaparam.brokeradr, kafkaparam.topicname, int(kafkaparam.partition))
	// if err != nil {
	// 	log.Println("consumer failed to dial leader:", err)
	// }

	// conn.SetReadDeadline(time.Now().Add(10 * time.Second))
	// var batch_mes *kafka.Batch
	// var b []byte
	// batch_mes = conn.ReadBatch(1e3, 100e6)
	// b = make([]byte, 10e6) // 10mb max per message
	// for {
	// 	n, err := batch_mes.Read(b)
	// 	if err == io.EOF {
	// 		log.Println("consumer EOF")
	// 		break
	// 	}
	// 	log.Println("THE MESSAGE", string(b[:n]))
	// }

	// if err := batch_mes.Close(); err != nil {
	// 	log.Println("consumer failed to close batch :", err)
	// }

	// if err := conn.Close(); err != nil {
	// 	log.Println("consumer failed to close connection:", err)
	// }

}
