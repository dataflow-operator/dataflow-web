/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
*/

package gui

import (
	"fmt"
	"strings"
)

// panelQuery builds a PromQL query for a whitelisted panel id.
func panelQuery(panel, namespace, name string) (string, error) {
	selector := fmt.Sprintf(`{namespace=%q,name=%q}`, namespace, name)
	switch panel {
	case "throughput":
		return fmt.Sprintf(`sum(rate(dataflow_messages_received_total%s[5m])) by (source_type)`, selector), nil
	case "sink_rate":
		return fmt.Sprintf(`sum(rate(dataflow_messages_sent_total%s[5m])) by (sink_type, route)`, selector), nil
	case "e2e_latency_p95":
		return fmt.Sprintf(`histogram_quantile(0.95, sum(rate(dataflow_task_end_to_end_latency_seconds_bucket%s[5m])) by (le))`, selector), nil
	case "connector_errors":
		return fmt.Sprintf(`sum(rate(dataflow_connector_errors_total%s[5m])) by (connector_type, connector_name, operation, error_type)`, selector), nil
	case "queue_size":
		return fmt.Sprintf(`dataflow_task_queue_size%s`, selector), nil
	case "connector_health":
		return fmt.Sprintf(`dataflow_connector_connection_status%s`, selector), nil
	case "processing_latency_p95":
		return fmt.Sprintf(`histogram_quantile(0.95, sum(rate(dataflow_processing_duration_seconds_bucket%s[5m])) by (le))`, selector), nil
	case "success_rate":
		return fmt.Sprintf(`dataflow_task_success_rate%s * 100`, selector), nil
	case "throughput_gauge":
		return fmt.Sprintf(`dataflow_task_throughput_messages_per_second%s`, selector), nil
	default:
		return "", fmt.Errorf("unknown panel %q", panel)
	}
}

var validPanels = map[string]struct{}{
	"throughput":             {},
	"sink_rate":              {},
	"e2e_latency_p95":        {},
	"connector_errors":       {},
	"queue_size":             {},
	"connector_health":       {},
	"processing_latency_p95": {},
	"success_rate":           {},
	"throughput_gauge":       {},
}

func isValidPanel(panel string) bool {
	_, ok := validPanels[panel]
	return ok
}

// instantMetricNames returns metric names used for instant fallback per panel.
func instantMetricNames(panel string) []string {
	switch panel {
	case "throughput":
		return []string{"dataflow_messages_received_total", "dataflow_task_throughput_messages_per_second"}
	case "sink_rate":
		return []string{"dataflow_messages_sent_total"}
	case "e2e_latency_p95", "processing_latency_p95":
		return []string{"dataflow_task_end_to_end_latency_seconds", "dataflow_processing_duration_seconds"}
	case "connector_errors":
		return []string{"dataflow_connector_errors_total"}
	case "queue_size":
		return []string{"dataflow_task_queue_size"}
	case "connector_health":
		return []string{"dataflow_connector_connection_status", "dataflow_connector_source_poll_healthy"}
	case "success_rate":
		return []string{"dataflow_task_success_rate"}
	case "throughput_gauge":
		return []string{"dataflow_task_throughput_messages_per_second"}
	default:
		return nil
	}
}

func sampleMatchesPanel(panel string, sampleName string) bool {
	for _, n := range instantMetricNames(panel) {
		if sampleName == n || strings.HasPrefix(sampleName, n+"_") {
			return true
		}
	}
	return false
}
