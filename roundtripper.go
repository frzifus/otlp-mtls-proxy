package main

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
)

type compressionRoundTripper struct {
	Proxied http.RoundTripper
}

func (c *compressionRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	encodings := req.Header["Content-Encoding"]
	if len(encodings) > 0 {
		encoding := encodings[0]
		var compressedBody io.ReadCloser
		var err error

		// Compress based on the encoding
		switch encoding {
		case "gzip":
			var buf bytes.Buffer
			gzipWriter := gzip.NewWriter(&buf)
			_, err = io.Copy(gzipWriter, req.Body)
			if err != nil {
				return nil, err
			}
			err = gzipWriter.Close()
			if err != nil {
				return nil, err
			}
			compressedBody = io.NopCloser(&buf)
			req.Body = compressedBody
			req.ContentLength = int64(buf.Len())
			req.Header.Set("Content-Encoding", "gzip")
		default:
			return nil, &http.ProtocolError{ErrorString: "Unsupported Content-Encoding"}
		}
	}

	return c.Proxied.RoundTrip(req)
}
