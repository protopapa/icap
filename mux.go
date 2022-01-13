// Copyright 2011 Andy Balholm. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package icap

import (
	"fmt"
	"net/http"
	"net/url"
	"path"
	"strings"
)

// ServeMux is an ICAP request multiplexer.
// It matches the URL of each incoming request against a list of registered
// patterns and calls the handler for the pattern that
// most closely matches the URL.
//
// For more details, see the documentation for http.ServeMux
type ServeMux struct {
	h Handler
}

// NewServeMux allocates and returns a new ServeMux.
func NewServeMux() *ServeMux { return &ServeMux{} }

// DefaultServeMux is the default ServeMux used by Serve.
var DefaultServeMux = NewServeMux()

// Does path match pattern?
func pathMatch(pattern, path string) bool {
	if len(pattern) == 0 {
		// should not happen
		return false
	}
	n := len(pattern)
	if pattern[n-1] != '/' {
		return pattern == path
	}
	return len(path) >= n && path[0:n] == pattern
}

// Return the canonical path for p, eliminating . and .. elements.
func cleanPath(p string) string {
	if p == "" {
		return "/"
	}
	if p[0] != '/' {
		p = "/" + p
	}
	np := path.Clean(p)
	// path.Clean removes trailing slash except for root;
	// put the trailing slash back if necessary.
	if p[len(p)-1] == '/' && np != "/" {
		np += "/"
	}
	return np
}

// Find a handler on a handler map given a path string
// Most-specific (longest) pattern wins
//func (mux *ServeMux) match(path string) Handler {
//	var h Handler
//	var n = 0
//	for k, v := range mux.m {
//		if !pathMatch(k, path) {
//			continue
//		}
//		if h == nil || len(k) > n {
//			n = len(k)
//			h = v
//		}
//	}
//	return h
//}

// ServeICAP dispatches the request to the handler whose
// pattern most closely matches the request URL.
func (mux *ServeMux) ServeICAP(w ResponseWriter, r *Request) {
	// Clean path to canonical form and redirect.
	if p := cleanPath(r.URL.Path); p != r.URL.Path {
		w.Header().Set("Location", p)
		w.WriteHeader(http.StatusMovedPermanently, nil, false)
		return
	}
	// Host-specific pattern takes precedence over generic ones
	h := mux.h
	h.ServeICAP(w, r)
}

// Handle registers the handler for the given pattern.
func (mux *ServeMux) Handle(handler Handler) {
	fmt.Printf("rregister")
	mux.h = handler
}

// HandleFunc registers the handler function for the given pattern.
func (mux *ServeMux) HandleFunc(handler func(ResponseWriter, *Request)) {
	mux.Handle(HandlerFunc(handler))
}

// Handle registers the handler for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func Handle(pattern string, handler Handler) { DefaultServeMux.Handle(handler) }

// HandleFunc registers the handler function for the given pattern
// in the DefaultServeMux.
// The documentation for ServeMux explains how patterns are matched.
func HandleFunc(handler func(ResponseWriter, *Request)) {
	DefaultServeMux.HandleFunc(handler)
}

// NotFound replies to the request with an HTTP 404 not found error.
func NotFound(w ResponseWriter, r *Request) {
	w.WriteHeader(http.StatusNotFound, nil, false)
}

// NotFoundHandler returns a simple request handler
// that replies to each request with a ``404 page not found'' reply.
func NotFoundHandler() Handler { return HandlerFunc(NotFound) }

// Redirect to a fixed URL
type redirectHandler struct {
	url  string
	code int
}

func (rh *redirectHandler) ServeICAP(w ResponseWriter, r *Request) {
	Redirect(w, r, rh.url, rh.code)
}

// RedirectHandler returns a request handler that redirects
// each request it receives to the given url using the given
// status code.
func RedirectHandler(url_ string, code int) Handler {
	return &redirectHandler{url_, code}
}

// Redirect replies to the request with a redirect to url,
// which may be a path relative to the request path.
func Redirect(w ResponseWriter, r *Request, url_ string, code int) {
	if u, err := url.Parse(url_); err == nil {
		// If url was relative, make absolute by
		// combining with request path.
		// The browser would probably do this for us,
		// but doing it ourselves is more reliable.
		oldpath := r.URL.Path
		if oldpath == "" { // should not happen, but avoid a crash if it does
			oldpath = "/"
		}
		if u.Scheme == "" {
			// no leading icap://server
			if url_ == "" || url_[0] != '/' {
				// make relative path absolute
				olddir, _ := path.Split(oldpath)
				url_ = olddir + url_
			}

			var query string
			if i := strings.Index(url_, "?"); i != -1 {
				url_, query = url_[:i], url_[i:]
			}

			// clean up but preserve trailing slash
			trailing := url_[len(url_)-1] == '/'
			url_ = path.Clean(url_)
			if trailing && url_[len(url_)-1] != '/' {
				url_ += "/"
			}
			url_ += query
		}
	}

	w.Header().Set("Location", url_)
	w.WriteHeader(code, nil, false)
}
