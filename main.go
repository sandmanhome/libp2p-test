package main

import (
	"context"
	"fmt"

	csms "github.com/libp2p/go-conn-security-multistream"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/metrics"
	"github.com/libp2p/go-libp2p-core/peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
	pstoremem "github.com/libp2p/go-libp2p-peerstore/pstoremem"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	secio "github.com/libp2p/go-libp2p-secio"
	swarm "github.com/libp2p/go-libp2p-swarm"
	tptu "github.com/libp2p/go-libp2p-transport-upgrader"
	yamux "github.com/libp2p/go-libp2p-yamux"
	bhost "github.com/libp2p/go-libp2p/p2p/host/basic"
	msmux "github.com/libp2p/go-stream-muxer-multistream"
	"github.com/libp2p/go-tcp-transport"

	"github.com/multiformats/go-multiaddr"
)

func main() {
	// The context governs the lifetime of the libp2p node.
	// Cancelling it will stop the the host.
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Now, normally you do not just want a simple host, you want
	// that is fully configured to best support your p2p application.
	// Let's create a second host setting some more options.

	// Set your own keypair
	priv, _, err := crypto.GenerateKeyPair(
		crypto.Sm2p256v1, // Select your key type. Ed25519 are nice short
		-1,               // Select key length when possible (i.e. RSA).
	)
	if err != nil {
		panic(err)
	}

	//psk := []byte{0x55, 0x48, 0xb7, 0xf2, 0xeb, 0xd9, 0x56, 0x80, 0x81, 0xab, 0x23, 0xa6, 0x1f, 0x15, 0xff, 0x68, 0x24, 0xe8, 0x35, 0xe5, 0xca, 0xc4, 0xc5, 0x1c, 0xaf, 0x27, 0xd8, 0xbc, 0x5c, 0x3e, 0x6c, 0x4d}

	// h, err := libp2p.New(ctx,
	// 	// Use the keypair we generated
	// 	libp2p.Identity(priv),
	// 	// Multiple listen addresses
	// 	libp2p.ListenAddrStrings(
	// 		"/ip4/0.0.0.0/tcp/9392", // regular tcp connections
	// 	),
	// 	libp2p.PrivateNetwork(psk),
	// )
	// if err != nil {
	// 	panic(err)
	// }

	pid, err := peer.IDFromPublicKey(priv.GetPublic())
	if err != nil {
		panic(err)
	}

	pstore := pstoremem.NewPeerstore()

	if err := pstore.AddPrivKey(pid, priv); err != nil {
		panic(err)
	}
	if err := pstore.AddPubKey(pid, priv.GetPublic()); err != nil {
		panic(err)
	}

	s := swarm.NewSwarm(ctx, pid, pstore, metrics.NewBandwidthCounter())

	id := s.LocalPeer()
	pk := s.Peerstore().PrivKey(id)
	secMuxer := new(csms.SSMuxer)
	secMuxer.AddTransport(secio.ID, &secio.Transport{
		LocalID:    id,
		PrivateKey: pk,
	})

	stMuxer := msmux.NewBlankTransport()
	stMuxer.AddTransport("/yamux/1.0.0", yamux.DefaultTransport)

	upgrader := &tptu.Upgrader{
		Secure:  secMuxer,
		Muxer:   stMuxer,
		Filters: s.Filters,
	}

	//upgrader.ConnGater = cfg.connectionGater
	tcpTransport := tcp.NewTCPTransport(upgrader)
	//tcpTransport.DisableReuseport = cfg.disableReuseport

	if err := s.AddTransport(tcpTransport); err != nil {
		panic(err)
	}

	var ZeroLocalTCPAddress, _ = multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/9000")
	if err := s.Listen(ZeroLocalTCPAddress); err != nil {
		panic(err)
	}

	s.Peerstore().AddAddrs(pid, s.ListenAddresses(), peerstore.PermanentAddrTTL)

	bh := bhost.New(s)

	fmt.Printf("Hello World, my second hosts ID is %s\n", bh.ID())

	// The last step to get fully up and running would be to connect to
	// bootstrap peers (or any other peers). We leave this commented as
	// this is an example and the peer will die as soon as it finishes, so
	// it is unnecessary to put strain on the network.

	var DefaultBootstrapPeers []multiaddr.Multiaddr
	for _, s := range []string{
		//"/dns4/public.baasze.com/tcp/4001/p2p/bafzm3jqbebirvt7q5o2o35kgxr2vskstfvxhzztukp7xfhgd3vjkfexaznujy",
		//"/ip4/127.0.0.1/tcp/9000/p2p/bafzm3jqbedqdpe5giegij4oiiczwupkk4cqlh3ufjvx3tzwchbhshilm3p7ci",
	} {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			panic(err)
		}
		DefaultBootstrapPeers = append(DefaultBootstrapPeers, ma)
	}

	// This connects to public bootstrappers
	//for _, addr := range dht.DefaultBootstrapPeers {
	for _, addr := range DefaultBootstrapPeers {
		pi, _ := peer.AddrInfoFromP2pAddr(addr)
		// We ignore errors as some bootstrap peers may be down
		// and that is fine.
		err := bh.Connect(ctx, *pi)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("success connect %s\n", addr.String())
		}
	}

	fmt.Printf("peers: %v\n", bh.Peerstore().Peers())

	var pubsubOptions []pubsub.Option
	pubsubOptions = append(
		pubsubOptions,
		pubsub.WithMessageSigning(true),
		pubsub.WithStrictSignatureVerification(false),
	)
	ps, err := pubsub.NewGossipSub(ctx, bh, pubsubOptions...)
	if err != nil {
		panic(err)
	}

	sub, err := ps.Subscribe("fuck")

	for {
		got, err := sub.Next(ctx)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Println(got.GetData())
		}

		select {
		case <-ctx.Done():
			break
		default:
		}
	}

}
