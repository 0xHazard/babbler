package main

import (
	"flag"
	"log"
	"math"
	"os"
	"time"

	"github.com/hashicorp/logutils"
	"github.com/hashicorp/memberlist"
)

var (
	localAddr, location, remoteAddr string
	localPort, mport                int
	checkInterval                   int64
)

var Logger *log.Logger = log.New(os.Stdout, "", log.Lmicroseconds)

func init() {
	flag.Set("logtostderr", "true")
	flag.Set("v", "1")
	flag.StringVar(&remoteAddr, "bootstrap", "", "remote host to bootstrap. Can be multiple comma-separated hosts")
	flag.StringVar(&localAddr, "addr", "0.0.0.0", "local address")
	flag.IntVar(&localPort, "port", 7964, "local port")
	flag.IntVar(&mport, "mport", 8080, "metrics port")
	flag.Int64Var(&checkInterval, "interval", 3, "checks interval")
	flag.StringVar(&location, "location", "local", "host location / datacenter")
	flag.Parse()
}

func main() {

	Logger.SetOutput(&logutils.LevelFilter{
		Levels:   []logutils.LogLevel{"DEBUG", "INFO", "WARN", "ERROR"},
		MinLevel: logutils.LogLevel("INFO"),
		Writer:   os.Stdout,
	})

	// run metric exporter
	go exporter()

	config := newConfig(localAddr, localPort, remoteAddr, location, checkInterval)
	list, err := memberlist.Create(config)
	if err != nil {
		panic("Failed to create memberlist: " + err.Error())
	}

	// Connect to existing cluster
	if remoteAddr != "" {
		go func() {

			interval := 10.0
			maxInterval := 300
			multiplier := 1.2

			for {
				// Join an existing cluster by specifying at least one known member.
				_, err = list.Join(parseAddr(remoteAddr))
				if err != nil {
					config.Logger.Printf("[ERROR] Failed to join cluster: " + err.Error())
					// Back-off
					time.Sleep(time.Duration(interval) * time.Second)
					if interval < float64(maxInterval) {
						interval = math.Pow(interval, multiplier)
					}
					continue
				}
				break
			}
		}()
	}
	config.Logger.Printf("Ready to babble.")

	// Main loop
	for {
		select {
		case <-config.Delegate.(*msgController).interval.C:
			for _, member := range list.Members() {
				if member.Addr.String() != localAddr {
					msg, err := newMessage(localAddr, member.Addr.String(), time.Now().UnixNano())
					if err != nil {
						config.Logger.Printf("[ERROR] %v", err)
						continue
					}
					list.SendReliable(member, msg)
				}
			}
		case data := <-config.Delegate.(*msgController).msgCh:
			var msg message
			if err := msg.unmarshal(data); err != nil {
				config.Logger.Printf("[ERROR] couldn't parse message, %v", err)
				continue
			}
			rtt := time.Now().UnixNano() - msg.Time
			updateRTT("tcp", msg.From, msg.To, rtt)
			config.Logger.Printf("[INFO] (TCP) From: %q , To: %q, RTT: %dns", msg.From, msg.To, rtt)
		}
	}

}
