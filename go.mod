module server

go 1.14

require (
	github.com/libp2p/go-conn-security-multistream v0.2.0
	github.com/libp2p/go-libp2p v0.8.3
	github.com/libp2p/go-libp2p-core v0.5.3
	github.com/libp2p/go-libp2p-peerstore v0.2.3
	github.com/libp2p/go-libp2p-pubsub v0.2.7
	github.com/libp2p/go-libp2p-secio v0.2.2
	github.com/libp2p/go-libp2p-swarm v0.2.3
	github.com/libp2p/go-libp2p-transport-upgrader v0.2.0
	github.com/libp2p/go-libp2p-yamux v0.2.7
	github.com/libp2p/go-stream-muxer-multistream v0.3.0
	github.com/libp2p/go-tcp-transport v0.2.0
	github.com/multiformats/go-multiaddr v0.2.1

)

replace (
	github.com/btcsuite/btcd v0.20.1-beta => gitlab.com/ihyperdata/icfs/btcd.git v0.1.0
	github.com/ipfs/go-ipfs-config v0.5.3 => gitlab.com/ihyperdata/icfs/go-ipfs-config.git v0.2.0
	github.com/ipfs/go-verifcid v0.0.1 => gitlab.com/ihyperdata/icfs/go-verifcid.git v0.0.0-20200616091530-02d96dec2b07
	github.com/libp2p/go-libp2p v0.8.3 => gitlab.com/ihyperdata/icfs/go-libp2p.git v0.1.0
	github.com/libp2p/go-libp2p-core v0.5.3 => gitlab.com/ihyperdata/icfs/go-libp2p-core.git v0.2.0
	github.com/multiformats/go-multiaddr v0.2.1 => gitlab.com/ihyperdata/icfs/go-multiaddr.git v0.1.0
	github.com/multiformats/go-multihash v0.0.13 => gitlab.com/ihyperdata/icfs/go-multihash.git v0.1.0
)
