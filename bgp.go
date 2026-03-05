// https://github.com/osrg/gobgp/blob/master/docs/sources/lib.md
package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/netip"
	"time"

	"github.com/osrg/gobgp/v4/api"
	"github.com/osrg/gobgp/v4/pkg/apiutil"
	"github.com/osrg/gobgp/v4/pkg/packet/bgp"
	"github.com/osrg/gobgp/v4/pkg/server"
	log "github.com/sirupsen/logrus"
)

var (
	v4Family = &api.Family{Afi: api.Family_AFI_IP, Safi: api.Family_SAFI_UNICAST} // &gobgpapi.Family literal is not a constant
)

type BgpServer struct {
	server *server.BgpServer
}

func initBgpServer(routerId string, asn uint32, listenPort int32) (*BgpServer, error) {

	log := slog.Default()
	lvl := &slog.LevelVar{}
	lvl.Set(slog.LevelDebug)

	s := server.NewBgpServer(server.LoggerOption(log, lvl))
	go s.Serve()

	// global configuration
	if err := s.StartBgp(context.Background(), &api.StartBgpRequest{
		Global: &api.Global{
			Asn:        asn,
			RouterId:   routerId,
			ListenPort: listenPort,
		},
	}); err != nil {
		return nil, err
	}

	// set default import policy
	s.SetPolicyAssignment(context.Background(), &api.SetPolicyAssignmentRequest{
		Assignment: &api.PolicyAssignment{
			Direction:     api.PolicyDirection_POLICY_DIRECTION_IMPORT,
			DefaultAction: api.RouteAction_ROUTE_ACTION_REJECT,
		},
	})

	// monitor the change of the peer state
	if err := s.WatchEvent(context.Background(), server.WatchEventMessageCallbacks{
		OnPeerUpdate: func(peer *apiutil.WatchEventMessage_PeerEvent, _ time.Time) {
			if peer.Type == apiutil.PEER_EVENT_STATE {
				log.Info("peer state changed", slog.Any("Peer", peer.Peer))
			}
		}}, server.WatchPeer()); err != nil {
		log.Error("failed to watch event", slog.String("Error", err.Error()))
		return nil, err
	}

	return &BgpServer{server: s}, nil
}

func (bs *BgpServer) AddPeer(address string, asn uint32) error {
	n := &api.Peer{
		Conf: &api.PeerConf{
			NeighborAddress: address,
			PeerAsn:         asn,
		},
	}
	return bs.server.AddPeer(context.Background(), &api.AddPeerRequest{Peer: n})
}

func (bs *BgpServer) AddV4Path(prefix string, prefixLen int, nextHop string, asn uint32) error {
	path := fmt.Sprintf("%s/%d", prefix, prefixLen)
	nlri, _ := bgp.NewIPAddrPrefix(netip.MustParsePrefix(path))
	a1 := bgp.NewPathAttributeOrigin(0) // the prefix originates from an interior routing protocol (IGP)
	a2, _ := bgp.NewPathAttributeNextHop(netip.MustParseAddr(nextHop))
	a3 := bgp.NewPathAttributeAsPath([]bgp.AsPathParamInterface{bgp.NewAs4PathParam(bgp.BGP_ASPATH_ATTR_TYPE_SEQ, []uint32{asn})})
	attrs := []bgp.PathAttributeInterface{a1, a2, a3}

	log.Info("Adding Path",
		slog.String("path", path),
		slog.String("next hop", nextHop),
	)
	_, err := bs.server.AddPath(apiutil.AddPathRequest{Paths: []*apiutil.Path{{
		Nlri:  nlri,
		Attrs: attrs,
	}}})
	if err != nil {
		return err
	}
	setBGPPathAdvertisementMetric(prefix, fmt.Sprint(prefixLen), nextHop)
	return nil
}

func (bs *BgpServer) DeleteV4Path(prefix string, prefixLen int, nextHop string, asn uint32) error {
	path := fmt.Sprintf("%s/%d", prefix, prefixLen)
	nlri, _ := bgp.NewIPAddrPrefix(netip.MustParsePrefix(path))
	a1 := bgp.NewPathAttributeOrigin(0) // the prefix originates from an interior routing protocol (IGP)
	a2, _ := bgp.NewPathAttributeNextHop(netip.MustParseAddr(nextHop))
	a3 := bgp.NewPathAttributeAsPath([]bgp.AsPathParamInterface{bgp.NewAs4PathParam(bgp.BGP_ASPATH_ATTR_TYPE_SEQ, []uint32{asn})})
	attrs := []bgp.PathAttributeInterface{a1, a2, a3}

	err := bs.server.DeletePath(apiutil.DeletePathRequest{Paths: []*apiutil.Path{{
		Nlri:  nlri,
		Attrs: attrs,
	}}})
	if err != nil {
		return err
	}
	unsetBGPPathAdvertisementMetric(prefix, fmt.Sprint(prefixLen), nextHop)
	return nil
}

func (bs *BgpServer) ListV4Paths() {
	bs.server.ListPath(apiutil.ListPathRequest{
		TableType: api.TableType_TABLE_TYPE_GLOBAL,
	}, func(prefix bgp.NLRI, paths []*apiutil.Path) {
		log.Info(prefix.String())
		for _, p := range paths {
			log.Info("path",
				slog.Uint64("peer_asn", uint64(p.PeerASN)),
				slog.String("peer_address", p.PeerAddress.String()),
				slog.Uint64("age", uint64(p.Age)),
				slog.Bool("best", p.Best),
			)
		}
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
