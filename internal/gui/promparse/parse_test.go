/*
Copyright 2024.
*/

package promparse

import (
	"strings"
	"testing"
)

func TestParseTextAndFilter(t *testing.T) {
	input := `# HELP dataflow_messages_received_total Messages received
# TYPE dataflow_messages_received_total counter
dataflow_messages_received_total{namespace="default",name="flow1",source_type="kafka"} 42
dataflow_messages_received_total{namespace="other",name="flow2",source_type="kafka"} 10
other_metric{namespace="default",name="flow1"} 1
`
	samples, err := ParseText(strings.NewReader(input))
	if err != nil {
		t.Fatal(err)
	}
	if len(samples) != 2 {
		t.Fatalf("expected 2 dataflow samples, got %d", len(samples))
	}
	filtered := FilterByDataFlow(samples, "default", "flow1")
	if len(filtered) != 1 {
		t.Fatalf("expected 1 filtered sample, got %d", len(filtered))
	}
	if filtered[0].Value != 42 {
		t.Errorf("expected value 42, got %v", filtered[0].Value)
	}
}
