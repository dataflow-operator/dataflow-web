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
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/kubernetes"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	dataflowv1 "github.com/dataflow-operator/dataflow/api/v1"
)

// staticHandler serves from staticDir; for "/" and "/index.html" falls back to rootDir if not in static (local dev).
// For SPA (Vue Router history mode), any non-file GET request is served index.html.
type staticHandler struct {
	staticDir string
	rootDir   string
}

func (h staticHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "" || path == "/" {
		path = "/index.html"
	}

	// Try to serve as file from static dir
	staticPath := filepath.Join(h.staticDir, filepath.FromSlash(path))
	if info, err := os.Stat(staticPath); err == nil && !info.IsDir() {
		http.ServeFile(w, r, staticPath)
		return
	}

	// For /index.html: fallback to root dir (local dev when index.html is in repo root)
	if path == "/index.html" {
		rootIndex := filepath.Join(h.rootDir, "index.html")
		if _, err := os.Stat(rootIndex); err == nil {
			http.ServeFile(w, r, rootIndex)
			return
		}
	}

	// SPA fallback: serve index.html for GET so Vue Router can handle the route
	if r.Method == http.MethodGet {
		indexPath := filepath.Join(h.staticDir, "index.html")
		if _, err := os.Stat(indexPath); err == nil {
			http.ServeFile(w, r, indexPath)
			return
		}
		rootIndex := filepath.Join(h.rootDir, "index.html")
		if _, err := os.Stat(rootIndex); err == nil {
			http.ServeFile(w, r, rootIndex)
			return
		}
	}

	http.NotFound(w, r)
}

// Server represents the web server for the GUI.
type Server struct {
	bindAddr   string
	httpServer *http.Server
	client     client.Client
	k8sClient  *kubernetes.Clientset
	logger     logr.Logger
}

// NewServer creates a new GUI server.
func NewServer(bindAddr, kubeconfig string) (*Server, error) {
	logger := log.Log.WithName("gui-server")

	var config *rest.Config
	var err error

	if kubeconfig != "" {
		config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	} else {
		config, err = rest.InClusterConfig()
		if err != nil {
			home, homeErr := os.UserHomeDir()
			if homeErr == nil {
				defaultKubeconfig := filepath.Join(home, ".kube", "config")
				config, err = clientcmd.BuildConfigFromFlags("", defaultKubeconfig)
			}
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get kubernetes config: %w", err)
	}

	scheme := runtime.NewScheme()
	utilruntime.Must(clientgoscheme.AddToScheme(scheme))
	utilruntime.Must(dataflowv1.AddToScheme(scheme))

	k8sClient, err := client.New(config, client.Options{Scheme: scheme})
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes client: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	server := &Server{
		bindAddr:  bindAddr,
		client:    k8sClient,
		k8sClient: clientset,
		logger:    logger,
	}

	mux := http.NewServeMux()
	apiHandler := NewAPIHandler(server)
	mux.Handle("/api/", http.StripPrefix("/api", apiHandler))
	// Static files: prefer ./static (in container /app/static), fallback to . for / and /index.html (local dev)
	mux.Handle("/", staticHandler{staticDir: "./static", rootDir: "."})

	server.httpServer = &http.Server{
		Addr:         bindAddr,
		Handler:      mux,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	return server, nil
}

// Start starts the HTTP server.
func (s *Server) Start() error {
	s.logger.Info("Starting GUI server", "address", s.bindAddr)
	return s.httpServer.ListenAndServe()
}

// Stop stops the HTTP server.
func (s *Server) Stop(ctx context.Context) error {
	s.logger.Info("Stopping GUI server")
	return s.httpServer.Shutdown(ctx)
}
