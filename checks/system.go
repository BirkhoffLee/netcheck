package checks

import (
	"fmt"
	"net"
	"os"
	"regexp"

	"github.com/jackpal/gateway"
)

func getSystemDefaultProxy() (string, error) {
  http_proxy := os.Getenv("http_proxy")
  https_proxy := os.Getenv("https_proxy")
  all_proxy := os.Getenv("all_proxy")
  
  if all_proxy != "" {
    return all_proxy, nil
  }
  
  if http_proxy != "" {
    return http_proxy, nil
  }
  
  if https_proxy != "" {
    return https_proxy, nil
  }
  
  return "", fmt.Errorf("no proxy detected")
}

func GetDefaultRoute() (net.IP, string, error) {
	// check default gateway
	gw, err := gateway.DiscoverGateway()

	if err != nil {
		return nil, "", fmt.Errorf("error reading default route: %s", err)
	}

  stats, err := CheckReachabilityWithICMP(gw.String())

	if err != nil {
    return gw, stats, fmt.Errorf("default route is unreachable: %s", err)
	}

  return gw, stats, nil
}

func GetDefaultNameserver() (string, error) {
	// get default ns from /etc/resolv.conf
	byteString, err := os.ReadFile("/etc/resolv.conf")

	if err != nil {
		return "", err
	}

	s := string(byteString)

	re := regexp.MustCompile(`(?m)^nameserver( *|\t*)(.*?)$`)
	match := re.FindStringSubmatch(s)

	if len(match) < 2 {
    return "", fmt.Errorf("match is less than 2")
	}

	return match[2], nil
}
