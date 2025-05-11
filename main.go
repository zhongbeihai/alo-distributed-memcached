package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/alo-distributed-memcached/pkg"
)

var db = map[string]string{
	"Tom":  "630",
	"Jack": "589",
	"Sam":  "567",
}

func main() {
	pkg.NewGroup("scores", 2<<10, pkg.GetterFunc(
		func(key string) ([]byte, error) {
			log.Println("[SlowDB] search key", key)
			if v, ok := db[key]; ok {
				return []byte(v), nil
			}
			return nil, fmt.Errorf("%s not exist", key)
		}))

	addr := "localhost:9999"
	peers := pkg.NewHTTPPool(addr)
	log.Println("alo-distributed-cache is running at", addr)
	log.Fatal(http.ListenAndServe(addr, peers))
}
