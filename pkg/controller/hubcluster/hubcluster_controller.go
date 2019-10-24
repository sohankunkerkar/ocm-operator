package hubcluster

import (
	"context"
	"flag"
	"strings"

	mf "github.com/jcrossley3/manifestival"
	"github.com/operator-framework/operator-sdk/pkg/predicate"
	onpremv1alpha1 "github.com/sohankunkerkar/onprem-operator/pkg/apis/onprem/v1alpha1"
	"github.com/sohankunkerkar/onprem-operator/version"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var (
	filename = flag.String("filename", "deploy/resources",
		"The filename containing the YAML resources to apply")
	recursive = flag.Bool("recursive", false,
		"If filename is a directory, process all manifests recursively")
	namespace = flag.String("namespace", "",
		"Overrides namespace in manifest (env vars resolved in-container)")
	log = logf.Log.WithName("controller_hubcluster")
)

// Add creates a new HubCluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	manifest, err := mf.NewManifest(*filename, *recursive, mgr.GetClient())
	if err != nil {
		return err
	}
	return add(mgr, newReconciler(mgr, manifest))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager, man mf.Manifest) reconcile.Reconciler {
	return &ReconcileHubCluster{client: mgr.GetClient(), scheme: mgr.GetScheme(), config: man}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("hubcluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource HubCluster
	err = c.Watch(&source.Kind{Type: &onpremv1alpha1.HubCluster{}}, &handler.EnqueueRequestForObject{}, predicate.GenerationChangedPredicate{})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileHubCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileHubCluster{}

// ReconcileHubCluster reconciles a HubCluster object
type ReconcileHubCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	config mf.Manifest
}

// Reconcile reads that state of the cluster for a HubCluster object and makes changes based on the state read
// and what is in the HubCluster.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileHubCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling HubCluster")

	// Fetch the HubCluster instance
	instance := &onpremv1alpha1.HubCluster{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			r.config.DeleteAll()
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}
	stages := []func(*onpremv1alpha1.HubCluster) error{
		r.install,
	}

	for _, stage := range stages {
		if err := stage(instance); err != nil {
			return reconcile.Result{}, err
		}
	}

	reqLogger.Info("Finished reconciling hubcluster")

	return reconcile.Result{}, nil
}

// This is a transform method that updates the namespace field of the clusterrolebinding/rolebinding resource
func resourceNamespaceUpdate(ns, name string) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		reqLogger := log.WithValues("Instance.Namespace", ns, "Instance.Name", name)
		kind := strings.ToLower(u.GetKind())
		if kind == "clusterrolebinding" || kind == "rolebinding" {
			if subjects, ok, _ := unstructured.NestedSlice(u.Object, "subjects"); ok {
				if subjectKind, ok, err := unstructured.NestedString(subjects[0].(map[string]interface{}), "kind"); ok && subjectKind == "ServiceAccount" {
					err = unstructured.SetNestedField(subjects[0].(map[string]interface{}), ns, "namespace")
					if err != nil {
						reqLogger.Info("Failed to set the namespace nested field")
					} else {
						reqLogger.Info("Added the namespace to the clusterrolebinding subjects element")
						err = unstructured.SetNestedSlice(u.Object, subjects, "subjects")
						if err != nil {
							reqLogger.Info("Failed to update the subjects slice")
						}
					}
				} else {
					reqLogger.Info("Failed to get the kind of the subject or the kind is not of interest")
				}
			} else {
				reqLogger.Info("Failed to get subjects slice")
			}
		}
		return nil
	}
}

func overWriteNamespace(ns, name string) mf.Transformer {
	return func(u *unstructured.Unstructured) error {
		reqLogger := log.WithValues("Instance.Namespace", ns, "Instance.Name", name)
		kind := strings.ToLower(u.GetKind())
		label := u.GetLabels()
		var res string = label["resource"]
		if (kind == "namespace") && (res == "joinedcluster") {
			reqLogger.Info("Updating the namespace to onprem-hub-system")
			u.SetName("onprem-hub-system")
		}
		return nil
	}
}

// Apply the embedded resources
func (r *ReconcileHubCluster) install(instance *onpremv1alpha1.HubCluster) error {
	// Transform resources as appropriate
	fns := []mf.Transformer{mf.InjectOwner(instance)}
	fns = append(fns, mf.InjectNamespace(instance.Namespace))
	fns = append(fns, resourceNamespaceUpdate(instance.Namespace, instance.Name))
	fns = append(fns, overWriteNamespace(instance.Namespace, instance.Name))
	r.config.Transform(fns...)

	// Apply the resources in the YAML file
	if err := r.config.ApplyAll(); err != nil {
		return err
	}

	// Update status
	instance.Status.Version = version.Version
	if err := r.client.Status().Update(context.TODO(), instance); err != nil {
		return err
	}
	log.Info("Install succeeded", "version", version.Version)

	return nil
}
