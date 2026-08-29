package otelkafka

import (
    "github.com/confluentinc/confluent-kafka-go/v2/kafka"
    "go.opentelemetry.io/otel/propagation"
)

type KafkaHeaderCarrier []kafka.Header

// Ensure it implements propagation.TextMapCarrier
var _ propagation.TextMapCarrier = (*KafkaHeaderCarrier)(nil)

// This is necessary to implement the propagation.TextMapCarrier interface for Kafka headers.
func (c *KafkaHeaderCarrier) Get(key string) string {
    for _, h := range *c {
        if h.Key == key {
            return string(h.Value)
        }
    }
    return ""
}

// This is necessary to implement the propagation.TextMapCarrier interface for Kafka headers.
func (c *KafkaHeaderCarrier) Set(key string, value string) {
    // Update if exists
    for i, h := range *c {
        if h.Key == key {
            (*c)[i].Value = []byte(value)
            return
        }
    }
    // Otherwise append new header
    *c = append(*c, kafka.Header{
        Key:   key,
        Value: []byte(value),
    })
}

// This is necessary to implement the propagation.TextMapCarrier interface for Kafka headers.
func (c *KafkaHeaderCarrier) Keys() []string {
    keys := make([]string, 0, len(*c))
    for _, h := range *c {
        keys = append(keys, h.Key)
    }
    return keys
}