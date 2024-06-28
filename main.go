package main

import (
	"crypto/tls"
	"crypto/x509"
	"flag"
	"log"
	"net/http"
	"os"
)

func main() {
	var (
		target   = flag.String("target", "http://localhost:8318", "endpoint")
		rootCA   = flag.String("root-ca", "tls/rootCA.crt", "path")
		certFile = flag.String("cert-file", "tls/server.crt", "path")
		keyFile  = flag.String("key-file", "tls/server.key", "path")
	)
	flag.Parse()
	// TODO: allow configuring endpoint.
	caCert, err := os.ReadFile(*rootCA)
	if err != nil {
		log.Fatalf("Failed to read CA certificate: %v", err)
	}
	caCertPool := x509.NewCertPool()
	if ok := caCertPool.AppendCertsFromPEM(caCert); !ok {
		log.Fatalf("Failed to parse CA certificate: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("POST /v1/metrics", newCommonNameInjector[*ExportMetricsServiceRequest](*target))
	mux.HandleFunc("POST /v1/traces", newCommonNameInjector[*ExportTraceServiceRequest](*target))
	mux.HandleFunc("POST /v1/logs", newCommonNameInjector[*ExportLogsServiceRequest](*target))

	srv := &http.Server{
		Addr: "0.0.0.0:4318",
		TLSConfig: &tls.Config{
			ClientAuth: tls.RequireAndVerifyClientCert,
			ClientCAs:  caCertPool,
		},
		Handler: mwDecompression(mux),
	}

	if err := srv.ListenAndServeTLS(*certFile, *keyFile); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}
}
