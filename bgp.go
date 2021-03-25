package main

import (
	"context"
	"fmt"

	"github.com/golang/protobuf/ptypes"
	"github.com/golang/protobuf/ptypes/any"
	api "github.com/osrg/gobgp/api"
	gobgp "github.com/osrg/gobgp/pkg/server"
	log "github.com/sirupsen/logrus"
)

var (
	v4Family = &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST} // &gobgpapi.Family literal is not a constant
)

type BgpServer struct {
	server *gobgp.BgpServer
}

func initBgpServer(routerId string, as uint32, listenPort int32) (*BgpServer, error) {
	s := gobgp.NewBgpServer()
	go s.Serve()

	// global configuration
	if err := s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			As:         as,
			RouterId:   routerId,
			ListenPort: listenPort,
		},
	}); err != nil {
		return nil, err
	}

	// monitor the change of the peer state
	if err := s.MonitorPeer(context.Background(), &api.MonitorPeerRequest{}, func(p *api.Peer) { log.Info(p) }); err != nil {
		log.Fatal(err)
	}

	return &BgpServer{server: s}, nil
}

func (bs *BgpServer) AddPeer(address string, as uint32) error {
	n := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: address,
			PeerAs:          as,
		},
	}
	return bs.server.AddPeer(context.Background(), &api.AddPeerRequest{Peer: n})
}

func (bs *BgpServer) AddV4Path(prefix string, prefixLen uint32, nextHop string) error {
	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    prefix,
		PrefixLen: prefixLen,
	})

	a1, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0, // the prefix originates from an interior routing protocol (IGP)
	})
	a2, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: nextHop,
	})
	attrs := []*any.Any{a1, a2}

	_, err := bs.server.AddPath(context.Background(), &api.AddPathRequest{
		Path: &api.Path{
			Family: v4Family,
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})
	if err != nil {
		return err
	}
	setBGPPathAdvertisementMetric(prefix, fmt.Sprint(prefixLen), nextHop)
	return nil
}

func (bs *BgpServer) DeleteV4Path(prefix string, prefixLen uint32, nextHop string) error {
	nlri, _ := ptypes.MarshalAny(&api.IPAddressPrefix{
		Prefix:    prefix,
		PrefixLen: prefixLen,
	})

	a1, _ := ptypes.MarshalAny(&api.OriginAttribute{
		Origin: 0, // the prefix originates from an interior routing protocol (IGP)
	})
	a2, _ := ptypes.MarshalAny(&api.NextHopAttribute{
		NextHop: nextHop,
	})
	attrs := []*any.Any{a1, a2}

	err := bs.server.DeletePath(context.Background(), &api.DeletePathRequest{
		Path: &api.Path{
			Family: v4Family,
			Nlri:   nlri,
			Pattrs: attrs,
		},
	})
	if err != nil {
		return err
	}
	unsetBGPPathAdvertisementMetric(prefix, fmt.Sprint(prefixLen), nextHop)
	return nil
}

func (bs *BgpServer) ListV4Paths() {
	bs.server.ListPath(context.Background(), &api.ListPathRequest{Family: v4Family}, func(p *api.Destination) {
		log.Info(p)
	})
}

// bgpSetup starts the bgp server and adds the peers
func bgpSetup(bgpConfig bgpConfig) *BgpServer {
	// Start bgp server
	bgp, err := initBgpServer(
		bgpConfig.Local.RouterId,
		bgpConfig.Local.AS,
		bgpConfig.Local.ListenPort,
	)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Fatal("Cannot start bgp server")
	}
	// Add Peers
	for _, peer := range bgpConfig.Peers {
		if err := bgp.AddPeer(peer.Address, peer.AS); err != nil {
			log.WithFields(log.Fields{
				"error": err,
			}).Fatal("Cannot add bgpp peer")
		}
	}
	return bgp
}
