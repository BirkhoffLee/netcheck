package main

import (
	"context"
	"fmt"
	"io"
	"io/fs"
	"log"
	"net"
	"net/http"
	"os"
	"regexp"
	"sync"
	"time"

	"github.com/go-ping/ping"
	"github.com/jackpal/gateway"
	"github.com/miekg/dns"

	"tailscale.com/client/tailscale"
	"tailscale.com/types/key"
	"tailscale.com/ipn/ipnstate"
)

func checkCaptivePortal() error {
	resp, err := http.Get("http://connectivitycheck.gstatic.com/generate_204")

	if err != nil || resp.StatusCode != 204 {
		log.Printf("Captive portal error: %s", err)
		return err
	}

	log.Printf("Captive portal ok")
	return nil
}

func getCloudflareEdgeTrace() (string, error) {
	resp, err := http.Get("https://www.cloudflare.com/cdn-cgi/trace")

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	body := string(bodyBytes)
	re := regexp.MustCompile(`loc=(.*?)\n`)
	match := re.FindStringSubmatch(body)

	if len(match) < 2 {
		return "", fmt.Errorf("could not determine edge pop")
	}

	return match[1], nil
}

func getDefaultNameserver() (string, error) {
	// get default ns from /etc/resolv.conf
	byteString, err := fs.ReadFile(os.DirFS("/etc"), "resolv.conf")

	if err != nil {
		log.Printf("Get default nameserver error: %s", err)
		return "", err
	}

	s := string(byteString)

	re := regexp.MustCompile(`(?m)^nameserver( *|\t*)(.*?)$`)
	match := re.FindStringSubmatch(s)

	if len(match) < 2 {
		panic("nameserver not found")
	}

	return match[2], nil
}

func checkReachabilityWithICMP(host string) bool {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return false
	}

	pinger.Count = 3
	pinger.Debug = true
	pinger.Interval = 200 * time.Millisecond
	pinger.Timeout = 3 * time.Second

	err = pinger.Run()

	if err != nil {
		return false
	}

	stats := pinger.Statistics()

	return stats.PacketsRecv != 0
}

func checkDefaultRoute() (net.IP, error) {
	// check default gateway
	gw, err := gateway.DiscoverGateway()

	if err != nil {
		return nil, fmt.Errorf("error reading default route: %s", err)
	}

	if checkReachabilityWithICMP(gw.String()) {
		return gw, nil
	}

	return gw, fmt.Errorf("default route is unreachable")
}

func checkNameserverAvailability(s string) error {
	c := new(dns.Client)
	c.Dialer = &net.Dialer{
		Timeout: 3 * time.Second,
	}
	m := new(dns.Msg)
	m.SetQuestion("apple.com.", dns.TypeA)
	_, _, err := c.Exchange(m, s)

	if err != nil {
		return err
	}

	// fmt.Printf("%v %v", in, rtt)
	return nil
}

// func findTailscalePeerByStableNodeID(peers map[key.NodePublic]*ipnstate.PeerStatus, id tailcfg.StableNodeID) *ipnstate.PeerStatus {
//   for _, p := range peers {
//     if p.ID == id {
//       return p
//     }
//   }

//   return nil
// }

func findActiveExitNodeFromPeersMap(peers map[key.NodePublic]*ipnstate.PeerStatus) *ipnstate.PeerStatus {
	for _, p := range peers {
		if p.ExitNode {
			return p
		}
	}

	return nil
}

func getTailscaleStatus() string {
	// check tailscale status
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	ts, err := tailscale.Status(ctx) // https://pkg.go.dev/tailscale.com@v1.40.0/ipn/ipnstate#Status
	defer cancel()

	if err != nil {
		return fmt.Sprintf("Could not determine tailscaled status: %s", err)
	}

	// https://github.com/tailscale/tailscale/blob/9bdaece3d7c3c83aae01e0736ba54e833f4aea51/cmd/tailscale/cli/status.go#L162-L196

	if !ts.Self.Online {
		return fmt.Sprintf("We're not connected to tsnet: BackendState = %s", ts.BackendState)
	}

	exitNodeStatus := findActiveExitNodeFromPeersMap(ts.Peer)

	if exitNodeStatus == nil {
		return "We're online on tsnet and not using any exit node"
	}

	if exitNodeStatus.Active {
		if exitNodeStatus.Relay != "" && exitNodeStatus.CurAddr == "" {
			return fmt.Sprintf("We're online on tsnet, exit node \"%s\" via relay %s", exitNodeStatus.HostName, exitNodeStatus.Relay)
		}

		if exitNodeStatus.CurAddr != "" {
			return fmt.Sprintf("We're online on tsnet, exit node \"%s\" via %s", exitNodeStatus.HostName, exitNodeStatus.CurAddr)
		}

		return fmt.Sprintf("We're online on tsnet, exit node \"%s\" (unknown connection)", exitNodeStatus.HostName)
	}

	return fmt.Sprintf("We're online on tsnet, exit node \"%s\" is inactive", exitNodeStatus.HostName)
}

func main() {
	if gw, err := checkDefaultRoute(); err == nil {
		log.Printf("Default route: %s (reachable via ICMP)", gw.String())
	} else {
		log.Printf("Default route: %s (unreachable via ICMP)", gw.String())
	}

	ns, err := getDefaultNameserver()

	if err == nil {
		a := ""

		if !checkReachabilityWithICMP(ns) {
			a = "un"
		}

		if err := checkNameserverAvailability(ns + ":53"); err != nil {
			log.Printf("Default nameserver: %s (ICMP %sreachable, DNS unreachable: %s)", ns, a, err)
		} else {
			log.Printf("Default nameserver: %s (ICMP %sreachable, DNS reachable)", ns, a)
		}
	} else {
		log.Printf("Error reading default nameserver from system: %s", err)
	}

	log.Println(getTailscaleStatus())

	var wg sync.WaitGroup
	wg.Add(4)

	go func() {
		defer wg.Done()

		if err := checkNameserverAvailability("1.1.1.1:53"); err != nil {
			log.Printf("Remote DNS server 1.1.1.1 is unavailable: %s", err)
		} else {
			log.Printf("Remote DNS server 1.1.1.1 is available")
		}
	}()

	go func() {
		defer wg.Done()

		if err := checkNameserverAvailability("1.1.1.2:53"); err != nil {
			log.Printf("Remote DNS server 1.1.1.2 is unavailable: %s", err)
		} else {
			log.Printf("Remote DNS server 1.1.1.2 is available")
		}
	}()

	go func() {
		defer wg.Done()

		if err := checkNameserverAvailability("8.8.8.8:53"); err != nil {
			log.Printf("Remote DNS server 8.8.8.8 is unavailable: %s", err)
		} else {
			log.Printf("Remote DNS server 8.8.8.8 is available")
		}
	}()

	go func() {
		defer wg.Done()

		if err := checkNameserverAvailability("8.8.4.4:53"); err != nil {
			log.Printf("Remote DNS server 8.8.4.4 is unavailable: %s", err)
		} else {
			log.Printf("Remote DNS server 8.8.4.4 is available")
		}
	}()

	wg.Wait()

	go checkCaptivePortal()

	if pop, err := getCloudflareEdgeTrace(); err == nil {
		log.Printf("Request to cloudflare.com hit %s edge", pop)
	} else {
		log.Printf("Request to cloudflare.com failed: %s", err)
	}
}
