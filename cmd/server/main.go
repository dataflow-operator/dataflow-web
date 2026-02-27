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

package main

import (
	"flag"
	"os"
	"path/filepath"

	_ "k8s.io/client-go/plugin/pkg/client/auth"

	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	ctrl "sigs.k8s.io/controller-runtime"

	"github.com/dataflow-operator/dataflow-web/internal/gui"
)

var setupLog = ctrl.Log.WithName("setup")

func main() {
	var bindAddr string
	var logLevel string

	flag.StringVar(&bindAddr, "bind-address", ":8080", "The address the GUI server binds to.")
	flag.StringVar(&logLevel, "log-level", "info", "Log level: debug, info, warn, error")
	flag.Parse()

	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		if f := flag.Lookup("kubeconfig"); f != nil {
			kubeconfig = f.Value.String()
		}
	}
	if kubeconfig == "" {
		if home, err := os.UserHomeDir(); err == nil && home != "" {
			defaultPath := filepath.Join(home, ".kube", "config")
			if _, err := os.Stat(defaultPath); err == nil {
				kubeconfig = defaultPath
			}
		}
	}

	var level zapcore.Level
	switch logLevel {
	case "debug":
		level = zapcore.DebugLevel
	case "info":
		level = zapcore.InfoLevel
	case "warn":
		level = zapcore.WarnLevel
	case "error":
		level = zapcore.ErrorLevel
	default:
		level = zapcore.InfoLevel
	}

	config := zap.NewProductionConfig()
	config.Level = zap.NewAtomicLevelAt(level)
	config.EncoderConfig = zap.NewProductionEncoderConfig()
	zapLogger, err := config.Build()
	if err != nil {
		os.Stderr.WriteString("gui-server: failed to create logger: " + err.Error() + "\n")
		os.Exit(2)
	}
	ctrl.SetLogger(zapr.NewLogger(zapLogger))

	setupLog.Info("Starting GUI server", "bind-address", bindAddr)

	server, err := gui.NewServer(bindAddr, kubeconfig)
	if err != nil {
		setupLog.Error(err, "unable to create GUI server")
		_ = zapLogger.Sync()
		os.Exit(1)
	}

	if err := server.Start(); err != nil {
		setupLog.Error(err, "unable to start GUI server")
		_ = zapLogger.Sync()
		os.Exit(1)
	}
}
