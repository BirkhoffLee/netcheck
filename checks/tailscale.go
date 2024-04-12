package checks

import (
	"context"
	"fmt"
	"time"

	"tailscale.com/client/tailscale"
	"tailscale.com/types/key"
	"tailscale.com/ipn/ipnstate"
)

// func findTailscalePeerByStableNodeID(peers map[key.NodePublic]*ipnstate.PeerStatus, id tailcfg.StableNodeID) *ipnstate.PeerStatus {
//   for _, p := range peers {
//     if p.ID == id {
//       return p
//     }
//   }

//   return nil
// }

func FindActiveExitNodeFromPeersMap(peers map[key.NodePublic]*ipnstate.PeerStatus) *ipnstate.PeerStatus {
	for _, p := range peers {
		if p.ExitNode {
			return p
		}
	}

	return nil
}

func GetTailscaleStatus() string {
	// check tailscale status
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	ts, err := tailscale.Status(ctx) // https://pkg.go.dev/tailscale.com@v1.40.0/ipn/ipnstate#Status
	defer cancel()

	if err != nil {
		return fmt.Sprintf("[!] Could not determine tailscaled status: %s", err)
	}

	// https://github.com/tailscale/tailscale/blob/9bdaece3d7c3c83aae01e0736ba54e833f4aea51/cmd/tailscale/cli/status.go#L162-L196

	if !ts.Self.Online {
		return fmt.Sprintf("[~] We're offline on tsnet: BackendState=%s", ts.BackendState)
	}

	exitNodeStatus := FindActiveExitNodeFromPeersMap(ts.Peer)

	if exitNodeStatus == nil {
		return "[+] We're online on tsnet; not using any exit node"
	}

	if exitNodeStatus.Active {
		if exitNodeStatus.Relay != "" && exitNodeStatus.CurAddr == "" {
			return fmt.Sprintf("[~] We're online on tsnet; exit node \"%s\" via relay %s", exitNodeStatus.HostName, exitNodeStatus.Relay)
		}

		if exitNodeStatus.CurAddr != "" {
			return fmt.Sprintf("[+] We're online on tsnet; exit node \"%s\" via %s", exitNodeStatus.HostName, exitNodeStatus.CurAddr)
		}

		return fmt.Sprintf("[!] We're online on tsnet; exit node \"%s\" (unknown connection)", exitNodeStatus.HostName)
	}

	return fmt.Sprintf("[+] We're online on tsnet; exit node \"%s\" is inactive", exitNodeStatus.HostName)
}
