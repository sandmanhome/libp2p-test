package main

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"sort"
	"time"

	"github.com/ipfs/go-datastore"
	syncds "github.com/ipfs/go-datastore/sync"
	config "github.com/ipfs/go-ipfs-config"
	"github.com/ipfs/go-ipfs/core"
	"github.com/ipfs/go-ipfs/core/node/libp2p"
	"github.com/ipfs/go-ipfs/repo"
	libp2p1 "github.com/libp2p/go-libp2p"
	relay "github.com/libp2p/go-libp2p-circuit"
	host1 "github.com/libp2p/go-libp2p-core/host"
	peer "github.com/libp2p/go-libp2p-peer"
	peerstore "github.com/libp2p/go-libp2p-peerstore"
)

// printSwarmAddrs prints the addresses of the host
func printSwarmAddrs(node *core.IpfsNode) {
	if !node.IsOnline {
		fmt.Println("Swarm not listening, running in offline mode.")
		return
	}

	var lisAddrs []string
	ifaceAddrs, err := node.PeerHost.Network().InterfaceListenAddresses()
	if err != nil {
		fmt.Printf("failed to read listening addresses: %s", err)
	}
	for _, addr := range ifaceAddrs {
		lisAddrs = append(lisAddrs, addr.String())
	}
	sort.Strings(lisAddrs)
	for _, addr := range lisAddrs {
		fmt.Printf("Swarm listening on %s\n", addr)
	}

	var addrs []string
	for _, addr := range node.PeerHost.Addrs() {
		addrs = append(addrs, addr.String())
	}
	sort.Strings(addrs)
	for _, addr := range addrs {
		fmt.Printf("Swarm announcing %s\n", addr)
	}

}

// isolates the complex initialization steps
func constructPeerHost(ctx context.Context, id peer.ID, ps peerstore.Peerstore, options ...libp2p1.Option) (host1.Host, error) {
	pkey := ps.PrivKey(id)
	if pkey == nil {
		return nil, fmt.Errorf("missing private key for node ID: %s", id.Pretty())
	}
	relayOpts := []relay.RelayOpt{}
	options = append([]libp2p1.Option{libp2p1.Identity(pkey), libp2p1.Peerstore(ps), libp2p1.EnableRelay(relayOpts...), libp2p1.EnableAutoRelay(), libp2p1.DefaultStaticRelays()}, options...)
	return libp2p1.New(ctx, options...)
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	ds := syncds.MutexWrap(datastore.NewMapDatastore())
	cfg, err := config.Init(ioutil.Discard, 2048)
	if err != nil {
		fmt.Println(err)
		return
	}

	cfg.Datastore = config.Datastore{}

	nd, err := core.NewNode(ctx, &core.BuildCfg{
		Permanent: true,
		Online:    true,
		Routing:   libp2p.DHTOption,
		ExtraOpts: map[string]bool{
			"pubsub": true,
		},
		Repo: &repo.Mock{
			C: *cfg,
			D: ds,
		},
		Host: constructPeerHost,
	})

	if err != nil {
		log.Fatal(err)
	}

	nd.IsDaemon = true

	fmt.Println(nd.PeerHost.Addrs())
	fmt.Println(nd.Identity.Pretty())
	fmt.Println(nd.Peerstore.Peers())
	printSwarmAddrs(nd)

	pubsub := nd.PubSub
	// sub, err := pubsub.Subscribe("fuck")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	ticker := time.NewTicker(time.Second * 1)
	for {
		// got, err := sub.Next(ctx)
		// if err != nil {
		// 	fmt.Println(err)
		// } else {
		// 	fmt.Println(got.GetData())
		// }
		select {
		case <-ticker.C:
			fmt.Println(nd.PeerHost.Addrs())
			pubsub.Publish("fuck", []byte("fuck fuck"))
		}

	}

	select {}

	// cctx := commands.Context{
	// 	Online: true,
	// 	ConstructNode: func() (*core.IpfsNode, error) {
	// 		return nd, nil
	// 	},
	// }

	// list, err := net.Listen("tcp", ":0")
	// if err != nil {
	// 	log.Fatal(err)
	// }

	// log.Println("listening on: ", list.Addr())

	// if err := corehttp.Serve(nd, list, corehttp.CommandsOption(cctx)); err != nil {
	// 	log.Fatal(err)
	// }
}
