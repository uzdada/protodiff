// Copyright 2025 ProtoDiff Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package k8s provides a Kubernetes client for discovering gRPC pods and
// loading configuration.
//
// This package wraps the official Kubernetes client-go library to provide
// application-specific operations:
//   - Discovering pods labeled with grpc-service=true
//   - Loading service-to-BSR mappings from ConfigMaps
//   - Retrieving pod network information for gRPC connections
//
// The client uses in-cluster configuration when running inside Kubernetes,
// automatically authenticating via the service account token.
//
// Example usage:
//
//	client, err := k8s.NewClient()
//	pods, err := client.DiscoverGRPCPods(ctx)
//	mappings, err := client.LoadServiceMappings(ctx, "default", "config")
package k8s

import (
	"context"
	"fmt"

	"github.com/uzdada/protodiff/internal/core/domain"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

const (
	// GRPCServiceLabel is the label used to identify gRPC-enabled pods
	GRPCServiceLabel = "grpc-service"
	// ServiceNameLabel is the label containing the logical service name
	ServiceNameLabel = "app"
	// DefaultGRPCPort is the default port for gRPC reflection
	DefaultGRPCPort = 9090
)

// Client provides Kubernetes API operations
type Client struct {
	clientset *kubernetes.Clientset
}

// NewClient creates a new Kubernetes client using in-cluster configuration
func NewClient() (*Client, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to create in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create kubernetes clientset: %w", err)
	}

	return &Client{
		clientset: clientset,
	}, nil
}

// PodInfo contains information about a discovered gRPC pod
type PodInfo struct {
	Name        string
	Namespace   string
	ServiceName string
	IP          string
	GRPCPort    int32
}

// DiscoverGRPCPods finds all pods labeled with grpc-service=true
// Deprecated: Use DiscoverPodsForServices for ConfigMap-based discovery
func (c *Client) DiscoverGRPCPods(ctx context.Context) ([]PodInfo, error) {
	pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
		LabelSelector: fmt.Sprintf("%s=true", GRPCServiceLabel),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list pods: %w", err)
	}

	var podInfos []PodInfo
	for _, pod := range pods.Items {
		// Skip pods that are not running
		if pod.Status.Phase != corev1.PodRunning {
			continue
		}

		// Extract service name from labels
		serviceName := pod.Labels[ServiceNameLabel]
		if serviceName == "" {
			serviceName = "unknown"
		}

		// Determine gRPC port from container ports, fallback to 9090
		grpcPort := int32(DefaultGRPCPort)
		if len(pod.Spec.Containers) > 0 {
			for _, container := range pod.Spec.Containers {
				for _, port := range container.Ports {
					if port.Name == "grpc" || port.Protocol == corev1.ProtocolTCP {
						grpcPort = port.ContainerPort
						break
					}
				}
				if grpcPort != DefaultGRPCPort {
					break
				}
			}
		}

		podInfos = append(podInfos, PodInfo{
			Name:        pod.Name,
			Namespace:   pod.Namespace,
			ServiceName: serviceName,
			IP:          pod.Status.PodIP,
			GRPCPort:    grpcPort,
		})
	}

	return podInfos, nil
}

// DiscoverPodsForServices finds pods for specific service names from ConfigMap
// This is more efficient than label-based discovery when you have explicit service mappings
func (c *Client) DiscoverPodsForServices(ctx context.Context, serviceNames []string) ([]PodInfo, error) {
	var podInfos []PodInfo

	for _, serviceName := range serviceNames {
		// Find pods with app=serviceName label across all namespaces
		pods, err := c.clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			LabelSelector: fmt.Sprintf("%s=%s", ServiceNameLabel, serviceName),
		})
		if err != nil {
			return nil, fmt.Errorf("failed to list pods for service %s: %w", serviceName, err)
		}

		for _, pod := range pods.Items {
			// Skip pods that are not running
			if pod.Status.Phase != corev1.PodRunning {
				continue
			}

			// Determine gRPC port from container ports, fallback to 9090
			grpcPort := int32(DefaultGRPCPort)
			if len(pod.Spec.Containers) > 0 {
				for _, container := range pod.Spec.Containers {
					for _, port := range container.Ports {
						if port.Name == "grpc" || port.Protocol == corev1.ProtocolTCP {
							grpcPort = port.ContainerPort
							break
						}
					}
					if grpcPort != DefaultGRPCPort {
						break
					}
				}
			}

			podInfos = append(podInfos, PodInfo{
				Name:        pod.Name,
				Namespace:   pod.Namespace,
				ServiceName: serviceName,
				IP:          pod.Status.PodIP,
				GRPCPort:    grpcPort,
			})
		}
	}

	return podInfos, nil
}

// GetConfigMap retrieves a ConfigMap from the specified namespace.
// It uses the Kubernetes client to fetch the ConfigMap by name and returns
// an error if the ConfigMap does not exist or cannot be accessed.
func (c *Client) GetConfigMap(ctx context.Context, namespace, name string) (*corev1.ConfigMap, error) {
	cm, err := c.clientset.CoreV1().
		ConfigMaps(namespace).
		Get(ctx, name, metav1.GetOptions{}) // default value (no options)

	if err != nil {
		return nil, fmt.Errorf("failed to get configmap %s/%s: %w", namespace, name, err)
	}
	return cm, nil
}

// LoadServiceMappings loads service-to-BSR mappings from a ConfigMap
func (c *Client) LoadServiceMappings(ctx context.Context, namespace, configMapName string) (domain.ServiceMappings, error) {
	cm, err := c.GetConfigMap(ctx, namespace, configMapName)
	if err != nil {
		return domain.ServiceMappings{}, err
	}

	// Convert ConfigMap data to domain.ServiceMappings
	// The ConfigMap data has keys as service names and values as BSR module URLs
	return domain.NewServiceMappings(cm.Data), nil
}
