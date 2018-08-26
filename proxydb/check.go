// vim:set sw=2 sts=2:
package proxydb

import (
  "context"
  "net"
  "net/http"
  "time"

  "github.com/daemonwhisper0101/openproxy"
)

type disposableConn struct {
  conn net.Conn
  closed bool
}

func disposableDialContext(ctx context.Context, network, addr string) (net.Conn, error) {
  d := &net.Dialer{ Timeout: time.Second * 10, KeepAlive: time.Second }
  conn, err := d.DialContext(ctx, network, addr)
  if err != nil {
    return nil, err
  }
  // 30 secs alive
  go func() {
    time.Sleep(time.Second * 30)
    conn.Close()
  }()
  return conn, nil
}

func checkOpenProxy(p openproxy.OpenProxy, url string) uint64 {
  tr := &http.Transport{
    Proxy: http.ProxyURL(p.URL()),
    DialContext: disposableDialContext,
    TLSHandshakeTimeout: time.Second * 5,
    DisableKeepAlives: true,
    IdleConnTimeout: time.Second,
    MaxIdleConns: 1,
  }
  defer tr.CloseIdleConnections()
  cl := &http.Client{ Transport: tr, Timeout: time.Second * 10 }
  resp, err := cl.Get(url)
  if err != nil {
    return 0
  }
  defer resp.Body.Close()
  //
  return 1
}
