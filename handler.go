package proxies

import (
	"errors"
	"log/slog"
	"net/http"
	"net/url"
	"sync"
)

type (
	Backend struct {
		Slug      string
		upstreams map[string]*Upstream
		mu        sync.RWMutex
	}

	ReverseRule struct {
		Host       string // eg. www.google.com
		RequestURI string // eg. /search?q=反向代理&newwindow=1
	}
)

func NewBackend(slug string) *Backend {
	return &Backend{
		Slug:      slug,
		upstreams: make(map[string]*Upstream),
	}
}

func (b *Backend) Handle(w http.ResponseWriter, in *http.Request) {
	target, err := b.lookupHost(in)
	if err != nil {
		slog.Error("bad request", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	b.mu.RLock()
	upstream, ok := b.upstreams[target.Host]
	b.mu.RUnlock()
	if !ok {
		newUpstream := NewUpstream(b.Slug+target.Host, target.Host)
		b.mu.Lock()
		upstream, ok = b.upstreams[target.Host]
		if !ok {
			b.upstreams[target.Host] = newUpstream
		}
		b.mu.Unlock()
		upstream = newUpstream
	}

	upstream.ServeHTTP(w, in)
}

// /proxy/google.com/search?q=proxy&newwindow=1
// /proxy/google.com
// /proxy/google.com/
// /proxy/
// /proxy
func (b *Backend) lookupHost(in *http.Request) (*url.URL, error) {
	if len(in.RequestURI) <= len(b.Slug) {
		return nil, errors.New("no target url provided")
	}
	target := "https://" + in.URL.RequestURI()[len(b.Slug):]
	t, err := url.ParseRequestURI(target)
	if err != nil {
		return nil, err
	}
	return t, nil
}
