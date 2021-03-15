package main

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hashicorp/memberlist"
	"github.com/prometheus/client_golang/prometheus"
)

// Implements memberlist.EventDelegate interface
type eventController struct {
}

func (e *eventController) NotifyJoin(n *memberlist.Node) {
	Logger.Printf("%s joined", n.Addr)
}

func (e *eventController) NotifyLeave(n *memberlist.Node) {
	Logger.Printf("%s left", n.Addr)

	// Cleanup the metrics
	udpRTT.Delete(prometheus.Labels{"source": localAddr, "destination": n.Addr.String(), "location": location})
	// this is reverted because we get info from the other host
	// and we're destination from it's point of view
	tcpRTT.Delete(prometheus.Labels{"source": n.Addr.String(), "destination": localAddr, "location": location})
}

func (e *eventController) NotifyUpdate(n *memberlist.Node) {
	// Not implemented
}

// implements memberlist.Delegate
type msgController struct {
	msgCh    chan []byte
	interval *time.Ticker
}

func (c *msgController) NotifyMsg(msg []byte) {
	c.msgCh <- msg
}
func (c *msgController) NodeMeta(limit int) []byte {
	// Not implemented
	return nil
}
func (c *msgController) LocalState(join bool) []byte {
	// Not implemented
	return nil
}
func (c *msgController) GetBroadcasts(overhead, limit int) [][]byte {
	// Not implemented
	return nil
}
func (c *msgController) MergeRemoteState(buf []byte, join bool) {
	// Not implemented
}

type message struct {
	From     string `json:"from"`
	To       string `json:"to"`
	Time     int64  `json:"timestamp"`
	Location string `json:"location"`
}

func newMessage(from, to string, ts int64) ([]byte, error) {
	if from == "" || to == "" || ts == 0 {
		return nil, fmt.Errorf("input arguments can't be empty")
	}
	m := message{
		From:     from,
		To:       to,
		Time:     ts,
		Location: location,
	}

	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}

	return b, nil
}

func (m *message) unmarshal(b []byte) error {
	if b == nil {
		return fmt.Errorf("input can't be nil")
	}
	if err := json.Unmarshal(b, &m); err != nil {
		return err
	}
	return nil
}

type Ping struct {
}

func (p Ping) AckPayload() []byte {
	return []byte("pAck")
}

func (p Ping) NotifyPingComplete(other *memberlist.Node, rtt time.Duration, payload []byte) {
	updateRTT("udp", localAddr, other.Addr.String(), rtt.Nanoseconds())
	Logger.Printf("[INFO] (UDP) Name: %s, Addr: %s, RTT: %v", other.Name, other.Addr, rtt)
}

// Returns current time in unixnano format *string
func timeNow() string {
	return strconv.FormatInt(time.Now().UnixNano(), 10)
}

func parseAddr(addresses string) []string {
	result := strings.Split(addresses, ",")
	// Remove our own address
	for i, addr := range result {
		if addr == localAddr {
			copy(result[i:], result[i+1:])
		}
	}
	return result
}
