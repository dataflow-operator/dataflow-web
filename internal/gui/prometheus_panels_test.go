/*
Copyright 2024.
*/

package gui

import (
	"strings"
	"testing"
)

func TestPanelQuery(t *testing.T) {
	q, err := panelQuery("throughput", "ns1", "df1")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(q, `namespace="ns1"`) || !strings.Contains(q, `name="df1"`) {
		t.Errorf("unexpected query: %s", q)
	}
}

func TestIsValidPanel(t *testing.T) {
	if !isValidPanel("throughput") {
		t.Error("throughput should be valid")
	}
	if isValidPanel("drop_database") {
		t.Error("arbitrary panel should be invalid")
	}
}
