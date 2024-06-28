package main

import (
	"crypto/tls"
	"crypto/x509"
	"log"
	"net/http"
	"os"
)

func main() {
	// TODO: allow configuring endpoint.
	const target = "http://localhost:8318"
	caCert, err := os.ReadFile("tls/rootCA.crt")
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatalf("Failed to parse CA certificate: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/metrics", newCommonNameInjector[*ExportMetricsServiceRequest](target))
	mux.HandleFunc("POST /v1/traces", newCommonNameInjector[*ExportTraceServiceRequest](target))
	mux.HandleFunc("POST /v1/logs", newCommonNameInjector[*ExportLogsServiceRequest](target))

	srv := &http.Server{
		Addr: "localhost.localdomain:4318",
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  caCertPool,
		},
		Handler: mwDecompression(mux),
	}

	if err := srv.ListenAndServeTLS("tls/server.crt", "tls/server.key"); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
