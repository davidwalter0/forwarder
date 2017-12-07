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

	// Uncomment the following line to load the gcp plugin (only required to authenticate against GKE clusters).
	// _ "k8s.io/client-go/plugin/pkg/client/auth/gcp"

	"github.com/davidwalter0/forwarder/share"
	"github.com/davidwalter0/forwarder/tracer"
	"github.com/davidwalter0/transform"

	yaml "gopkg.in/yaml.v2"
	"k8s.io/client-go/tools/clientcmd"
)

// kubeRestConfig kubernetes config object
var kubeRestConfig *restclient.Config

// clientSet api calls
var clientSet *kubernetes.Clientset

// CheckInCluster reports if the env variable is set for cluster
func CheckInCluster() bool {
	return len(os.Getenv("KUBERNETES_PORT")) > 0
}

// LoadKubernetesConfig sets up kubernetes authentication options
func LoadKubernetesConfig(cfg *share.ServerCfg) {
	var err error
	var jsonText []byte

	jsonText, _ = json.MarshalIndent(cfg, "", "  ")
	if cfg.Debug {
		fmt.Printf("\n%v\n", string(jsonText))
	}
	trace.Enabled = cfg.Debug

	// creates the in-cluster configuration
	kubeRestConfig, err = restclient.InClusterConfig()
	if err != nil {
		// try with a kubeconfig file
		kubeRestConfig, err = clientcmd.BuildConfigFromFlags("", cfg.Kubeconfig)
	}

	if err == nil {
		// creates the clientSet
		clientSet, err = kubernetes.NewForConfig(kubeRestConfig)
		if err == nil {
			cfg.Kubernetes = true
		}
	}
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
func Endpoints(name, namespace string) (endpoints []string) {
	if clientSet != nil {
		var eps *v1.EndpointsList
		var err error
		// use the current context in kubeconfig
		eps, err = clientSet.CoreV1().Endpoints(namespace).List(metav1.ListOptions{})
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

// Yamlify object to yaml string
func Yamlify(data interface{}) string {
	data, err := transform.TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	s, err := yaml.Marshal(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}

// Jsonify an object
func Jsonify(data interface{}) string {
	var err error
	data, err = transform.TransformData(data)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	s, err := json.MarshalIndent(data, "", "  ") // spaces)
	if err != nil {
		return fmt.Sprintf("%v", err)
	}
	return string(s)
}
