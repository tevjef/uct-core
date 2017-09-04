package metrics

import (
	"fmt"
	"log"
	"net"
	"net/http"

	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func init() {
	go func() {
		listener, err := net.Listen("tcp", ":13000")
		if err != nil {
			listener, _ = net.Listen("tcp", ":0")
			fmt.Println("Using port:", listener.Addr().(*net.TCPAddr).Port)
		}

		http.Handle("/metrics", promhttp.Handler())
		log.Fatal(http.Serve(listener, nil))
	}()
}
