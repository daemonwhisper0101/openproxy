// vim:set sw=2 sts=2:
package proxydb

import (
  "fmt"
  "sync"
  "time"

  "github.com/daemonwhisper0101/openproxy"
)

type Flags int64
const (
  SSL	= 0x1
  Socks = 0x2
)

type cache struct {
  created time.Time
  proxies []*Proxy
  index int
  maxsize int
}

type FilterFunc func(p interface{}) bool

type DB struct {
  Proxies map[string]*Proxy
  inactive map[string]*Proxy
  flags Flags
  filter FilterFunc
  running bool
  signal, done chan bool
  livecache cache
}

func New(flags Flags, opts ...interface{}) *DB {
  if (flags & (SSL | Socks)) == 0 {
    return nil
  }
  var filter FilterFunc = nil
  for _, opt := range opts {
    switch v := opt.(type) {
    case FilterFunc: filter = v
    }
  }
  return &DB{
    Proxies: map[string]*Proxy{},
    inactive: map[string]*Proxy{},
    flags: flags,
    filter: filter,
    running: false,
    signal: make(chan bool),
    done: make(chan bool),
    livecache: cache{
      created: time.Now(),
      proxies: []*Proxy{},
      maxsize: 128,
    },
  }
}

func (db *DB)Update(opts ...interface{}) {
  t := time.Now().Add(-time.Hour)
  if (db.flags & SSL) != 0 {
    proxies, _ := openproxy.GetSSLProxies(opts...)
    for _, p := range proxies {
      if db.filter != nil && db.filter(p) == false {
	continue
      }
      key := p.HostPort()
      _, ok := db.Proxies[key]
      if !ok {
	db.Proxies[key] = &Proxy{ p: p, checktime: t, live: ^uint64(1) }
      }
    }
  }
  if (db.flags & Socks) != 0 {
    proxies, _ := openproxy.GetSocksProxies(opts...)
    for _, p := range proxies {
      if db.filter != nil && db.filter(p) == false {
	continue
      }
      key := p.HostPort()
      _, ok := db.Proxies[key]
      if !ok {
	db.Proxies[key] = &Proxy{ p: p, checktime: t, live: ^uint64(1) }
      }
    }
  }
}

func (db *DB)Start() {
  if db.running {
    return
  }
  db.running = true
  // checking goroutine
  go func() {
    // 16 workers
    var wg sync.WaitGroup
    queue := make(chan *Proxy, 16)
    stop := false
    for i := 0; i < 16; i++ {
      wg.Add(1)
      go func() {
	defer wg.Done()
	for {
	  p, ok := <-queue
	  if !ok || stop {
	    return
	  }
	  p.Check("https://www.google.com")
	  time.Sleep(time.Second) // interval
	}
      }()
    }
loop:
    for {
      inact := map[string]*Proxy{}
      for k, p := range db.Proxies {
	select {
	case <-db.signal: break loop
	default:
	}
	queue <- p
	if (p.live & 0xf) == 0 {
	  inact[k] = p
	}
      }
      for k, p := range inact {
	delete(db.Proxies, k)
	db.inactive[k] = p
      }
      select {
      case <-db.signal: break loop
      case <-time.After(time.Minute):
      }
    }
    stop = true
    close(queue)
    wg.Wait()
    db.done <- true
  }()
}

func (db *DB)Stop() {
  if !db.running {
    return
  }
  db.signal <- true // stop
  db.running = false
  <-db.done // wait
}

func (db *DB)ShowAll() {
  for k, p := range db.Proxies {
    fmt.Printf("%s %016x\n", k, p.live)
  }
}

func (db *DB)updateCache() {
  now := time.Now()
  inv := db.livecache.created.Add(time.Minute * 10)
  if now.After(inv) || len(db.livecache.proxies) == 0 {
    // create cache
    proxies := []*Proxy{}
    n := 0
    for _, p := range db.Proxies {
      if (p.live & 1) == 0 {
	continue
      }
      pp := p // dereference
      proxies = append(proxies, pp)
      n++
      if n > db.livecache.maxsize {
	break
      }
    }
    db.livecache.created = now
    db.livecache.proxies = proxies
    db.livecache.index = 0
  }
}

func (db *DB)GetProxy() *Proxy {
  db.updateCache()
  for db.livecache.index < len(db.livecache.proxies) {
    p := db.livecache.proxies[db.livecache.index]
    db.livecache.index++
    if (p.live & 1) != 0 {
      return p
    }
  }
  // no proxies
  return nil
}
