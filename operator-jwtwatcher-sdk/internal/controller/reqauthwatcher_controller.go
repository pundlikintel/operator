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
	"github.com/go-logr/logr"
	"github.com/google/uuid"
	securityv1 "istio.io/api/security/v1"
	"istio.io/api/type/v1beta1"
	v12 "istio.io/client-go/pkg/apis/security/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	authorizerv1alpha1 "github.com/pundlikintel/reqauth/api/v1alpha1"
)

// ReqAuthWatcherReconciler reconciles a ReqAuthWatcher object
type ReqAuthWatcherReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=authorizer.watcher.reqauth.com,resources=reqauthwatchers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=authorizer.watcher.reqauth.com,resources=reqauthwatchers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=authorizer.watcher.reqauth.com,resources=reqauthwatchers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ReqAuthWatcher object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.16.3/pkg/reconcile
func (r *ReqAuthWatcherReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Inside reconciliation", "Request", req)
	watcher := &authorizerv1alpha1.ReqAuthWatcher{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: req.Namespace,
		Name:      req.Name,
	}, watcher)
	if err != nil {
		logger.Error(err, "Error occurred in get NamespacedName")
		return ctrl.Result{}, err
	}
	logger.Info("Request watcher", "watcher-spec", watcher.Spec)
	err = r.reconcileAuthorizationPolicy(ctx, watcher, logger)
	if err != nil {
		logger.Error(err, "Error in reconciliation")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *ReqAuthWatcherReconciler) reconcileAuthorizationPolicy(ctx context.Context, watcher *authorizerv1alpha1.ReqAuthWatcher, l logr.Logger) error {
	reqAuth := &v12.RequestAuthentication{}
	err := r.Get(ctx, types.NamespacedName{
		Namespace: watcher.Namespace,
		Name:      watcher.Spec.Name,
	}, reqAuth)
	if err == nil {
		l.Info("RequestAuthentication already exists")
		hasUpdate := false
		if reqAuth.Spec.JwtRules[0].Issuer != watcher.Spec.Issuer {
			l.Info("Issuer mismatched")
			reqAuth.Spec.JwtRules[0].Issuer = watcher.Spec.Issuer
			hasUpdate = true
		}

		if reqAuth.Spec.JwtRules[0].Jwks != watcher.Spec.Jwks {
			l.Info("Jwks mismatched")
			reqAuth.Spec.JwtRules[0].Jwks = watcher.Spec.Jwks
			hasUpdate = true
		}

		if reqAuth.Name != watcher.Spec.Name {
			l.Info("name mismatched")
			reqAuth.Name = watcher.Spec.Name
			hasUpdate = true
		}

		if hasUpdate {
			err = r.Update(ctx, reqAuth)
			if err != nil {
				l.Error(err, "Error in RequestAuth Updated")
				return err
			}
		} else {
			l.Info("no update")
		}

		return nil
	}

	Build(reqAuth, watcher, l)

	err = r.Create(ctx, reqAuth)
	if err != nil {
		return err
	}
	return nil
}

func Build(ra *v12.RequestAuthentication, reqAuth *authorizerv1alpha1.ReqAuthWatcher, l logr.Logger) *v12.RequestAuthentication {
	ra.Namespace = reqAuth.Namespace
	ra.Name = reqAuth.Spec.Name

	//metadata
	ra.Annotations = map[string]string{"meta.helm.sh/release-name": "amber-cp", "meta.helm.sh/release-namespace": "dev"}
	ra.Generation = 3
	ra.Labels = map[string]string{"app.kubernetes.io/managed-by": "Helm", "app.kubernetes.io/name": "amber-cp", "app.kubernetes.io/version": "1.0.0"}
	ra.UID = types.UID(uuid.New().String())

	//spec
	ra.Spec = securityv1.RequestAuthentication{}
	ra.Spec.JwtRules = []*securityv1.JWTRule{{
		Issuer:      reqAuth.Spec.Issuer,
		Jwks:        reqAuth.Spec.Jwks,
		FromHeaders: []*securityv1.JWTHeader{{Name: reqAuth.Spec.HeaderName, Prefix: ""}},
	}}
	ra.Spec.Selector = &v1beta1.WorkloadSelector{MatchLabels: map[string]string{"interface-jwt-authn": "yes"}}
	return ra
}

// SetupWithManager sets up the controller with the Manager.
func (r *ReqAuthWatcherReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&authorizerv1alpha1.ReqAuthWatcher{}).
		Complete(r)
}
