package main

import (
	"log/slog"
	"net/http"

	"github.com/jacexh/proxies"
)

func main() {
	backend := proxies.NewBackend("/proxy/")
	http.HandleFunc(backend.Slug, backend.Handle)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		slog.Error("something was wrong", slog.String("error", err.Error()))
	}
}
