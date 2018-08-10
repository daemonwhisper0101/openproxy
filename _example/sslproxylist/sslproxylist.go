// vim:set sw=2 sts=2:
package main

import (
  "fmt"
  "net/http"
  "net/url"
  "os"

  "github.com/daemonwhisper0101/openproxy"
)

func simple() []openproxy.OpenProxy {
  proxies, err := openproxy.GetSSLProxies()
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  return proxies
}

func withproxy(proxy string) []openproxy.OpenProxy {
  u, err := url.Parse(proxy)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  tr := &http.Transport{ Proxy: http.ProxyURL(u) }
  proxies, err := openproxy.GetSSLProxies(tr)
  if err != nil {
    fmt.Println(err)
    os.Exit(1)
  }
  return proxies
}

func main() {
  var proxies []openproxy.OpenProxy
  if len(os.Args) > 1 {
    proxies = withproxy(os.Args[1])
  } else {
    proxies = simple()
  }
  for _, proxy := range proxies {
    fmt.Printf("%v\n", &proxy)
  }
}
