# netcheck

A self-contained utility for checking internet connectivity.

## Usage

```shell
$ go install github.com/BirkhoffLee/netcheck
...
$ netcheck
2023/05/05 15:43:04 Default route: 198.18.0.1 (reachable via ICMP)
2023/05/05 15:43:04 Default nameserver: 100.100.100.100 (ICMP reachable, DNS reachable)
2023/05/05 15:43:05 We're online on tsnet, exit node "clash" via 10.0.1.101:57321
2023/05/05 15:43:05 Remote DNS server 1.1.1.2 is available
2023/05/05 15:43:05 Remote DNS server 8.8.8.8 is available
2023/05/05 15:43:05 Remote DNS server 8.8.4.4 is available
2023/05/05 15:43:05 Remote DNS server 1.1.1.1 is available
2023/05/05 15:43:05 Request to cloudflare.com hit TW edge
```

## License

The project is licensed under [the Unlicense](https://unlicense.org/).
