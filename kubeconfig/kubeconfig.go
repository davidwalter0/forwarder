/*
Copyright 2016 The Kubernetes Authors.

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

// Note: the example only works with the code within the same release/branch.
package kubeconfig

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	restclient "k8s.io/client-go/rest"

	// Uncomment the following line to load the gcp plugin (only
	// required to authenticate against GKE clusters).

	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/davidwalter0/forwarder/share"
	"github.com/davidwalter0/forwarder/tracer"

	"k8s.io/client-go/tools/clientcmd"
	// log "github.com/davidwalter0/logwriter"
)

// InCluster true when endpoints are accessible, when the service is
// running with the cluster's network namespaces
var InCluster bool

// CheckInCluster reports if the env variable is set for cluster
func CheckInCluster() bool {
	return len(os.Getenv("KUBERNETES_PORT")) > 0
}

// NewClientset returns a new handle to a kubernetes client
func NewClientset(cfg *share.ServerCfg) *kubernetes.Clientset {
	if cfg.Kubernetes {
		// kubeRestConfig kubernetes config object
		var kubeRestConfig *restclient.Config
		// clientset is a handle to execute kubernetes commands
		var clientset *kubernetes.Clientset
		var err error
		var jsonText []byte

		if cfg.Debug {
			jsonText, _ = json.MarshalIndent(cfg, "", "  ")
			fmt.Printf("\n%v\n", string(jsonText))
		}
		trace.Enabled = cfg.Debug

		// creates the in-cluster configuration
		kubeRestConfig, err = restclient.InClusterConfig()
		if err != nil {
			// try with a kubeconfig file
			kubeRestConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
		} else {
			InCluster = true
		}

		if err == nil {
			// creates the clientset
			clientset, err = kubernetes.NewForConfig(kubeRestConfig)
			if err == nil {
				cfg.Kubernetes = true
			}
		}
		return clientset
	}
	return nil
}

// ErrorHandler print error message based on error type
func ErrorHandler(name string, err error) {
	if errors.IsNotFound(err) {
		log.Printf("%s not found\n", name)
	} else if statusError, isStatus := err.(*errors.StatusError); isStatus {
		log.Printf("Error getting %s %v\n", name, statusError.ErrStatus.Message)
	} else if err != nil {
		panic(err.Error())
	} else {
		log.Printf("Found %s\n", name)
	}
}

// Endpoints for a service name in the given namespace
func Endpoints(clientset *kubernetes.Clientset, name, namespace string) (endpoints []string) {
	if clientset != nil {
		var eps *v1.EndpointsList
		var err error
		// use the current context in kubeconfig
		eps, err = clientset.CoreV1().Endpoints(namespace).List(metav1.ListOptions{})
		if err == nil {
			for _, ep := range eps.Items {
				meta := ep.ObjectMeta
				if name == meta.Name {
					for _, set := range ep.Subsets {
						for _, address := range set.Addresses {
							for _, port := range set.Ports {
								if len(meta.Namespace) > 0 {
									namespace = meta.Namespace
								}
								endpoint := fmt.Sprintf("%s:%d", address.IP, port.Port)
								endpoints = append(endpoints, endpoint)
							}
						}
					}
				}
			}
		} else {
			log.Println("Endpoint error", err)
		}
	}
	return
}

func homeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	return os.Getenv("USERPROFILE") // windows
}
