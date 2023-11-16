package proxies

import (
	"net/http"
	"net/http/httputil"
)

type (
	Upstream struct {
		Host   string
		Prefix string
		Proxy  *httputil.ReverseProxy
	}
)

func NewUpstream(prefix, host string) *Upstream {
	up := &Upstream{
		Prefix: prefix,
		Host:   host,
	}
	up.Proxy = &httputil.ReverseProxy{Director: up.Direct}
	return up
}

func (up *Upstream) Direct(out *http.Request) {
	out.URL.Scheme = "https"
	out.URL.Host = up.Host
	out.URL.Path = out.URL.Path[len(up.Prefix):]
}

func (up *Upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	r.Host = up.Host
	up.Proxy.ServeHTTP(w, r)
}
