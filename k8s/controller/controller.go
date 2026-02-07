package controller

import (
	"context"
	"log"
	"time"

	"jabberwocky238/console/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type Controller struct {
	client     dynamic.Interface
	k8sClient  *kubernetes.Clientset
	worker     *WorkerController
	combinator *CombinatorController
}

func NewController(client dynamic.Interface, k8sClient *kubernetes.Clientset) *Controller {
	c := &Controller{client: client, k8sClient: k8sClient}
	c.worker = &WorkerController{ctrl: c}
	c.combinator = &CombinatorController{ctrl: c}
	return c
}

func (c *Controller) Start(stopCh <-chan struct{}) {
	// 1. CR informer: watch WorkerApp in worker namespace
	dynFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		c.client, 30*time.Second, k8s.WorkerNamespace, nil,
	)
	crInformer := dynFactory.ForResource(WorkerAppGVR).Informer()
	c.worker.crCache = crInformer.GetStore()

	crInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.worker.onAdd,
		UpdateFunc: c.worker.onUpdate,
		DeleteFunc: c.worker.onDelete,
	})

	// 2. Sub-resource informer: Deployment + Service + ConfigMap + Secret in worker namespace
	k8sFactory := informers.NewSharedInformerFactoryWithOptions(
		c.k8sClient, 30*time.Second,
		informers.WithNamespace(k8s.WorkerNamespace),
	)
	subHandler := cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.worker.onSubResourceDelete,
	}
	k8sFactory.Apps().V1().Deployments().Informer().AddEventHandler(subHandler)
	k8sFactory.Core().V1().Services().Informer().AddEventHandler(subHandler)

	// Watch ConfigMap and Secret updates to trigger Deployment rolling restart
	configHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.worker.onConfigUpdate,
	}
	k8sFactory.Core().V1().ConfigMaps().Informer().AddEventHandler(configHandler)
	k8sFactory.Core().V1().Secrets().Informer().AddEventHandler(configHandler)

	// 3. IngressRoute informer: watch IngressRoute in ingress namespace
	ingressDynFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		c.client, 30*time.Second, k8s.IngressNamespace, nil,
	)
	irInformer := ingressDynFactory.ForResource(k8s.IngressRouteGVR).Informer()
	irInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.worker.onSubResourceDelete,
	})

	// 4. CombinatorApp CR informer: watch CombinatorApp in combinator namespace
	combinatorDynFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		c.client, 30*time.Second, k8s.CombinatorNamespace, nil,
	)
	combinatorCrInformer := combinatorDynFactory.ForResource(CombinatorAppGVR).Informer()
	c.combinator.crCache = combinatorCrInformer.GetStore()

	combinatorCrInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.combinator.onAdd,
		UpdateFunc: c.combinator.onUpdate,
		DeleteFunc: c.combinator.onDelete,
	})

	// 5. Sub-resource informer for combinator namespace
	combinatorK8sFactory := informers.NewSharedInformerFactoryWithOptions(
		c.k8sClient, 30*time.Second,
		informers.WithNamespace(k8s.CombinatorNamespace),
	)
	combinatorSubHandler := cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.combinator.onSubResourceDelete,
	}
	combinatorK8sFactory.Apps().V1().Deployments().Informer().AddEventHandler(combinatorSubHandler)
	combinatorK8sFactory.Core().V1().Services().Informer().AddEventHandler(combinatorSubHandler)

	log.Println("[controller] starting informers")
	go dynFactory.Start(stopCh)
	go k8sFactory.Start(stopCh)
	go ingressDynFactory.Start(stopCh)
	go combinatorDynFactory.Start(stopCh)
	go combinatorK8sFactory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, crInformer.HasSynced, combinatorCrInformer.HasSynced) {
		log.Println("[controller] failed to sync CR informer cache")
		return
	}
	log.Println("[controller] informer cache synced")
}

func (c *Controller) updateStatus(u *unstructured.Unstructured, gvr schema.GroupVersionResource, phase, message string) {
	client := c.client.Resource(gvr).Namespace(u.GetNamespace())

	latest, err := client.Get(context.Background(), u.GetName(), metav1.GetOptions{})
	if err != nil {
		log.Printf("[controller] get latest %s for status update failed: %v", u.GetName(), err)
		return
	}

	if latest.Object["status"] == nil {
		latest.Object["status"] = map[string]interface{}{}
	}
	status := latest.Object["status"].(map[string]interface{})
	status["phase"] = phase
	status["message"] = message

	_, err = client.UpdateStatus(context.Background(), latest, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("[controller] update status for %s failed: %v", u.GetName(), err)
	}
}
