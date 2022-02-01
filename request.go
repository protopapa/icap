// Copyright 2011 Andy Balholm. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Reading and parsing of ICAP requests.

// Package icap provides an extensible ICAP server.
package icap

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
	"strings"
)

type badStringError struct {
	what string
	str  string
}

func (e *badStringError) Error() string { return fmt.Sprintf("%s %q", e.what, e.str) }

// A Request represents a parsed ICAP request.
type Request struct {
	Method     string               // REQMOD, RESPMOD, OPTIONS, etc.
	RawURL     string               // The URL given in the request.
	URL        *url.URL             // Parsed URL.
	Proto      string               // The protocol version.
	Header     textproto.MIMEHeader // The ICAP header
	RemoteAddr string               // the address of the computer sending the request
	//	Preview    []byte               // the body data for an ICAP preview

	// The HTTP messages.
	Request  *http.Request
	Response *http.Response
}

// ReadRequest reads and parses a request from b.
func ReadRequest(b *bufio.ReadWriter) (req *Request, err error) {
	tp := textproto.NewReader(b.Reader)
	req = new(Request)

	// Read first line.
	var s string
	s, err = tp.ReadLine()
	if err != nil {
		if err == io.EOF {
			err = io.ErrUnexpectedEOF
		}
		return nil, err
	}

	f := strings.SplitN(s, " ", 3)
	if len(f) < 3 {
		return nil, &badStringError{"malformed ICAP request", s}
	}
	req.Method, req.RawURL, req.Proto = f[0], f[1], f[2]
	Std.Printf("req.Method: %s, req.RawURL: %s, req.Proto: %s\n", req.Method, req.RawURL, req.Proto)

	req.URL, err = url.ParseRequestURI("/filter")
	if err != nil {
		return nil, err
	}

	Std.Printf("req.URL: %s\n", req.URL)

	req.Header, err = tp.ReadMIMEHeader()
	Std.Printf("req.Header: %+v\n", req.Header)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Request: %+v\n", req)
	return
}
