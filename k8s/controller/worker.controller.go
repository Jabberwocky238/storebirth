package controller

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	"jabberwocky238/console/k8s"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/dynamic/dynamicinformer"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type WorkerAppController struct {
	client    dynamic.Interface
	k8sClient *kubernetes.Clientset
	crCache   cache.Store
}

func NewController(client dynamic.Interface, k8sClient *kubernetes.Clientset) *WorkerAppController {
	return &WorkerAppController{client: client, k8sClient: k8sClient}
}

func (c *WorkerAppController) Start(stopCh <-chan struct{}) {
	// 1. CR informer: watch WorkerApp in worker namespace
	dynFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		c.client, 30*time.Second, k8s.WorkerNamespace, nil,
	)
	crInformer := dynFactory.ForResource(WorkerAppGVR).Informer()
	c.crCache = crInformer.GetStore()

	crInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    c.onAdd,
		UpdateFunc: c.onUpdate,
		DeleteFunc: c.onDelete,
	})

	// 2. Sub-resource informer: Deployment + Service + ConfigMap + Secret in worker namespace
	k8sFactory := informers.NewSharedInformerFactoryWithOptions(
		c.k8sClient, 30*time.Second,
		informers.WithNamespace(k8s.WorkerNamespace),
	)
	subHandler := cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.onSubResourceDelete,
	}
	k8sFactory.Apps().V1().Deployments().Informer().AddEventHandler(subHandler)
	k8sFactory.Core().V1().Services().Informer().AddEventHandler(subHandler)

	// Watch ConfigMap and Secret updates to trigger Deployment rolling restart
	configHandler := cache.ResourceEventHandlerFuncs{
		UpdateFunc: c.onConfigUpdate,
	}
	k8sFactory.Core().V1().ConfigMaps().Informer().AddEventHandler(configHandler)
	k8sFactory.Core().V1().Secrets().Informer().AddEventHandler(configHandler)

	// 3. IngressRoute informer: watch IngressRoute in ingress namespace
	ingressDynFactory := dynamicinformer.NewFilteredDynamicSharedInformerFactory(
		c.client, 30*time.Second, k8s.IngressNamespace, nil,
	)
	irInformer := ingressDynFactory.ForResource(k8s.IngressRouteGVR).Informer()
	irInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		DeleteFunc: c.onSubResourceDelete,
	})

	log.Println("[controller] starting informers")
	go dynFactory.Start(stopCh)
	go k8sFactory.Start(stopCh)
	go ingressDynFactory.Start(stopCh)

	if !cache.WaitForCacheSync(stopCh, crInformer.HasSynced) {
		log.Println("[controller] failed to sync CR informer cache")
		return
	}
	log.Println("[controller] informer cache synced")
}

// --- CR event handlers ---

func (c *WorkerAppController) onAdd(obj interface{}) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] WorkerApp added: %s", u.GetName())
	c.reconcile(u)
}

func (c *WorkerAppController) onUpdate(oldObj, newObj interface{}) {
	u, ok := newObj.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] WorkerApp updated: %s", u.GetName())
	c.reconcile(u)
}

func (c *WorkerAppController) onDelete(obj interface{}) {
	u, ok := obj.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] WorkerApp deleted: %s", u.GetName())

	w := workerFromUnstructured(u)
	if w == nil {
		return
	}
	w.DeleteAll(context.Background())
}

// --- Sub-resource delete handler ---

func (c *WorkerAppController) onSubResourceDelete(obj interface{}) {
	// Extract "app" label from the deleted sub-resource
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

	// Find the parent CR in cache
	key := k8s.WorkerNamespace + "/" + appName
	item, exists, err := c.crCache.GetByKey(key)
	if err != nil || !exists {
		return
	}
	u, ok := item.(*unstructured.Unstructured)
	if !ok {
		return
	}
	log.Printf("[controller] sub-resource deleted for %s, re-reconciling", appName)
	c.reconcile(u)
}

// --- ConfigMap / Secret update handler ---

func (c *WorkerAppController) onConfigUpdate(oldObj, newObj interface{}) {
	old, ok1 := oldObj.(metav1.Object)
	cur, ok2 := newObj.(metav1.Object)
	if !ok1 || !ok2 {
		return
	}
	// Skip if resourceVersion unchanged (re-list, not a real update)
	if old.GetResourceVersion() == cur.GetResourceVersion() {
		return
	}
	appName := cur.GetLabels()["app"]
	if appName == "" {
		return
	}
	log.Printf("[controller] config/secret updated for %s, restarting deployment", appName)
	c.restartDeployment(appName)
}

func (c *WorkerAppController) restartDeployment(name string) {
	if k8s.K8sClient == nil {
		return
	}
	patch := fmt.Sprintf(
		`{"spec":{"template":{"metadata":{"annotations":{"console.app238.com/restartedAt":"%s"}}}}}`,
		strconv.FormatInt(time.Now().Unix(), 10),
	)
	_, err := k8s.K8sClient.AppsV1().Deployments(k8s.WorkerNamespace).Patch(
		context.Background(), name, types.StrategicMergePatchType,
		[]byte(patch), metav1.PatchOptions{},
	)
	if err != nil {
		log.Printf("[controller] restart deployment %s failed: %v", name, err)
	}
}

// --- Reconcile: ensure all 5 sub-resources exist ---

func (c *WorkerAppController) reconcile(u *unstructured.Unstructured) {
	w := workerFromUnstructured(u)
	if w == nil {
		return
	}

	ctx := context.Background()
	c.updateStatus(u, "Deploying", "")

	if err := w.EnsureConfigMap(ctx); err != nil {
		log.Printf("[controller] ensure configmap for %s failed: %v", u.GetName(), err)
		c.updateStatus(u, "Failed", err.Error())
		return
	}
	if err := w.EnsureSecret(ctx); err != nil {
		log.Printf("[controller] ensure secret for %s failed: %v", u.GetName(), err)
		c.updateStatus(u, "Failed", err.Error())
		return
	}
	if err := w.EnsureDeployment(ctx); err != nil {
		log.Printf("[controller] ensure deployment for %s failed: %v", u.GetName(), err)
		c.updateStatus(u, "Failed", err.Error())
		return
	}
	if err := w.EnsureService(ctx); err != nil {
		log.Printf("[controller] ensure service for %s failed: %v", u.GetName(), err)
		c.updateStatus(u, "Failed", err.Error())
		return
	}
	if err := w.EnsureIngressRoute(ctx); err != nil {
		log.Printf("[controller] ensure ingress route for %s failed: %v", u.GetName(), err)
		c.updateStatus(u, "Failed", err.Error())
		return
	}

	log.Printf("[controller] reconcile %s success", u.GetName())
	c.updateStatus(u, "Running", "")
}

func (c *WorkerAppController) updateStatus(u *unstructured.Unstructured, phase, message string) {
	patch := u.DeepCopy()
	if patch.Object["status"] == nil {
		patch.Object["status"] = map[string]interface{}{}
	}
	status := patch.Object["status"].(map[string]interface{})
	status["phase"] = phase
	status["message"] = message

	client := c.client.Resource(WorkerAppGVR).Namespace(u.GetNamespace())
	_, err := client.UpdateStatus(context.Background(), patch, metav1.UpdateOptions{})
	if err != nil {
		log.Printf("[controller] update status for %s failed: %v", u.GetName(), err)
	}
}

// --- Helpers ---

func workerFromUnstructured(u *unstructured.Unstructured) *Worker {
	spec, _ := u.Object["spec"].(map[string]interface{})
	if spec == nil {
		return nil
	}
	port, _ := spec["port"].(int64)
	return &Worker{
		WorkerID: fmt.Sprintf("%v", spec["workerID"]),
		OwnerID:  fmt.Sprintf("%v", spec["ownerID"]),
		Image:    fmt.Sprintf("%v", spec["image"]),
		Port:     int(port),
	}
}

// --- CR CRUD (used by handlers) ---

func CreateWorkerAppCR(
	client dynamic.Interface,
	name, workerID, ownerID, image string,
	port int,
) error {
	cr := &unstructured.Unstructured{
		Object: map[string]interface{}{
			"apiVersion": Group + "/" + Version,
			"kind":       WorkerKind,
			"metadata": map[string]interface{}{
				"name":      name,
				"namespace": k8s.WorkerNamespace,
			},
			"spec": map[string]interface{}{
				"workerID": workerID,
				"ownerID":  ownerID,
				"image":    image,
				"port":     int64(port),
			},
		},
	}

	ctx := context.Background()
	res := client.Resource(WorkerAppGVR).Namespace(k8s.WorkerNamespace)

	existing, err := res.Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		_, err = res.Create(ctx, cr, metav1.CreateOptions{})
		return err
	}

	cr.SetResourceVersion(existing.GetResourceVersion())
	_, err = res.Update(ctx, cr, metav1.UpdateOptions{})
	return err
}

func DeleteWorkerAppCR(client dynamic.Interface, name string) error {
	return client.Resource(WorkerAppGVR).
		Namespace(k8s.WorkerNamespace).
		Delete(context.Background(), name, metav1.DeleteOptions{})
}
