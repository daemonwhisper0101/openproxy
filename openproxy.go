// vim:set sw=2 sts=2:
package openproxy

import (
  "io/ioutil"
  "net/http"
  "strings"
  "fmt"
)

type OpenProxy struct {
  Host, Port, Code string
}

func (p *OpenProxy)String() string {
  return fmt.Sprintf("%s:%s %s", p.Host, p.Port, p.Code)
}

func (p *OpenProxy)HostPort() string {
  return fmt.Sprintf("%s:%s", p.Host, p.Port)
}

func GetSSLProxies() ([]OpenProxy, error) {
  proxies := []OpenProxy{}
  url := "https://www.sslproxies.org/"
  resp, err := http.Get(url)
  if err != nil {
    return proxies, fmt.Errorf("GET %s error %v\n", url, err)
  }
  defer resp.Body.Close()
  body, err := ioutil.ReadAll(resp.Body)
  if err != nil {
    return proxies, fmt.Errorf("ReadAll error %v\n", err)
  }
  for _, val := range strings.Split(string(body), "<tr>") {
    if strings.Index(val, "<tfoot>") != -1 {
      if len(proxies) > 0 {
	break
      }
    }
    if strings.Index(val, "elite") == -1 {
      continue
    }
    a := strings.Split(val, "</td><td")
    ip := a[0][4:]
    port := a[1][1:]
    code := a[2][1:]
    proxy := OpenProxy{ Host: ip, Port: port, Code: code }
    proxies = append(proxies, proxy)
  }
  return proxies, nil
}
