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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	prefectv1alpha1 "github.com/aigency/prefect-operator/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=prefect.aigency.com,resources=clusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=prefect.aigency.com,resources=clusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=prefect.aigency.com,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	cluster := prefectv1alpha1.Cluster{}

	if err := r.Get(ctx, req.NamespacedName, &cluster); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.reconcileClusterController(ctx, &cluster); err != nil {
		return ctrl.Result{}, err
	}

	for _, agentPool := range cluster.Spec.AgentPools {
		if err := r.reconcileAgentPool(ctx, &cluster, &agentPool); err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&prefectv1alpha1.Cluster{}).
		Complete(r)
}

func (r *ClusterReconciler) reconcileClusterController(ctx context.Context, cluster *prefectv1alpha1.Cluster) error {
	deploymentName := fmt.Sprintf("%s-prefect-server", cluster.GetName())
	deployment := &appsv1.Deployment{}

	if err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: cluster.GetNamespace()}, deployment); err != nil {
		if errors.IsNotFound(err) {
			return r.createControllerDeployment(ctx, cluster)
		}

		return err
	}

	if err := r.updateControllerDeployment(ctx, cluster); err != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) reconcileAgentPool(ctx context.Context, cluster *prefectv1alpha1.Cluster, agentPool *prefectv1alpha1.AgentPoolSpec) error {
	return nil
}

func (r *ClusterReconciler) createControllerDeployment(ctx context.Context, cluster *prefectv1alpha1.Cluster) error {
	deploymentName := fmt.Sprintf("%s-prefect-server", cluster.GetName())
	deploymentLabels := map[string]string{
		"prefect.aigency.com/cluster": cluster.GetName(),
		"prefect.aigency.com/role":    "controller",
	}

	deployment := &appsv1.Deployment{
		ObjectMeta: ctrl.ObjectMeta{
			Name:      deploymentName,
			Namespace: cluster.GetNamespace(),
			Labels:    deploymentLabels,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: cluster.Spec.Controller.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: deploymentLabels,
			},
			Template: v1.PodTemplateSpec{
				Spec: v1.PodSpec{
					Containers: []v1.Container{
						{
							Name:  "prefect-server",
							Image: cluster.Spec.Controller.Image,
							Command: []string{
								"prefect",
								"server",
								"--host",
								"0.0.0.0",
								"--port",
								"4200",
							},
							Ports: []v1.ContainerPort{
								{
									Name:          "http-api",
									ContainerPort: 4200,
								},
							},
							Resources: cluster.Spec.Controller.Resources,
						},
					},
				},
			},
		},
	}

	if err := ctrl.SetControllerReference(cluster, deployment, r.Scheme); err != nil {
		return err
	}

	if err := r.Create(ctx, deployment); err != nil {
		return err
	}

	return nil
}

func (r *ClusterReconciler) updateControllerDeployment(ctx context.Context, cluster *prefectv1alpha1.Cluster) error {
	deploymentName := fmt.Sprintf("%s-prefect-server", cluster.GetName())
	deployment := &appsv1.Deployment{}
	namespacedName := types.NamespacedName{
		Name:      deploymentName,
		Namespace: cluster.GetNamespace(),
	}

	if err := r.Get(ctx, namespacedName, deployment); err != nil {
		return err
	}

	deployment.Spec.Replicas = cluster.Spec.Controller.Replicas
	deployment.Spec.Template.Spec.Containers[0].Image = cluster.Spec.Controller.Image
	deployment.Spec.Template.Spec.Containers[0].Resources = cluster.Spec.Controller.Resources

	if err := r.Update(ctx, deployment); err != nil {
		return err
	}

	return nil
}
