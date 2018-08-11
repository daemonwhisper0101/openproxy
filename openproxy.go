// vim:set sw=2 sts=2:
package openproxy

import (
  "io/ioutil"
  "net"
  "net/http"
  "strings"
  "fmt"
  "time"
)

type base struct {
  Host, Port, Code string
}

func (b *base)HostPort() string {
  return fmt.Sprintf("%s:%s", b.Host, b.Port)
}

type OpenProxy struct {
  base
  Anon string
}

func (p *OpenProxy)String() string {
  return fmt.Sprintf("%s %s %s", p.HostPort(), p.Code, p.Anon)
}

func NewOpenProxy(host, port, code, anon string) *OpenProxy {
  return &OpenProxy{ base: base { Host: host, Port: port, Code: code }, Anon: anon }
}

type SocksProxy struct {
  base
  Socks string
}

func (p *SocksProxy)String() string {
  return fmt.Sprintf("%s %s %s", p.HostPort(), p.Code, p.Socks)
}

func NewSocksProxy(host, port, code, socks string) *SocksProxy {
  return &SocksProxy{ base: base { Host: host, Port: port, Code: code }, Socks: socks }
}

func getHTML(url string, opts []interface{}) ([]byte, error) {
  cl := &http.Client{}
  for _, opt := range opts {
    switch v := opt.(type) {
    case http.Client: cl = &v
    case *http.Client: cl = v
    case http.Transport: cl.Transport = &v
    case *http.Transport: cl.Transport = v
    case time.Duration: cl.Timeout = v
    default: // unknown
    }
  }
  resp, err := cl.Get(url)
  if err != nil {
    return nil, fmt.Errorf("GET %s error %v\n", url, err)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return nil, fmt.Errorf("ReadAll error %v\n", err)
  }
  return body, nil
}

func GetSSLProxies(opts ...interface{}) ([]OpenProxy, error) {
  proxies := []OpenProxy{}
  url := "https://www.sslproxies.org/"
  body, err := getHTML(url, opts)
  if err != nil {
    return proxies, err
  }
  for _, val := range strings.Split(string(body), "<tr>") {
    if strings.Index(val, "<tfoot>") != -1 {
      if len(proxies) > 0 {
	break
      }
    }
    a := strings.Split(val, "</td><td")
    ip := a[0][4:]
    netip := net.ParseIP(ip)
    if netip == nil {
      continue
    }
    port := a[1][1:]
    code := a[2][1:]
    anon := a[4][1:]
    if anon != "elite proxy" && anon != "anonymous" {
      continue
    }
    proxy := NewOpenProxy(netip.String(), port, code, anon)
    proxies = append(proxies, *proxy)
  }
  return proxies, nil
}

func GetSocksProxies(opts ...interface{}) ([]SocksProxy, error) {
  proxies := []SocksProxy{}
  url := "https://www.socks-proxy.net"
  body, err := getHTML(url, opts)
  if err != nil {
    return proxies, err
  }
  for _, val := range strings.Split(string(body), "<tr>") {
    if strings.Index(val, "<tfoot>") != -1 {
      if len(proxies) > 0 {
	break
      }
    }
    a := strings.Split(val, "</td><td")
    ip := a[0][4:]
    netip := net.ParseIP(ip)
    if netip == nil {
      continue
    }
    port := a[1][1:]
    code := a[2][1:]
    socks := a[4][1:]
    if socks != "Socks4" && socks != "Socks5" {
      continue
    }
    proxy := NewSocksProxy(netip.String(), port, code, socks)
    proxies = append(proxies, *proxy)
  }
  return proxies, nil
}
