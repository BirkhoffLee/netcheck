package checks

import (
	"fmt"
	"time"

	"github.com/go-ping/ping"
)

func CheckReachabilityWithICMP(host string) (string, error) {
	pinger, err := ping.NewPinger(host)
	if err != nil {
		return "", err
	}

	pinger.Count = 3
	pinger.Debug = true
	pinger.Interval = 200 * time.Millisecond
	pinger.Timeout = 2 * time.Second

	err = pinger.Run()

	if err != nil {
		return "", err
	}

	stats := pinger.Statistics()

  statsString := fmt.Sprintf(
    "%s/%s/%s",
    stats.MinRtt.Round(time.Millisecond),
    stats.AvgRtt.Round(time.Millisecond),
    stats.MaxRtt.Round(time.Millisecond),
  )

	return statsString, nil
}
