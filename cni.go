package main

import (
	"bytes"
	"io"
	"mime"
	"net/http"

	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	v1 "go.opentelemetry.io/proto/otlp/common/v1"
	otlpresv1 "go.opentelemetry.io/proto/otlp/resource/v1"
	"google.golang.org/protobuf/proto"
)

type signal interface {
	proto.Message
	GetResources() []*otlpresv1.Resource
}

func newCommonNameInjector[T signal](target string) http.HandlerFunc {
	cni := new(commonNameInjector[T])
	cni.target = target
	cni.cl = &http.Client{
		Transport: &compressionRoundTripper{
			Proxied: http.DefaultTransport,
		},
	}
	return cni.handle
}

type commonNameInjector[T signal] struct {
	target string
	cl     *http.Client
}

func (c *commonNameInjector[T]) handle(w http.ResponseWriter, r *http.Request) {
	if mt, _, err := mime.ParseMediaType(r.Header.Get("Content-Type")); err != nil || mt != "application/x-protobuf" {
		http.Error(w, "unsupported content-type", http.StatusBadRequest)
		return
	}

	if len(r.TLS.PeerCertificates) == 0 {
		http.Error(w, "No client certificate provided", http.StatusForbidden)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	if err = r.Body.Close(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// TODO: remove this shitty workaround...
	var otelRes signal
	switch any(*new(T)).(type) {
	case *ExportMetricsServiceRequest:
		otelRes = &ExportMetricsServiceRequest{ExportMetricsServiceRequest: &metricspb.ExportMetricsServiceRequest{}}
	case *ExportTraceServiceRequest:
		otelRes = &ExportTraceServiceRequest{ExportTraceServiceRequest: &tracepb.ExportTraceServiceRequest{}}
	case *ExportLogsServiceRequest:
		otelRes = &ExportLogsServiceRequest{ExportLogsServiceRequest: &logspb.ExportLogsServiceRequest{}}
	default:
		http.Error(w, "unsupported request type", http.StatusBadRequest)
		return
	}
	if err = proto.Unmarshal(body, otelRes); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	for _, res := range otelRes.GetResources() {
		// TODO: overwrite attribute value if key already exists.
		res.Attributes = append(res.Attributes, &v1.KeyValue{
			Key: "tls.cert.common.name",
			Value: &v1.AnyValue{
				Value: &v1.AnyValue_StringValue{
					StringValue: r.TLS.PeerCertificates[0].Subject.CommonName,
				},
			},
		})
	}

	b, err := proto.Marshal(otelRes)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	newReq, err := http.NewRequestWithContext(r.Context(), r.Method, c.target+r.URL.Path, bytes.NewReader(b))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	newReq.Header.Add("Content-Type", r.Header.Get("Content-Type"))
	newReq.Header.Set("Content-Encoding", "gzip")

	resp, err := c.cl.Do(newReq)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	newBody, err := io.ReadAll(resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	if _, err := w.Write(newBody); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.WriteHeader(resp.StatusCode)
}
