package checks

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"regexp"
	"time"

	"github.com/miekg/dns"
)

func CheckNameserverAvailability(s string) error {
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

func CheckCaptivePortal() error {
	resp, err := http.Get("http://connectivitycheck.gstatic.com/generate_204")

	if err != nil || resp.StatusCode != 204 {
		return err
	}

	return nil
}

func GetCloudflareEdgeTrace() (string, error) {
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
