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

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
)

func boolPtrMaintenance(b bool) *bool {
	return &b
}

func setDataFlowSuspended(df *dataflowv1.DataFlow, suspended bool) {
	if df.Spec.Maintenance == nil {
		df.Spec.Maintenance = &dataflowv1.MaintenanceSpec{}
	}
	df.Spec.Maintenance.Suspended = boolPtrMaintenance(suspended)
}

func (h *APIHandler) stopDataFlow(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var df dataflowv1.DataFlow
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &df); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlow for stop")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setDataFlowSuspended(&df, true)

	if err := h.server.client.Update(r.Context(), &df); err != nil {
		h.server.logger.Error(err, "Failed to stop DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "stopping"})
}

func (h *APIHandler) startDataFlow(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var df dataflowv1.DataFlow
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &df); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlow for start")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	setDataFlowSuspended(&df, false)

	if err := h.server.client.Update(r.Context(), &df); err != nil {
		h.server.logger.Error(err, "Failed to start DataFlow")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "starting"})
}

func (h *APIHandler) getMaintenanceStatus(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var df dataflowv1.DataFlow
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &df); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlow maintenance status")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	status := map[string]interface{}{
		"maintenanceConfigured": df.Spec.Maintenance != nil &&
			(df.Spec.Maintenance.StartTime != "" || dataflowv1.IsMaintenanceSuspended(df.Spec.Maintenance)),
		"inMaintenance":       false,
		"nextMaintenanceTime": nil,
		"lastMaintenanceTime": nil,
		"suspended":           dataflowv1.IsMaintenanceSuspended(df.Spec.Maintenance),
	}

	if df.Status.MaintenanceStatus != nil {
		ms := df.Status.MaintenanceStatus
		status["inMaintenance"] = ms.InMaintenance
		status["nextMaintenanceTime"] = ms.NextMaintenanceTime
		status["lastMaintenanceTime"] = ms.LastMaintenanceTime
		status["suspended"] = ms.Suspended
	}

	_ = json.NewEncoder(w).Encode(status)
}

func (h *APIHandler) stopAllDataFlows(w http.ResponseWriter, r *http.Request, namespace string) {
	var list dataflowv1.DataFlowList
	if err := h.server.client.List(r.Context(), &list, client.InNamespace(namespace)); err != nil {
		h.server.logger.Error(err, "Failed to list DataFlows for stop-all")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	stopped := 0
	failed := 0
	for i := range list.Items {
		df := &list.Items[i]
		setDataFlowSuspended(df, true)
		if err := h.server.client.Update(r.Context(), df); err != nil {
			h.server.logger.Error(err, "Failed to stop DataFlow", "name", df.Name)
			failed++
		} else {
			stopped++
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]int{"stopped": stopped, "failed": failed})
}

func (h *APIHandler) startAllDataFlows(w http.ResponseWriter, r *http.Request, namespace string) {
	var list dataflowv1.DataFlowList
	if err := h.server.client.List(r.Context(), &list, client.InNamespace(namespace)); err != nil {
		h.server.logger.Error(err, "Failed to list DataFlows for start-all")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	started := 0
	failed := 0
	for i := range list.Items {
		df := &list.Items[i]
		if !dataflowv1.IsMaintenanceSuspended(df.Spec.Maintenance) {
			continue
		}
		setDataFlowSuspended(df, false)
		if err := h.server.client.Update(r.Context(), df); err != nil {
			h.server.logger.Error(err, "Failed to start DataFlow", "name", df.Name)
			failed++
		} else {
			started++
		}
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]int{"started": started, "failed": failed})
}
