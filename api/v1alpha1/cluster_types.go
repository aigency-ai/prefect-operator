/*
Copyright 2023 Aigency.

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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// AgentPoolSpec defines the desired state for a single agent pool
type AgentPoolSpec struct {
	// Name defines how the agent pool is identified.
	Name string `json:"name"`
	// Resources controls the resources assigned to each pod in the agent pool.
	//+kubebuilder:default={requests:{cpu: "1.0", memory: "1Gi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Image controls what container image is used for the pods in the agent pool.
	//+kubebuilder:default="prefecthq/prefect:2.10-python3.10"
	Image string `json:"image,omitempty"`
	// Replicas controls the number of pods deployed for the agent pool.
	//+kubebuilder:default=1
	//+kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`
}

// ControllerSpec defines the desired state for the prefect server component
type ControllerSpec struct {
	// Resources controls the resources assigned to each pod that runs the prefect server.
	//+kubebuilder:default={requests:{cpu: "1.0", memory: "1Gi"}}
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`
	// Image controls what container image is used to for the server pods.
	//+kubebuilder:default="prefecthq/prefect:2.10-python3.10"
	Image string `json:"image,omitempty"`
	// Replicas controls the number of pods deployed for the prefect server.
	//+kubebuilder:default=1
	//+kubebuilder:validation:Minimum=1
	Replicas *int32 `json:"replicas,omitempty"`
}

// ClusterSpec defines the desired state of Cluster
type ClusterSpec struct {
	// Controller defines the desired state for the prefect server in the cluster.
	Controller ControllerSpec `json:"controller,omitempty"`
	// AgentPools defines a list of agent pools to deploy in the cluster.
	AgentPools []AgentPoolSpec `json:"agentPools,omitempty"`
}

// ClusterStatus defines the observed state of Cluster
type ClusterStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// Cluster is the Schema for the clusters API
type Cluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   ClusterSpec   `json:"spec,omitempty"`
	Status ClusterStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// ClusterList contains a list of Cluster
type ClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Cluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Cluster{}, &ClusterList{})
}
