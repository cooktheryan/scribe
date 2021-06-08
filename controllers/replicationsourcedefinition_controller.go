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
	"github.com/operator-framework/operator-lib/status"
	corev1 "k8s.io/api/core/v1"
	kerrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrlutil "sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	scribev1alpha1 "github.com/backube/scribe/api/v1alpha1"
)

// ReplicationSourceDefinitionReconciler reconciles a ReplicationSource object
type ReplicationSourceDefinitionReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

type SourceDefinitionReconciler struct {
	replicationSource *scribev1alpha1.ReplicationSource
	Instance          *scribev1alpha1.ReplicationSourceDefinition
	Ctx               context.Context
	Log               logr.Logger
	Scheme            *runtime.Scheme
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
	logger := r.Log.WithValues("replicationsourcedefinition", req.NamespacedName)
	inst := &scribev1alpha1.ReplicationSourceDefinition{}
	if err := r.Client.Get(ctx, req.NamespacedName, inst); err != nil {
		if kerrors.IsNotFound(err) {
			logger.Error(err, "Failed to get Definition")
		}
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if inst.Status == nil {
		inst.Status = &scribev1alpha1.ReplicationSourceDefinitionStatus{}
	}
	if inst.Status.Conditions == nil {
		inst.Status.Conditions = status.Conditions{}
	}

	var result ctrl.Result
	var err error
	if inst.Spec.ReplicationMethod != "" {
		result, err = RunRcloneSrcDefReconciler(ctx, inst, r, logger)
	} else {
		return ctrl.Result{}, nil
	}

	// Set reconcile status condition
	if err == nil {
		inst.Status.Conditions.SetCondition(
			status.Condition{
				Type:    scribev1alpha1.ConditionReconciled,
				Status:  corev1.ConditionTrue,
				Reason:  scribev1alpha1.ReconciledReasonComplete,
				Message: "Reconcile complete",
			})
	} else {
		inst.Status.Conditions.SetCondition(
			status.Condition{
				Type:    scribev1alpha1.ConditionReconciled,
				Status:  corev1.ConditionFalse,
				Reason:  scribev1alpha1.ReconciledReasonError,
				Message: err.Error(),
			})
	}

	return result, err
}

// RunRcloneSrcReconciler is invoked when ReplicationSource.Spec.Rclone != nil
func RunRcloneSrcDefReconciler(
	ctx context.Context,
	instance *scribev1alpha1.ReplicationSourceDefinition,
	sr *ReplicationSourceDefinitionReconciler,
	logger logr.Logger,
) (ctrl.Result, error) {
	r := SourceDefinitionReconciler{
		Instance: instance,
		Ctx:      ctx,
	}

	l := logger.WithValues("method", "Rclone")

	_, err := reconcileBatch(l,
		r.ensureSource,
	)
	return ctrl.Result{}, err
}

func (r *ReplicationSourceDefinitionReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&scribev1alpha1.ReplicationSourceDefinition{}).
		Owns(&scribev1alpha1.ReplicationSource{}).
		Complete(r)
}

//nolint:funlen
func (r *SourceDefinitionReconciler) ensureSource(l logr.Logger) (bool, error) {
	r.replicationSource = &scribev1alpha1.ReplicationSource{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "scribe-rclone-src-" + r.Instance.Name,
			Namespace: r.Instance.Namespace,
		},
	}
	logger := l.WithValues("job", nameFor(r.replicationSource))
	op, err := ctrlutil.CreateOrUpdate(r.Ctx, r.Client, r.replicationSource, func() error {
		if err := ctrl.SetControllerReference(r.Instance, r.replicationSource, r.Scheme); err != nil {
			logger.Error(err, "unable to set controller reference")
			return err
		}

		r.replicationSource.Spec.SourcePVC = "mypvc"
		r.replicationSource.Spec.Rclone.RcloneConfigSection = r.Instance.Spec.RcloneConfigSection
		r.replicationSource.Spec.Rclone.RcloneDestPath = r.Instance.Spec.RcloneConfigSection
		r.replicationSource.Spec.Rclone.RcloneConfig = r.Instance.Spec.RcloneConfigSection
		r.replicationSource.Spec.Rclone.CopyMethod = r.Instance.Spec.CopyMethod
		logger.V(1).Info("Job has PVC")
		return nil
	})
	if err != nil {
		logger.Error(err, "reconcile failed")
	} else {
		logger.V(1).Info("Definition reconciled", "operation", op)
	}
	return true, nil
}
