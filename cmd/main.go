package main

import (
	"log/slog"
	"net/http"

	"github.com/jacexh/proxies"
)

func main() {
	slog.Info("starting reverse proxy...")
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("hello world"))
	})

	proxies := proxies.NewMultipleReverseProxy()
	http.Handle(proxies.Endpoint(), proxies)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("something was wrong", slog.String("error", err.Error()))
	}
}
