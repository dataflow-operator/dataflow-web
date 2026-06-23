/*
Copyright 2024.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package gui

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
)

func createTestDataFlow(server *Server, name string, suspended bool) error {
	spec := dataflowv1.DataFlowSpec{
		Source: dataflowv1.SourceSpec{Type: "kafka"},
		Sink:   dataflowv1.SinkSpec{Type: "postgresql"},
	}
	if suspended {
		b := true
		spec.Maintenance = &dataflowv1.MaintenanceSpec{Suspended: &b}
	}
	df := &dataflowv1.DataFlow{
		ObjectMeta: metav1.ObjectMeta{Name: name, Namespace: "default"},
		Spec:       spec,
	}
	return server.client.Create(ctx, df)
}

func TestAPIHandler_StopStartDataFlow(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	handler := NewAPIHandler(server)

	if err := createTestDataFlow(server, "flow-a", false); err != nil {
		t.Fatalf("create: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/dataflows/flow-a/stop?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("stop status %d: %s", w.Code, w.Body.String())
	}

	var got dataflowv1.DataFlow
	if err := server.client.Get(ctx, types.NamespacedName{Namespace: "default", Name: "flow-a"}, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Spec.Maintenance == nil || got.Spec.Maintenance.Suspended == nil || !*got.Spec.Maintenance.Suspended {
		t.Fatal("expected suspended=true")
	}

	req = httptest.NewRequest(http.MethodPost, "/dataflows/flow-a/start?namespace=default", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start status %d: %s", w.Code, w.Body.String())
	}

	if err := server.client.Get(ctx, types.NamespacedName{Namespace: "default", Name: "flow-a"}, &got); err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Spec.Maintenance == nil || got.Spec.Maintenance.Suspended == nil || *got.Spec.Maintenance.Suspended {
		t.Fatal("expected suspended=false")
	}
}

func TestAPIHandler_MaintenanceStatus(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	handler := NewAPIHandler(server)
	if err := createTestDataFlow(server, "flow-b", true); err != nil {
		t.Fatalf("create: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/dataflows/flow-b/maintenance?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status %d: %s", w.Code, w.Body.String())
	}

	var body map[string]interface{}
	if err := json.NewDecoder(w.Body).Decode(&body); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if suspended, ok := body["suspended"].(bool); !ok || !suspended {
		t.Fatalf("expected suspended true, got %#v", body["suspended"])
	}
}

func TestAPIHandler_StopAllStartAll(t *testing.T) {
	server, err := setupTestServer()
	if err != nil {
		t.Fatalf("setup: %v", err)
	}
	handler := NewAPIHandler(server)

	for _, name := range []string{"flow-1", "flow-2"} {
		if err := createTestDataFlow(server, name, false); err != nil {
			t.Fatalf("create %s: %v", name, err)
		}
	}

	req := httptest.NewRequest(http.MethodPost, "/dataflows/stop-all?namespace=default", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("stop-all status %d: %s", w.Code, w.Body.String())
	}

	var stopResult map[string]int
	if err := json.NewDecoder(w.Body).Decode(&stopResult); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if stopResult["stopped"] != 2 {
		t.Fatalf("expected stopped=2, got %#v", stopResult)
	}

	req = httptest.NewRequest(http.MethodPost, "/dataflows/start-all?namespace=default", nil)
	w = httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("start-all status %d: %s", w.Code, w.Body.String())
	}

	var startResult map[string]int
	if err := json.NewDecoder(w.Body).Decode(&startResult); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if startResult["started"] != 2 {
		t.Fatalf("expected started=2, got %#v", startResult)
	}
}