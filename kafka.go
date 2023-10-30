package main

import (
	"bytes"
	"log"
	"time"

	"github.com/confluentinc/confluent-kafka-go/v2/kafka"
)

type LogElem struct {
	ltype string    // respomse or request or error
	u_id  string    // session id
	head  []byte    // headers
	body  []byte    // body
	t     time.Time //timestamp
}

type ConfigKafka struct {
	config *kafka.ConfigMap
	topic  string
}

var configKafka ConfigKafka

func configkafka(topicname string, brokeradr string) {
	configKafka = ConfigKafka{
		&kafka.ConfigMap{
			"bootstrap.servers": brokeradr,
		},
		topicname,
	}
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
	kafka_simle_one(buffer.Bytes(), []byte(elem.ltype))
}

func kafka_simle_one(b []byte, k []byte) {

	producer, err := kafka.NewProducer(configKafka.config)
	if err != nil {
		log.Printf("Failed create producer: %v\n", err)
		return //silent exit
	}

	// Listen to all the events on the default events channel
	// go func() {
	// 	for e := range producer.Events() {
	// 		switch ev := e.(type) {
	// 		case *kafka.Message:
	// 			// The message delivery report, indicating success or
	// 			// permanent failure after retries have been exhausted.
	// 			// Application level retries won't help since the client
	// 			// is already configured to do that.
	// 			m := ev
	// 			if m.TopicPartition.Error != nil {
	// 				log.Printf("Delivery failed: %v\n", m.TopicPartition.Error)
	// 			} else {
	// 				log.Printf("Delivered message to topic %s [%d] at offset %v\n",
	// 					*m.TopicPartition.Topic, m.TopicPartition.Partition, m.TopicPartition.Offset)
	// 			}
	// 		case kafka.Error:
	// 			// Generic client instance-level errors, such as
	// 			// broker connection failures, authentication issues, etc.
	// 			//
	// 			// These errors should generally be considered informational
	// 			// as the underlying client will automatically try to
	// 			// recover from any errors encountered, the application
	// 			// does not need to take action on them.
	// 			log.Printf("Error: %v\n", ev)
	// 		default:
	// 			log.Printf("Ignored event: %s\n", ev)
	// 		}
	// 	}
	// }()

	err = producer.Produce(&kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &configKafka.topic, Partition: kafka.PartitionAny},
		Value:          b,
		Key:            k,
	}, nil)
	if err != nil {
		log.Printf("Failed to produce message: %v\n", err)
	}

	// Flush and close the producer and the events channel
	for producer.Flush(1000) > 0 {
		//just wait
	}
	producer.Close()

}
