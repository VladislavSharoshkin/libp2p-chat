package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
	"libp2p-chat/controllers"
	"libp2p-chat/internal"
	"libp2p-chat/misk"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/libp2p/go-libp2p/p2p/net/connmgr"

	"github.com/libp2p/go-libp2p"
	gostream "github.com/libp2p/go-libp2p-gostream"
	p2phttp "github.com/libp2p/go-libp2p-http"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	// "github.com/libp2p/go-libp2p/core/routing"
	"github.com/libp2p/go-libp2p/p2p/security/noise"
	libp2ptls "github.com/libp2p/go-libp2p/p2p/security/tls"
)

func main() {
	run()
}

// discoveryNotifee gets notified when we find a new peer via mDNS discovery
type discoveryNotifee struct {
	h host.Host
}

func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	fmt.Printf("discovered new peer %s\n", pi.ID.Pretty())
	err := n.h.Connect(context.Background(), pi)
	if err != nil {
		fmt.Printf("error connecting to peer %s: %s\n", pi.ID.Pretty(), err)
	}
	_, err = internal.SavePeer(pi)
	if err != nil {
		log.Println(err)
	}
}

func run() {
	internal.DatabaseConnect()

	log.SetFlags(log.Lshortfile)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	_, _, err := crypto.GenerateKeyPair(
		crypto.Ed25519,
		-1,
	)
	if err != nil {
		panic(err)
	}

	//var idht *dht.IpfsDHT

	connmgr, err := connmgr.NewConnManager(
		100, // Lowwater
		400, // HighWater,
		connmgr.WithGracePeriod(time.Minute),
	)
	if err != nil {
		panic(err)
	}
	h2, err := libp2p.New(

		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/0",
			"/ip4/0.0.0.0/udp/0/quic",
		),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support noise connections
		libp2p.Security(noise.ID, noise.New),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		//libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
		//	idht, err = dht.New(ctx, h)
		//	return idht, err
		//}),

		libp2p.EnableNATService(),
	)
	if err != nil {
		panic(err)
	}
	defer h2.Close()

	s := mdns.NewMdnsService(h2, misk.ServiceName, &discoveryNotifee{h: h2})
	s.Start()

	//for _, addr := range dht.DefaultBootstrapPeers {
	//	pi, _ := peer.AddrInfoFromP2pAddr(addr)
	//	h2.Connect(ctx, *pi)
	//}

	listener, _ := gostream.Listen(h2, p2phttp.DefaultP2PProtocol)
	defer listener.Close()
	go func() {
		http.HandleFunc("/message/new", controllers.NewMessage)
		server := &http.Server{}
		server.Serve(listener)
	}()

	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(h2))
	//client := &http.Client{Transport: tr}

	ps, err := pubsub.NewGossipSub(ctx, h2)
	if err != nil {
		panic(err)
	}

	topic, err := ps.Join(misk.ServiceName)
	if err != nil {
		panic(err)
	}

	sub, err := topic.Subscribe()
	if err != nil {
		panic(err)
	}

	go func() {
		for {
			dhtMes, err := sub.Next(ctx)
			if err != nil {
				log.Println(err)
				continue
			}

			msg := internal.Message{}
			err = json.Unmarshal(dhtMes.Data, &msg)
			if err != nil {
				log.Println(err)
				continue
			}

			misk.PrintBlock(
				misk.ValueInfo("New message from", msg.From),
				misk.ValueInfo("To", msg.To),
				misk.ValueInfo("Date", msg.CreatedAt.Format("02.01.06 15:04")),
				misk.ValueInfo("Text", msg.Text),
			)
		}
	}()

	misk.PrintBlock(
		misk.ValueInfo("Welcome to", "libp2p-chat"),
		misk.ValueInfo("Coded by", "Vladislav Sharoshkin"),
		misk.ValueInfo("You ID", h2.ID()),
	)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := strings.Split(scanner.Text(), " ")

		switch command[0] {
		case "info":
			misk.PrintBlock(
				misk.ValueInfo("ID", h2.ID()),
				misk.ValueInfo("Peers", len(h2.Peerstore().Peers())),
			)
		case "clear":
			misk.ClearConsole()
		case "send":
			add, _ := peer.Decode(command[1])
			text := command[2]

			mes, _ := internal.MessageSend(topic, h2.ID(), add, text)

			misk.PrintBlock(
				misk.ValueInfo("Send message", mes.Text),
				misk.ValueInfo("To", mes.To),
			)
		case "put":
			err := topic.Publish(ctx, []byte(command[1]))
			if err != nil {
				log.Println(err)
			}
		case "get":

		}
	}
	select {}
}
