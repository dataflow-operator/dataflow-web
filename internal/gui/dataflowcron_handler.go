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
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
	"github.com/dataflow-operator/dataflow/pkg/k8snames"
)

const (
	dataFlowCronOwnerLabel        = "dataflow.dataflow.io/dataflow-cron"
	dataFlowCronTriggerIndexLabel = "dataflow.dataflow.io/trigger-index"
	dataFlowCronProcessorStep     = "processor"
	cronJobInstantiateAnnotation  = "cronjob.kubernetes.io/instantiate"
)

func (h *APIHandler) handleDataFlowCrons(w http.ResponseWriter, r *http.Request, parts []string) {
	namespace := r.URL.Query().Get("namespace")
	if namespace == "" {
		namespace = "default"
	}

	if len(parts) >= 2 && parts[1] == "trigger" {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.triggerDataFlowCron(w, r, namespace, parts[0])
		return
	}
	if len(parts) >= 2 && parts[1] == "suspend" {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		h.suspendDataFlowCron(w, r, namespace, parts[0])
		return
	}

	switch r.Method {
	case http.MethodGet:
		if len(parts) == 0 {
			h.listDataFlowCrons(w, r, namespace)
		} else {
			h.getDataFlowCron(w, r, namespace, parts[0])
		}
	case http.MethodPost:
		if len(parts) > 0 {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.createDataFlowCron(w, r, namespace)
	case http.MethodPut:
		if len(parts) > 0 {
			h.updateDataFlowCron(w, r, namespace, parts[0])
		} else {
			http.Error(w, "Name required", http.StatusBadRequest)
		}
	case http.MethodDelete:
		if len(parts) > 0 {
			h.deleteDataFlowCron(w, r, namespace, parts[0])
		} else {
			http.Error(w, "Name required", http.StatusBadRequest)
		}
	default:
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
	}
}

func (h *APIHandler) listDataFlowCrons(w http.ResponseWriter, r *http.Request, namespace string) {
	var list dataflowv1.DataFlowCronList
	opts := []client.ListOption{}
	if namespace != "" {
		opts = append(opts, client.InNamespace(namespace))
	}

	if err := h.server.client.List(r.Context(), &list, opts...); err != nil {
		h.server.logger.Error(err, "Failed to list DataFlowCrons")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(list.Items)
}

func (h *APIHandler) getDataFlowCron(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var dfc dataflowv1.DataFlowCron
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &dfc); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(dfc)
}

func (h *APIHandler) createDataFlowCron(w http.ResponseWriter, r *http.Request, namespace string) {
	var dfc dataflowv1.DataFlowCron
	if err := json.NewDecoder(r.Body).Decode(&dfc); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dfc.Namespace = namespace
	dfc.APIVersion = "dataflow.dataflow.io/v1"
	dfc.Kind = "DataFlowCron"

	if err := h.server.client.Create(r.Context(), &dfc); err != nil {
		h.server.logger.Error(err, "Failed to create DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(dfc)
}

func (h *APIHandler) updateDataFlowCron(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var dfc dataflowv1.DataFlowCron
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &dfc); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var updates dataflowv1.DataFlowCron
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dfc.Spec = updates.Spec

	if err := h.server.client.Update(r.Context(), &dfc); err != nil {
		h.server.logger.Error(err, "Failed to update DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(dfc)
}

func (h *APIHandler) deleteDataFlowCron(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var dfc dataflowv1.DataFlowCron
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &dfc); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if err := h.server.client.Delete(r.Context(), &dfc); err != nil {
		h.server.logger.Error(err, "Failed to delete DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

type suspendRequest struct {
	Suspend bool `json:"suspend"`
}

func (h *APIHandler) suspendDataFlowCron(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var dfc dataflowv1.DataFlowCron
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &dfc); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	var body suspendRequest
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	dfc.Spec.Suspend = &body.Suspend

	if err := h.server.client.Update(r.Context(), &dfc); err != nil {
		h.server.logger.Error(err, "Failed to suspend DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(dfc)
}

type triggerResponse struct {
	JobName   string `json:"jobName"`
	Namespace string `json:"namespace"`
}

func (h *APIHandler) triggerDataFlowCron(w http.ResponseWriter, r *http.Request, namespace, name string) {
	var dfc dataflowv1.DataFlowCron
	key := types.NamespacedName{Namespace: namespace, Name: name}

	if err := h.server.client.Get(r.Context(), key, &dfc); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get DataFlowCron")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if dfc.Spec.Suspend != nil && *dfc.Spec.Suspend {
		http.Error(w, "DataFlowCron is suspended", http.StatusConflict)
		return
	}

	if cronConcurrencyForbid(&dfc) {
		active, err := h.hasActiveDataFlowCronJob(r.Context(), namespace, name)
		if err != nil {
			h.server.logger.Error(err, "Failed to check active jobs")
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if active {
			http.Error(w, "A run is already in progress (concurrencyPolicy Forbid)", http.StatusConflict)
			return
		}
	}

	cronJobName := k8snames.CronJobName(name)
	var cronJob batchv1.CronJob
	cronKey := types.NamespacedName{Namespace: namespace, Name: cronJobName}
	if err := h.server.client.Get(r.Context(), cronKey, &cronJob); err != nil {
		if apierrors.IsNotFound(err) {
			http.Error(w, "CronJob not found; wait for operator reconciliation", http.StatusNotFound)
			return
		}
		h.server.logger.Error(err, "Failed to get CronJob")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	jobName := manualCronJobName(name, time.Now().Unix())
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      jobName,
			Namespace: namespace,
			Labels: map[string]string{
				dataFlowCronOwnerLabel:        name,
				dataFlowCronTriggerIndexLabel: dataFlowCronProcessorStep,
			},
			Annotations: map[string]string{
				cronJobInstantiateAnnotation: "manual",
			},
		},
		Spec: *cronJob.Spec.JobTemplate.Spec.DeepCopy(),
	}

	if err := h.server.client.Create(r.Context(), job); err != nil {
		h.server.logger.Error(err, "Failed to create manual Job")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(triggerResponse{JobName: jobName, Namespace: namespace})
}

func cronConcurrencyForbid(dfc *dataflowv1.DataFlowCron) bool {
	policy := dfc.Spec.ConcurrencyPolicy
	return policy == "" || policy == dataflowv1.DataFlowCronConcurrencyForbid
}

func (h *APIHandler) hasActiveDataFlowCronJob(ctx context.Context, namespace, name string) (bool, error) {
	var jobs batchv1.JobList
	if err := h.server.client.List(ctx, &jobs,
		client.InNamespace(namespace),
		client.MatchingLabels{dataFlowCronOwnerLabel: name},
	); err != nil {
		return false, err
	}
	for i := range jobs.Items {
		if isJobActive(&jobs.Items[i]) {
			return true, nil
		}
	}
	return false, nil
}

func isJobActive(job *batchv1.Job) bool {
	return !isJobSucceeded(job) && !isJobFailed(job)
}

func isJobSucceeded(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobComplete && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func isJobFailed(job *batchv1.Job) bool {
	for _, c := range job.Status.Conditions {
		if c.Type == batchv1.JobFailed && c.Status == corev1.ConditionTrue {
			return true
		}
	}
	return false
}

func manualCronJobName(cronName string, unix int64) string {
	base := fmt.Sprintf("%s-manual-%d", k8snames.CronJobName(cronName), unix)
	if len(base) <= 63 {
		return base
	}
	suffix := strconv.FormatInt(unix, 10)
	maxPrefix := 63 - len(suffix) - 1
	if maxPrefix < 1 {
		return base[:63]
	}
	return strings.TrimRight(base[:maxPrefix], "-") + "-" + suffix
}

func (h *APIHandler) listDataFlowCronPods(r *http.Request, namespace, name, step, runID string) (*corev1.PodList, string, error) {
	selector := dataFlowCronOwnerLabel + "=" + name
	if step != "" {
		selector += "," + dataFlowCronTriggerIndexLabel + "=" + step
	} else {
		selector += "," + dataFlowCronTriggerIndexLabel + "=" + dataFlowCronProcessorStep
	}
	if runID != "" {
		selector += ",dataflow.dataflow.io/run-id=" + runID
	}

	pods, err := h.server.k8sClient.CoreV1().Pods(namespace).List(r.Context(), metav1.ListOptions{
		LabelSelector: selector,
	})
	if err != nil {
		return nil, "", err
	}

	containerName := "processor"
	if step != "" && step != dataFlowCronProcessorStep {
		containerName = ""
	}
	return pods, containerName, nil
}

func selectNewestPod(pods []corev1.Pod) corev1.Pod {
	if len(pods) == 1 {
		return pods[0]
	}
	best := pods[0]
	for i := 1; i < len(pods); i++ {
		if pods[i].CreationTimestamp.After(best.CreationTimestamp.Time) {
			best = pods[i]
		}
	}
	return best
}
