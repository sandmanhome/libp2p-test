package main

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/libp2p/go-libp2p"
	autonat "github.com/libp2p/go-libp2p-autonat-svc"
	connmgr "github.com/libp2p/go-libp2p-connmgr"
	"github.com/libp2p/go-libp2p-core/crypto"
	"github.com/libp2p/go-libp2p-core/host"
	"github.com/libp2p/go-libp2p-core/peer"
	discovery "github.com/libp2p/go-libp2p-discovery"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"go.uber.org/fx"

	// libp2pquic "github.com/libp2p/go-libp2p-quic-transport"
	//routing "github.com/libp2p/go-libp2p-core/routing"
	disc "github.com/libp2p/go-libp2p-discovery"
	routing "github.com/libp2p/go-libp2p-routing"
	secio "github.com/libp2p/go-libp2p-secio"
	libp2ptls "github.com/libp2p/go-libp2p-tls"
	"github.com/multiformats/go-multiaddr"
)

func TopicDiscovery() interface{} {
	return func(mctx context.Context, lc fx.Lifecycle, host host.Host, cr routing.IpfsRouting) (service discovery.Discovery, err error) {
		baseDisc := disc.NewRoutingDiscovery(cr)
		minBackoff, maxBackoff := time.Second*60, time.Hour
		rng := rand.New(rand.NewSource(rand.Int63()))
		d, err := disc.NewBackoffDiscovery(
			baseDisc,
			disc.NewExponentialBackoff(minBackoff, maxBackoff, disc.FullJitter, time.Second, 5.0, 0, rng),
		)

		if err != nil {
			return nil, err
		}

		return d, nil
	}
}

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

	var idht *dht.IpfsDHT

	psk := []byte{0x55, 0x48, 0xb7, 0xf2, 0xeb, 0xd9, 0x56, 0x80, 0x81, 0xab, 0x23, 0xa6, 0x1f, 0x15, 0xff, 0x68, 0x24, 0xe8, 0x35, 0xe5, 0xca, 0xc4, 0xc5, 0x1c, 0xaf, 0x27, 0xd8, 0xbc, 0x5c, 0x3e, 0x6c, 0x4d}

	h2, err := libp2p.New(ctx,
		// Use the keypair we generated
		libp2p.Identity(priv),
		// Multiple listen addresses
		libp2p.ListenAddrStrings(
			"/ip4/0.0.0.0/tcp/4001",      // regular tcp connections
			"/ip4/0.0.0.0/udp/4001/quic", // a UDP endpoint for the QUIC transport
		),
		// support TLS connections
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		// support secio connections
		libp2p.Security(secio.ID, secio.New),
		// support QUIC - experimental
		// libp2p.Transport(libp2pquic.NewTransport),
		// support any other default transports (TCP)
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
		// Let's prevent our peer from having too many
		// connections by attaching a connection manager.
		libp2p.ConnectionManager(connmgr.NewConnManager(
			100,         // Lowwater
			400,         // HighWater,
			time.Minute, // GracePeriod
		)),
		// Attempt to open ports using uPNP for NATed hosts.
		libp2p.NATPortMap(),
		// Let this host use the DHT to find other hosts
		libp2p.Routing(func(h host.Host) (routing.PeerRouting, error) {
			idht, err = dht.New(ctx, h)
			return idht, err
		}),
		// Let this host use relays and advertise itself on relays if
		// it finds it is behind NAT. Use libp2p.Relay(options...) to
		// enable active relays and more.
		libp2p.EnableAutoRelay(),
		libp2p.PrivateNetwork(psk),
	)
	if err != nil {
		panic(err)
	}

	// If you want to help other peers to figure out if they are behind
	// NATs, you can launch the server-side of AutoNAT too (AutoRelay
	// already runs the client)
	_, err = autonat.NewAutoNATService(ctx, h2,
		// Support same non default security and transport options as
		// original host.
		libp2p.Security(libp2ptls.ID, libp2ptls.New),
		libp2p.Security(secio.ID, secio.New),
		// libp2p.Transport(libp2pquic.NewTransport),
		libp2p.DefaultTransports,
		libp2p.DefaultMuxers,
	)

	fmt.Printf("Hello World, my second hosts ID is %s\n", h2.ID())

	// The last step to get fully up and running would be to connect to
	// bootstrap peers (or any other peers). We leave this commented as
	// this is an example and the peer will die as soon as it finishes, so
	// it is unnecessary to put strain on the network.

	var DefaultBootstrapPeers []multiaddr.Multiaddr
	for _, s := range []string{
		"/ip4/39.104.89.79/tcp/4001/p2p/bafzm3jqbec7ulhfmm7s7ydt2mf32nbsjy4237mvzj5skzbkxrfxz7axghsyum",
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
		err := h2.Connect(ctx, *pi)
		if err != nil {
			fmt.Println(err)
		} else {
			fmt.Printf("success connect %s\n", addr.String())
		}
	}

	fmt.Printf("peers: %v\n", h2.Peerstore().Peers())

	var pubsubOptions []pubsub.Option
	pubsubOptions = append(
		pubsubOptions,
		pubsub.WithMessageSigning(true),
		pubsub.WithStrictSignatureVerification(false),
	)
	ps, err := pubsub.NewGossipSub(ctx, h2, pubsubOptions...)
	if err != nil {
		panic(err)
	}

	sub, err := ps.Subscribe("fuck")
	if err != nil {
		fmt.Println(err)
	}

	for {
		fmt.Println("sub Next: ")
		got, err := sub.Next(ctx)
		if err != nil {
			fmt.Println(err)
		}

		fmt.Println(got.Data)

		select {
		case <-ctx.Done():
			break
		default:
		}
	}

}
