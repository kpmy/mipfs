package ip2p_test

import (
	"context"
	"fmt"
	"github.com/ipfs/go-ipfs-util"
	"github.com/ipfs/go-libp2p-peer"
	"github.com/ipfs/go-libp2p-peerstore"
	"github.com/jbenet/go-multiaddr"
	"log"
	"testing"
	"time"

	"net"
)

func TestLink(t *testing.T) {
	addr, _ := multiaddr.NewMultiaddr("/ip4/0.0.0.0/tcp/7001")
	ps := peerstore.NewPeerstore()
	id := peer.ID(util.Hash([]byte(fmt.Sprint(time.Now().UnixNano()))))
	ctx := context.Background()
	sw, _ := swarm.NewNetwork(ctx, []multiaddr.Multiaddr{addr}, id, ps, nil)
	h := basichost.New(sw)
	h.SetStreamHandler("/echo/0.0.1", func(s net.Stream) {
		log.Println("new stream")
	})
	time.Sleep(1 * time.Second)
}
