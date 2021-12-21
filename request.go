// Copyright 2011 Andy Balholm. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Reading and parsing of ICAP requests.

// Package icap provides an extensible ICAP server.
package icap

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/textproto"
	"net/url"
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
	Preview    []byte               // the body data for an ICAP preview

	// The HTTP messages.
	Request  *http.Request
	Response *http.Response
}

// ReadRequest reads and parses a request from b.
func ReadRequest(b *bufio.ReadWriter) (req *Request, err error) {

	fmt.Printf("Inside ReadRequest\n")

	var buffer bytes.Buffer
	for {
		var p []byte
		if b.Reader.Buffered() > 0 {
			fmt.Printf("bytes to be read: %s\n", b.Reader.Buffered())
			p = make([]byte, b.Reader.Buffered())
		}
		_, err := b.Reader.Read(p)
		if err != nil {
			fmt.Printf("error occured while reading %s\n", err)
			if err == io.EOF {
				fmt.Printf("End of File")
				break
			}
			break
		}
		buffer.Write(p)
	}

	fmt.Printf("buffer alltogether is: %s\n", buffer.String())

	return nil, errors.New("a normal error when you do not know what you doing")
}

// An emptyReader is an io.ReadCloser that always returns os.EOF.
type emptyReader byte

func (emptyReader) Read(p []byte) (n int, err error) {
	return 0, io.EOF
}

func (emptyReader) Close() error {
	return nil
}

// A continueReader sends a "100 Continue" message the first time Read
// is called, creates a ChunkedReader, and reads from that.
type continueReader struct {
	buf *bufio.ReadWriter // the underlying connection
	cr  io.Reader         // the ChunkedReader
}

func (c *continueReader) Read(p []byte) (n int, err error) {
	if c.cr == nil {
		_, err := c.buf.WriteString("ICAP/1.0 100 Continue\r\n\r\n")
		if err != nil {
			return 0, err
		}
		err = c.buf.Flush()
		if err != nil {
			return 0, err
		}
		c.cr = newChunkedReader(c.buf)
	}

	return c.cr.Read(p)
}
