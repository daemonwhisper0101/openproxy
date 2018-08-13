// vim:set sw=2 sts=2:
package main

import (
  "fmt"
  "os"
  "os/signal"
  "syscall"
  "time"

  "github.com/daemonwhisper0101/openproxy/proxydb"
)

func main() {
  signal_chan := make(chan os.Signal)
  signal.Notify(signal_chan, syscall.SIGINT, syscall.SIGTERM)

  db := proxydb.New(proxydb.SSL)
  db.Update()
  db.Start()

loop:
  for {
    select {
    case <-signal_chan: break loop
    case <-time.After(time.Minute): db.ShowAll()
    }
  }
  fmt.Println("STOP!")
  // try to get 10 proxies
  for i := 0; i < 10; i++ {
    p := db.GetProxy()
    fmt.Println(p)
  }
  db.Stop()
  db.ShowAll()
  //
  fmt.Println("make all bad")
  for _, p := range db.Proxies {
    p.Bad()
  }
  db.ShowAll()
  fmt.Println("Done")
}
