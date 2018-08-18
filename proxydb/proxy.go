// vim:set sw=2 sts=2:
package proxydb

import (
  "net"
  "net/http"
  "time"

  "github.com/daemonwhisper0101/openproxy"
)

const (
  interval = time.Minute * 10 // 10 min.
)

type Proxy struct {
  p interface{}
  checktime time.Time
  live uint64
}

func (p *Proxy)OpenProxy() *openproxy.OpenProxy {
  switch proxy := p.p.(type) {
  case openproxy.OpenProxy:
    return &proxy
  }
  return nil
}

func (p *Proxy)IsAlive() bool {
  return (p.live & 1) == 1
}

func (p *Proxy)Check(url string) {
  now := time.Now()
  if now.Before(p.checktime.Add(interval)) {
    return // wait interval
  }
  p.checktime = now
  var live uint64 = 0
  switch proxy := p.p.(type) {
  case openproxy.OpenProxy:
    live = checkOpenProxy(proxy, url)
  default:
  }
  p.live <<= 1
  p.live |= live
}

func (p *Proxy)Bad() {
  now := time.Now()
  p.checktime = now
  p.live &= ^uint64(1) // drop the last bit
}

func checkOpenProxy(p openproxy.OpenProxy, url string) uint64 {
  d := &net.Dialer{ Timeout: time.Second * 10, KeepAlive: time.Second }
  tr := &http.Transport{
    Proxy: http.ProxyURL(p.URL()),
    DialContext: d.DialContext,
    DisableKeepAlives: true,
  }
  cl := &http.Client{ Transport: tr, Timeout: time.Second * 10 }
  resp, err := cl.Get(url)
  if err != nil {
    return 0
  }
  defer resp.Body.Close()
  //
  return 1
}
