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

package controller

import (
	"context"
	"time"

	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
	infrastructurev1beta1 "dcnlab.ssu.ac.kr/kt-cloud-operator/api/v1beta1"
)

// ClusterReconciler reconciles a Cluster object
type ClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=clusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=clusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=infrastructure.dcnlab.ssu.ac.kr,resources=clusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Cluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.19.1/pkg/reconcile
func (r *ClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	logger := log.FromContext(ctx)
	logger.V(1).Info("KTCluster Reconcile", "ktCluster", req)

	// Fetch the ktcluster instance
	cluster := &v1beta1.Cluster{}
	if err := r.Get(ctx, req.NamespacedName, cluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("KTCluster resource not found. Ignoring since it must be deleted")
			return ctrl.Result{}, nil
		}
		logger.Error(err, "Failed to get KTCluster resource")
		return ctrl.Result{}, err
	}

	// Check if the child resources already exists to add ownerRef
	foundCluster := &v1beta1.Cluster{}
	err := r.Get(ctx, types.NamespacedName{Name: cluster.Name, Namespace: cluster.Namespace}, foundCluster)
	if err != nil && apierrors.IsNotFound(err) {
		// Read through the cluster Object
		kt_clstr, err := r.ktClusterForCluster(cluster, ctx, req)
		logger.Info("adding owner ref for ktcluster ", "KTCluster.Namespace ", kt_clstr.Namespace, " KTCluster.Name ", kt_clstr.Name)
		if err != nil {
			if err := r.Create(ctx, kt_clstr); err != nil {
				logger.Error(err, "Failed to add owner ref to ", "KTCluster.Namespace ", kt_clstr.Namespace, "KTCluster.Name", kt_clstr.Name)
				return ctrl.Result{}, err
			}
		}

		// Requeue the request to ensure the Cluster is given Owner ref
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	} else if err != nil {
		logger.Error(err, "Failed to get KTCluster, maybe dont have")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// used to create Owner Refs
func (r *ClusterReconciler) ktClusterForCluster(cluster *v1beta1.Cluster, ctx context.Context, req ctrl.Request) (*v1beta1.KTCluster, error) {
	logger := log.FromContext(ctx)
	logger.Info("Cluster Reconcile In ktClusterForCluster FN")

	// Fetch the ktcluster instance
	ktCluster := &v1beta1.KTCluster{}
	if err := r.Get(ctx, req.NamespacedName, ktCluster); err != nil {
		if apierrors.IsNotFound(err) {
			logger.Info("Cluster resource not found. Ignoring since it might be deleted or not created")
			return &infrastructurev1beta1.KTCluster{}, err
		}
		logger.Error(err, "Failed to get Cluster resource")
		return &infrastructurev1beta1.KTCluster{}, err
	}

	// Set the ownerRef for the Cluster, ensuring that the Deployment
	// will be deleted when the KTCluster CR is deleted.
	controllerutil.SetControllerReference(cluster, ktCluster, r.Scheme)
	return ktCluster, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1beta1.Cluster{}).
		Owns(&v1beta1.KTCluster{}).
		Named("cluster").
		Complete(r)
}
