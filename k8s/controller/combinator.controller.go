package controller

import (
	"context"
	"log"

	"jabberwocky238/console/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/client-go/tools/cache"
)

type CombinatorController struct {
	ctrl    *Controller
	crCache cache.Store
}

// --- CR event handlers ---

func (cc *CombinatorController) onAdd(obj interface{}) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] CombinatorApp added: %s", u.GetName())
	cc.reconcile(u)
}

func (cc *CombinatorController) onUpdate(oldObj, newObj interface{}) {
	u, ok := newObj.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] CombinatorApp updated: %s", u.GetName())
	cc.reconcile(u)
}

func (cc *CombinatorController) onDelete(obj interface{}) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] CombinatorApp deleted: %s", u.GetName())

	cb := combinatorFromUnstructured(u)
	if cb == nil {
		return
	}
	cb.DeleteAll(context.Background())
}

// --- Sub-resource delete handler ---

func (cc *CombinatorController) onSubResourceDelete(obj interface{}) {
	var appName string
	switch o := obj.(type) {
	case metav1.Object:
		appName = o.GetLabels()["app"]
	case *unstructured.Unstructured:
		appName = o.GetLabels()["app"]
	default:
		return
	}
	if appName == "" {
		return
	}

	key := k8s.CombinatorNamespace + "/" + appName
	item, exists, err := cc.crCache.GetByKey(key)
	if err != nil || !exists {
		return
	}
	u, ok := item.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] combinator sub-resource deleted for %s, re-reconciling", appName)
	cc.reconcile(u)
}

// --- Reconcile ---

func (cc *CombinatorController) reconcile(u *unstructured.Unstructured) {
	cb := combinatorFromUnstructured(u)
	if cb == nil {
		return
	}

	ctx := context.Background()
	cc.ctrl.updateStatus(u, CombinatorAppGVR, "Deploying", "")

	if err := cb.EnsureConfigMap(ctx); err != nil {
		log.Printf("[controller] ensure configmap for %s failed: %v", u.GetName(), err)
		cc.ctrl.updateStatus(u, CombinatorAppGVR, "Failed", err.Error())
		return
	}
	if err := cb.EnsureDeployment(ctx); err != nil {
		log.Printf("[controller] ensure deployment for %s failed: %v", u.GetName(), err)
		cc.ctrl.updateStatus(u, CombinatorAppGVR, "Failed", err.Error())
		return
	}
	if err := cb.EnsureService(ctx); err != nil {
		log.Printf("[controller] ensure service for %s failed: %v", u.GetName(), err)
		cc.ctrl.updateStatus(u, CombinatorAppGVR, "Failed", err.Error())
		return
	}
	if err := cb.EnsureIngressRoute(ctx); err != nil {
		log.Printf("[controller] ensure ingress route for %s failed: %v", u.GetName(), err)
		cc.ctrl.updateStatus(u, CombinatorAppGVR, "Failed", err.Error())
		return
	}

	cb.ReloadConfig([]byte(cb.Config))

	log.Printf("[controller] reconcile combinator %s success", u.GetName())
	cc.ctrl.updateStatus(u, CombinatorAppGVR, "Running", "")
}
