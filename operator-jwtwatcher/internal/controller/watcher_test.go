package controller

import (
	"context"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"testing"

	_ "github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	v12 "istio.io/client-go/pkg/apis/security/v1"
	"k8s.io/apimachinery/pkg/types"
	watcherv1 "oprator-reqauth/api/v1"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
)

func TestReconcile(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	watcherv1.AddToScheme(scheme)
	v12.AddToScheme(scheme)

	t.Run("Reconcile with existing RequestAuthentication", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&v12.RequestAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing",
				Namespace: "default",
			},
		}).Build()

		r := &ReqAuthWatcherReconciler{
			Client: client,
			Scheme: scheme,
		}

		_, err := r.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      "existing",
				Namespace: "default",
			},
		})

		assert.NoError(t, err)
	})

	t.Run("Reconcile with non-existing RequestAuthentication", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).Build()

		r := &ReqAuthWatcherReconciler{
			Client: client,
			Scheme: scheme,
		}

		_, err := r.Reconcile(ctx, ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      "non-existing",
				Namespace: "default",
			},
		})

		assert.Error(t, err)
	})
}

func TestReconcileAuthorizationPolicy(t *testing.T) {
	ctx := context.Background()
	scheme := runtime.NewScheme()
	watcherv1.AddToScheme(scheme)
	v12.AddToScheme(scheme)
	logger := log.FromContext(ctx)
	t.Run("ReconcileAuthorizationPolicy with existing RequestAuthentication", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).WithObjects(&v12.RequestAuthentication{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing",
				Namespace: "default",
			},
		}).Build()

		r := &ReqAuthWatcherReconciler{
			Client: client,
			Scheme: scheme,
		}

		err := r.reconcileAuthorizationPolicy(ctx, &watcherv1.ReqAuthWatcher{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "existing",
				Namespace: "default",
			},
		}, logger)

		assert.NoError(t, err)
	})

	t.Run("ReconcileAuthorizationPolicy with non-existing RequestAuthentication", func(t *testing.T) {
		client := fake.NewClientBuilder().WithScheme(scheme).Build()
		logger := log.FromContext(ctx)
		r := &ReqAuthWatcherReconciler{
			Client: client,
			Scheme: scheme,
		}

		err := r.reconcileAuthorizationPolicy(ctx, &watcherv1.ReqAuthWatcher{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "non-existing",
				Namespace: "default",
			},
		}, logger)

		assert.Error(t, err)
	})
}
