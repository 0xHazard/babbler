package main

import (
	"os"
	"time"

	"github.com/hashicorp/memberlist"
)

func newConfig(localAddr string, localPort int, remoteAddr, location string, checkInterval int64) *memberlist.Config {
	hostname, _ := os.Hostname()
	interval := time.Duration(checkInterval)
	return &memberlist.Config{
		Name:                    hostname,
		BindAddr:                localAddr,
		BindPort:                localPort,
		AdvertiseAddr:           "",
		AdvertisePort:           localPort,
		ProtocolVersion:         memberlist.ProtocolVersion2Compatible,
		TCPTimeout:              30 * time.Second,       // Timeout after 30 seconds
		IndirectChecks:          3,                      // Use 3 nodes for the indirect ping
		RetransmitMult:          4,                      // Retransmit a message 4 * log(N+1) nodes
		SuspicionMult:           6,                      // Suspect a node for 6 * log(N+1) * Interval
		SuspicionMaxTimeoutMult: 6,                      // For 10k nodes this will give a max timeout of 120 seconds
		PushPullInterval:        60 * time.Second,       // Low frequency
		ProbeTimeout:            3 * time.Second,        // Reasonable RTT time for WAN
		ProbeInterval:           interval * time.Second, // Failure check every N seconds
		DisableTcpPings:         false,                  // TCP pings are safe, even with mixed versions
		AwarenessMaxMultiplier:  8,                      // Probe interval backs off to 8 seconds

		GossipNodes:          4,                      // Gossip to 4 nodes
		GossipInterval:       500 * time.Millisecond, // Gossip more rapidly
		GossipToTheDeadTime:  60 * time.Second,       // Same as push/pull
		GossipVerifyIncoming: true,
		GossipVerifyOutgoing: true,

		EnableCompression: true, // Enable compression by default

		SecretKey: nil,
		Keyring:   nil,

		DNSConfigPath: "/etc/resolv.conf",

		HandoffQueueDepth: 1024,
		UDPBufferSize:     1400,
		CIDRsAllowed:      nil, // same as allow all

		Ping: &Ping{},
		Delegate: &msgController{
			msgCh:    make(chan []byte),
			interval: time.NewTicker(interval * time.Second),
		},
		Events: &eventController{},
		Logger: Logger,
	}
}
