package main

import (
	"fmt"
	"log"
	"net"
	"time"
)

type HostedGraphite struct {
	Host   string
	Port   int
	APIKey string
	conn   net.Conn
}

func NewHostedGraphite(api_key string) (*HostedGraphite, error) {
	graphite := &HostedGraphite{Host: "carbon.hostedgraphite.com", Port: 2003, APIKey: api_key}
	err := graphite.connect()
	if err != nil {
		return nil, err
	}

	return graphite, nil
}

func (graphite *HostedGraphite) Send(stat string, value interface{}) error {
	metric := NewMetric(stat, value, time.Now())
	return graphite.sendMetric(metric)
}

func (graphite *HostedGraphite) connect() error {
	address := fmt.Sprintf("%s:%d", graphite.Host, graphite.Port)

	conn, err := net.DialTimeout("udp", address, 5*time.Second)
	if err != nil {
		return err
	}

	graphite.conn = conn
	return nil
}

func (graphite *HostedGraphite) sendMetric(metric Metric) error {
	payload := fmt.Sprintf("%s.%s %v %d\n", graphite.APIKey, metric.name, metric.value, metric.timestamp.Unix())
	log.Println("Sending metric", payload)
	_, err := graphite.conn.Write([]byte(payload))
	if err != nil {
		log.Println("Error recording metric", payload, err)
	}

	return err
}

type Metric struct {
	name      string
	value     interface{}
	timestamp time.Time
}

func NewMetric(name string, value interface{}, timestamp time.Time) Metric {
	return Metric{name, value, timestamp}
}
