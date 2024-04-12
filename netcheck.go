package main

import (
	"log"
	"sync"
  "fmt"
  
  "github.com/BirkhoffLee/netcheck/checks"
)

func checkNameserver(wg *sync.WaitGroup, address string) {
  defer wg.Done()
  
  addr := fmt.Sprintf("%s:53", address)

  // if err := checkNameserverAvailability(addr); err != nil {
  //   log.Printf("[!] Remote DNS server %s is unavailable: %s", address, err)
  //   return
  // }

  // log.Printf("[+] Remote DNS server %s is available", address)
  
  icmpStats, err := checks.CheckReachabilityWithICMP(addr)

  if err != nil {
    // ICMP unreachable

    if err := checks.CheckNameserverAvailability(addr); err != nil {
      log.Printf("[!] Nameserver %s: ICMP unreachable, DNS unreachable: %s", address, err)
    } else {
      log.Printf("[!] Nameserver %s: ICMP unreachable, DNS reachable", address)
    }
  } else {
    // ICMP reachable

    if err := checks.CheckNameserverAvailability(addr); err != nil {
      log.Printf("[!] Nameserver %s: ICMP %s, DNS unreachable: %s", address, icmpStats, err)
    } else {
      log.Printf("[+] Nameserver %s: ICMP %s, DNS reachable", address, icmpStats)
    }
  }
}

func main() {
  // Layer 3
	if gw, stats, err := checks.GetDefaultRoute(); err == nil {
		log.Printf("[+] Default route: %s (min/avg/max %s)", gw.String(), stats)
	} else {
		log.Printf("[!] Default route: %s (unreachable: %s)", gw.String(), err)
	}

  // DNS
	if ns, err := checks.GetDefaultNameserver(); err == nil {
    icmpStats, err := checks.CheckReachabilityWithICMP(ns)

		if err != nil {
      // ICMP unreachable

      if err := checks.CheckNameserverAvailability(ns + ":53"); err != nil {
        log.Printf("[!] Default nameserver: %s (ICMP unreachable, DNS unreachable: %s)", ns, err)
      } else {
        log.Printf("[!] Default nameserver: %s (ICMP unreachable, DNS reachable)", ns)
      }
		} else {
      // ICMP reachable

      if err := checks.CheckNameserverAvailability(ns + ":53"); err != nil {
        log.Printf("[!] Default nameserver: %s (ICMP %s, DNS unreachable: %s)", ns, icmpStats, err)
      } else {
        log.Printf("[+] Default nameserver: %s (ICMP %s, DNS reachable)", ns, icmpStats)
      }
    }
	} else {
		log.Printf("[!] Error reading default nameserver from system: %s", err)
	}

  // Tailscale
	log.Println(checks.GetTailscaleStatus())

	var wg sync.WaitGroup
	wg.Add(4)
  
  go checkNameserver(&wg, "1.1.1.1")
  go checkNameserver(&wg, "1.1.1.2")
  go checkNameserver(&wg, "8.8.8.8")
  go checkNameserver(&wg, "8.8.4.4")

	wg.Wait()

  // Layer 7
	if url, err := checks.GetClashApiBaseUrl(); err == nil {
    log.Printf("[+] Internet is handled by Clash (API live at %s)", url)
  } else if checks.IsRandomDomainResolvedToClashAddressSpace() {
    log.Printf("[+] Clash detected (random domain resolved to Clash address space)")
  } else {
    log.Printf("[!] Clash not detected: %s", err)
  }
  
	if err := checks.CheckCaptivePortal(); err != nil {
    log.Printf("[!] Captive Portal test failed: %s", err)
  } else {
    log.Printf("[+] Captive Portal test succeeded")
  }

	if pop, err := checks.GetCloudflareEdgeTrace(); err != nil {
    log.Printf("[!] Request to cloudflare.com failed: %s", err)
  } else {
		log.Printf("[+] Request to cloudflare.com hit %s edge", pop)
	}
}
