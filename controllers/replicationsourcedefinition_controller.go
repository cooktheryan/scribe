/*
Copyright 2020 The Scribe authors.

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published
by the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <https://www.gnu.org/licenses/>.
*/

package controllers

import (
	"context"

	"github.com/go-logr/logr"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	scribev1alpha1 "github.com/backube/scribe/api/v1alpha1"
)

// ReplicationSourceDefinitionReconciler reconciles a ReplicationSource object
type ReplicationSourceDefinitionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type SourceDefinitionReconciler struct {
	Instance *scribev1alpha1.ReplicationSourceDefinition
	Ctx      context.Context
	Log      logr.Logger
	Scheme   *runtime.Scheme
	client.Client
}

//nolint:lll
//nolint:funlen
//+kubebuilder:rbac:groups=scribe.backube,resources=replicationsourcedefinitions,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scribe.backube,resources=replicationsourcedefinitions/finalizers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=scribe.backube,resources=replicationsourcedefinitions/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=core,resources=persistentvolumeclaims,verbs=get;list

//nolint:funlen
func (r *ReplicationSourceDefinitionReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := ctrllog.FromContext(ctx)

	// Fetch the Extract instance
	instance := &scribev1alpha1.ReplicationSourceDefinition{}
	err := r.Get(ctx, req.NamespacedName, instance)
	if err != nil {
		if kerrors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("Extract resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get Extract")
		return ctrl.Result{}, err
	}

	// Check if the RS already exists, if not create a new one
	found := &scribev1alpha1.ReplicationSource{}
	err = r.Get(ctx, types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)
	if err != nil && kerrors.IsNotFound(err) {
		// Define a new SR
		SR := r.EnsureSR(instance)
		log.Info("Creating a new SR", "SR.Namespace", SR.Namespace, "SR.Name", SR.Name)
		err = r.Create(ctx, SR)
		if err != nil {
			log.Error(err, "Failed to create new SR", "SR.Namespace", SR.Namespace, "SR.Name", SR.Name)
			return ctrl.Result{}, err
		}
		// SR created successfully - return and requeue
		return ctrl.Result{Requeue: true}, nil
	} else if err != nil {
		log.Error(err, "Failed to get SR")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

//nolint:lll
//nolint:funlen
func (r *ReplicationSourceDefinitionReconciler) EnsureSR(m *scribev1alpha1.ReplicationSourceDefinition) *scribev1alpha1.ReplicationSource {
	SR := &scribev1alpha1.ReplicationSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: scribev1alpha1.ReplicationSourceSpec{
			SourcePVC: "mypvc",
			Trigger:   &scribev1alpha1.ReplicationSourceTriggerSpec{},
			Rclone: &scribev1alpha1.ReplicationSourceRcloneSpec{
				ReplicationSourceVolumeOptions: scribev1alpha1.ReplicationSourceVolumeOptions{
					CopyMethod: "None",
				},
				RcloneConfigSection: m.Spec.RcloneConfigSection,
				RcloneDestPath:      m.Spec.RcloneDestPath,
				RcloneConfig:        m.Spec.RcloneConfig,
			},
		},
	}
	// Set Memcached instance as the owner and controller
	ctrl.SetControllerReference(m, SR, r.Scheme)
	return SR
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReplicationSourceDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scribev1alpha1.ReplicationSourceDefinition{}).
		Owns(&scribev1alpha1.ReplicationSource{}).
		Complete(r)
}
