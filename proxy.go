package proxies

import (
	"errors"
	"log/slog"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
)

type (
	MultipleReverseProxy struct {
		slug      string
		upstreams map[string]*ReverseProxy
		mu        sync.RWMutex
	}
	ReverseProxy struct {
		Host   string
		Prefix string
		Proxy  *httputil.ReverseProxy
	}
)

func NewMultipleReverseProxy() *MultipleReverseProxy {
	return &MultipleReverseProxy{
		slug:      "/proxy/",
		upstreams: make(map[string]*ReverseProxy),
	}
}

func (mp *MultipleReverseProxy) Endpoint() string {
	return mp.slug
}

func (mp *MultipleReverseProxy) ServeHTTP(w http.ResponseWriter, in *http.Request) {
	target, err := mp.lookupHost(in)
	if err != nil {
		slog.Error("bad request", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(err.Error()))
		return
	}

	mp.mu.RLock()
	upstream, ok := mp.upstreams[target.Host]
	mp.mu.RUnlock()
	if !ok {
		newUpstream := NewReverseProxy(mp.slug+target.Host, target.Host)
		mp.mu.Lock()
		upstream, ok = mp.upstreams[target.Host]
		if !ok {
			mp.upstreams[target.Host] = newUpstream
		}
		mp.mu.Unlock()
		upstream = newUpstream
	}

	upstream.ServeHTTP(w, in)
}

// /proxy/google.com/search?q=proxy&newwindow=1
// /proxy/google.com
// /proxy/google.com/
// /proxy/
// /proxy
func (mp *MultipleReverseProxy) lookupHost(in *http.Request) (*url.URL, error) {
	if len(in.RequestURI) <= len(mp.slug) {
		return nil, errors.New("no target url provided")
	}
	target := "https://" + in.URL.RequestURI()[len(mp.slug):]
	t, err := url.ParseRequestURI(target)
	if err != nil {
		return nil, err
	}
	return t, nil
}

func NewReverseProxy(prefix, host string) *ReverseProxy {
	rp := &ReverseProxy{
		Prefix: prefix,
		Host:   host,
	}
	rp.Proxy = &httputil.ReverseProxy{Director: rp.Director}
	return rp
}

func (rp *ReverseProxy) Director(out *http.Request) {
	out.URL.Scheme = "https" // 只支持反代 HTTPS 网站
	out.URL.Host = rp.Host
	out.Host = rp.Host
	out.URL.Path = out.URL.Path[len(rp.Prefix):]
}

func (rp *ReverseProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	rp.Proxy.ServeHTTP(w, r)
}
