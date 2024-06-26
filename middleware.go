package main

import (
	"compress/gzip"
	"io"
	"net/http"
)

func mwDecompression(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		encodings := r.Header["Content-Encoding"]
		if len(encodings) > 0 {
			encoding := encodings[0]
			var reader io.ReadCloser
			var err error

			switch encoding {
			case "gzip":
				reader, err = gzip.NewReader(r.Body)
				if err != nil {
					http.Error(w, "Failed to decompress request body", http.StatusInternalServerError)
					return
				}
				defer reader.Close()
			default:
				http.Error(w, "Unsupported Content-Encoding", http.StatusUnsupportedMediaType)
				return
			}

			r.Body = reader
			r.Header.Del("Content-Encoding")
		}

		next.ServeHTTP(w, r)
	})
}
