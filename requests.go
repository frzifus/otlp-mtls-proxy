package main

import (
	logspb "go.opentelemetry.io/proto/otlp/collector/logs/v1"
	metricspb "go.opentelemetry.io/proto/otlp/collector/metrics/v1"
	tracepb "go.opentelemetry.io/proto/otlp/collector/trace/v1"
	otlpresv1 "go.opentelemetry.io/proto/otlp/resource/v1"
)

type ExportMetricsServiceRequest struct {
	*metricspb.ExportMetricsServiceRequest
}

func (e *ExportMetricsServiceRequest) GetResources() []*otlpresv1.Resource {
	res := []*otlpresv1.Resource{}
	for _, l := range e.ExportMetricsServiceRequest.GetResourceMetrics() {
		if r := l.GetResource(); res != nil {
			res = append(res, r)
		}
	}
	return res
}

type ExportTraceServiceRequest struct {
	*tracepb.ExportTraceServiceRequest
}

func (e *ExportTraceServiceRequest) GetResources() []*otlpresv1.Resource {
	res := []*otlpresv1.Resource{}
	for _, l := range e.ExportTraceServiceRequest.GetResourceSpans() {
		if r := l.GetResource(); res != nil {
			res = append(res, r)
		}
	}
	return res
}

type ExportLogsServiceRequest struct {
	*logspb.ExportLogsServiceRequest
}

func (e *ExportLogsServiceRequest) GetResources() []*otlpresv1.Resource {
	res := []*otlpresv1.Resource{}
	for _, l := range e.ExportLogsServiceRequest.GetResourceLogs() {
		if r := l.GetResource(); res != nil {
			res = append(res, r)
		}
	}
	return res
}
