package client

import (
	"context"
	"fmt"
	"github.com/BaliAutomation/sensetif-datasource/pkg/model"
	"github.com/apache/pulsar-client-go/pulsar"
	"github.com/grafana/grafana-plugin-sdk-go/backend/log"
	"time"
)

type Pulsar interface {
	Send(topic string, key string, value []byte)
}

type PulsarClient struct {
	client    pulsar.Client
	producers map[string]pulsar.Producer
}

func (p *PulsarClient) Send(topic string, key string, value []byte) string {
	topic = model.Namespace + "/" + topic
	parts, e := p.client.TopicPartitions(topic)
	if e != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed to create a producer for topic %s - Error=%+v", topic, e))
		return ""
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Partitions of %s : %+v", topic, parts))
	}
	producer := p.producers[topic]
	if producer == nil {
		var err error
		options := pulsar.ProducerOptions{
			Topic:           topic,
			DisableBatching: true,
		}
		producer, err = p.client.CreateProducer(options)
		if err != nil {
			log.DefaultLogger.Error(fmt.Sprintf("Failed to create a producer for topic %s - Error=%+v", topic, err))
			return ""
		} else {
			log.DefaultLogger.Info(fmt.Sprintf("Created a new producer for topic %s", topic))
		}
		p.producers[topic] = producer
		defer p.producers[topic].Close()
	}
	message := &pulsar.ProducerMessage{
		Payload: value,
		Key:     key,
	}
	msgId, err := producer.Send(context.Background(), message)
	if err != nil {
		log.DefaultLogger.Error(fmt.Sprintf("Failed to send a message: %s\n%s : %+v\n", err, message.Key, message.Value))
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Sent message on topic %s with key %s. Id: %s. Data: %+v\n", producer.Topic(), message.Key, msgId, message.Payload))
	}
	return string(msgId.Serialize())
}

func (p *PulsarClient) InitializePulsar(pulsarHosts string, clientId string) {
	var err error
	p.client, err = pulsar.NewClient(pulsar.ClientOptions{
		URL:               pulsarHosts,
		ConnectionTimeout: 30 * time.Second,
		OperationTimeout:  30 * time.Second,
	})
	if err != nil {
		log.DefaultLogger.Error("Failed to initialize Pulsar: " + err.Error())
		return
	} else {
		log.DefaultLogger.Info(fmt.Sprintf("Connecting %s to Pulsar cluster %s.", clientId, pulsarHosts))
	}
	defer p.client.Close()
}
