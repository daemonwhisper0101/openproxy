// vim:set sw=2 sts=2:
package proxydb

import (
  "net"
  "net/http"
  "time"

  "github.com/daemonwhisper0101/openproxy"
)

func checkOpenProxy(p openproxy.OpenProxy, url string) uint64 {
  d := &net.Dialer{ Timeout: time.Second * 10, KeepAlive: time.Second }
  tr := &http.Transport{
    Proxy: http.ProxyURL(p.URL()),
    DialContext: d.DialContext,
    TLSHandshakeTimeout: time.Second * 5,
    DisableKeepAlives: true,
    IdleConnTimeout: time.Second,
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
