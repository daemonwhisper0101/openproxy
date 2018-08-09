// vim:set sw=2 sts=2:
package main

import (
  "fmt"

  "github.com/daemonwhisper0101/openproxy"
)

func main() {
  proxies, err := openproxy.GetSSLProxies()
  if err != nil {
    fmt.Println(err)
    return
  }
  for _, proxy := range proxies {
    fmt.Printf("%v\n", &proxy)
  }
}
