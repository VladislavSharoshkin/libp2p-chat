package main

import (
	"bufio"
	"context"
	"encoding/json"
	"libp2p-chat/controllers"
	"libp2p-chat/internal"
	"libp2p-chat/misk"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/jessevdk/go-flags"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"

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

var MyHost host.Host
var Ctx = context.Background()
var options Options
var parser = flags.NewParser(&options, flags.Default)

type Options struct {
	Instance string `short:"i" long:"instance" description:"Instance name" default:"chat"`
}

type discoveryNotification struct {
	h host.Host
}

func main() {
	run()
}

func (n *discoveryNotification) HandlePeerFound(pi peer.AddrInfo) {
	err := n.h.Connect(Ctx, pi)
	if err != nil {
		log.Printf("error connecting to peer %s: %s\n", pi.ID, err)
	}
	_, err = internal.SavePeer(pi)
	if err != nil {
		log.Println(err)
	}
}

func ProcessDhtMessage(sub *pubsub.Subscription) {
	dhtMes, err := sub.Next(context.Background())
	if err != nil {
		log.Println(err)
		return
	}

	mes := internal.Message{}
	err = json.Unmarshal(dhtMes.Data, &mes)
	if err != nil {
		log.Println(err)
		return
	}

	internal.ProcessMessage(mes, MyHost.Peerstore().PrivKey(MyHost.ID()))
}

func run() {
	log.SetFlags(log.Lshortfile)

	if _, err := parser.Parse(); err != nil {
		switch flagsErr := err.(type) {
		case flags.ErrorType:
			if flagsErr == flags.ErrHelp {
				os.Exit(0)
			}
			os.Exit(1)
		default:
			os.Exit(1)
		}
	}

	internal.DatabaseConnect()

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
	MyHost, err = libp2p.New(

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
	defer MyHost.Close()

	s := mdns.NewMdnsService(MyHost, misk.ServiceName, &discoveryNotification{h: MyHost})
	s.Start()

	//for _, addr := range dht.DefaultBootstrapPeers {
	//	pi, _ := peer.AddrInfoFromP2pAddr(addr)
	//	MyHost.Connect(ctx, *pi)
	//}

	listener, _ := gostream.Listen(MyHost, p2phttp.DefaultP2PProtocol)
	defer listener.Close()
	go func() {
		http.HandleFunc("/message/new", controllers.NewMessage)
		server := &http.Server{}
		server.Serve(listener)
	}()

	tr := &http.Transport{}
	tr.RegisterProtocol("libp2p", p2phttp.NewTransport(MyHost))
	//client := &http.Client{Transport: tr}

	ps, err := pubsub.NewGossipSub(ctx, MyHost)
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
			ProcessDhtMessage(sub)
		}
	}()

	misk.PrintBlock(
		misk.ValueInfo("Welcome to", "libp2p-chat"),
		misk.ValueInfo("Coded by", "Vladislav Sharoshkin"),
		misk.ValueInfo("You ID", MyHost.ID()),
	)

	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		command := strings.Split(scanner.Text(), " ")

		switch command[0] {
		case "info":
			misk.PrintBlock(
				misk.ValueInfo("ID", MyHost.ID()),
				misk.ValueInfo("Peers", len(MyHost.Peerstore().Peers())),
			)
		case "clear":
			misk.ClearConsole()
		case "send":
			add, err := peer.Decode(command[1])
			if err != nil {
				log.Println(err)
			}
			text := command[2]

			internal.MessageSend(topic, MyHost.ID(), add, text, MyHost.Peerstore().PrivKey(MyHost.ID()))

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
