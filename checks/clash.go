package checks

import (
  "encoding/json"
  "fmt"
  "net/http"
  "strings"
  "regexp"
  "net"
)

type ClashAPIRootResponse struct {
	Hello string `json:"hello,omitempty"` // clash
	Message string `json:"message,omitempty"` // Unauthorized
}

func lookupFirstIp4Address(domain string) (string, error) {
  // lookup first ipv4 address of a domain
  addrs, err := net.LookupHost(domain)
  
  if err != nil {
    return "", err
  }
  
  for _, addr := range addrs {
    if ip := net.ParseIP(addr); ip.To4() != nil {
      return addr, nil
    }
  }
  
  return "", fmt.Errorf("no ipv4 address found")
}

func IsRandomDomainResolvedToClashAddressSpace() (bool) {
  addr, err := lookupFirstIp4Address("this.domain.does.not.exist.")
  
  if err != nil {
    return false
  }
  
  return strings.HasPrefix(addr, "198.18.")
}

func GetClashApiBaseUrl() (string, error) {
  // Get Clash endpoint from either default gateway or system proxy
  addr := ""

  if proxy, err := getSystemDefaultProxy(); err == nil {
    // get ip address from proxy string
    re := regexp.MustCompile(`(?m)\/\/(.*?)(:\d+)?$`)
    match := re.FindStringSubmatch(proxy)
    
    if len(match) < 2 {
      return "", fmt.Errorf("invalid proxy address")
    }
    
    addr = match[1]
  } else if gateway, _, err := GetDefaultRoute(); err == nil {
    addr = gateway.String()
  }

  // Construct Clash API URL
  url := fmt.Sprintf("http://%s:9090", addr)

  // Check if Clash API is reachable
  resp, err := http.Get(url)
  
  if err != nil {
    return "", err
  }
  
  // Check response
  var clashAPIRootResponse ClashAPIRootResponse
  decoder := json.NewDecoder(resp.Body)
  err = decoder.Decode(&clashAPIRootResponse)
  
  if err != nil {
    return "", err
  }
  
  if clashAPIRootResponse.Hello == "clash" {
    return url, nil
  }
  
  if clashAPIRootResponse.Message == "Unauthorized" {
    return url, nil
  }

  return "", fmt.Errorf("clash api not detected")
}

// func GetClashApiVersion() (string, error) {
//   url, err := getClashApiBaseUrl()
  
//   if err != nil {
//     return "", err
//   }

//   resp, err := http.Get(fmt.Sprintf("%s/version", url))

//   if err != nil {
//     return "", err
//   }

//   defer resp.Body.Close()

//   // {"premium":true,"version":"2023.08.17-11-g0f901d0"}
//   bodyBytes, err := io.ReadAll(resp.Body)

//   if err != nil {
//     return "", err
//   }

//   body := string(bodyBytes)
//   re := regexp.MustCompile(`"version":"(.*?)"`)
//   match := re.FindStringSubmatch(body)

//   if len(match) < 2 {
//     return "", fmt.Errorf("clash version not found in api response")
//   }

//   return match[1], nil
// }
