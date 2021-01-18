package firelite

import (
	"context"
	"fmt"

	litev1alpha1 "github.com/worroc/operator-lite/pkg/apis/lite/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_firelite")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new FireLite Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileFireLite{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("firelite-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource FireLite
	err = c.Watch(&source.Kind{Type: &litev1alpha1.FireLite{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner FireLite
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &litev1alpha1.FireLite{},
	})
	if err != nil {
		return err
	}

	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &litev1alpha1.FireLite{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that   implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileFireLite{}

// ReconcileFireLite reconciles a FireLite object
type ReconcileFireLite struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a FireLite object and makes changes based on the state read
// and what is in the FireLite.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  This example creates
// a Pod as an example
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileFireLite) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling FireLite")

	// Fetch the FireLite instance
	instance := &litev1alpha1.FireLite{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	// Define a new Pod object
	// pod := newPodForCR(instance)

	// Check if this Deployment already exists
	found := &appsv1.Deployment{}
	err = r.client.Get(context.TODO(), types.NamespacedName{Name: instance.Name, Namespace: instance.Namespace}, found)

	var result *reconcile.Result

	result, err = r.lineupService(request, instance, r.backendService(instance))
	if result != nil {
		return *result, err
	}

	result, err = r.lineupDeployment(request, instance, r.backendDeployment(instance))
	if result != nil {
		return *result, err
	}

	// Deployment and Service already exists - don't requeue
	reqLogger.Info("Deployment and service lineup done",
		"Deployment.Namespace", found.Namespace, "Deployment.Name", found.Name)
	return reconcile.Result{}, nil

	// // Set FireLite instance as the owner and controller
	// if err := controllerutil.SetControllerReference(instance, pod, r.scheme); err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Check if this Pod already exists
	// found := &corev1.Pod{}
	// err = r.client.Get(context.TODO(), types.NamespacedName{Name: pod.Name, Namespace: pod.Namespace}, found)
	// if err != nil && errors.IsNotFound(err) {
	// 	reqLogger.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
	// 	err = r.client.Create(context.TODO(), pod)
	// 	if err != nil {
	// 		return reconcile.Result{}, err
	// 	}

	// 	// Pod created successfully - don't requeue
	// 	return reconcile.Result{}, nil
	// } else if err != nil {
	// 	return reconcile.Result{}, err
	// }

	// // Pod already exists - don't requeue
	// reqLogger.Info("Skip reconcile: Pod already exists", "Pod.Namespace", found.Namespace, "Pod.Name", found.Name)
	// return reconcile.Result{}, nil
}

// // newPodForCR returns a busybox pod with the same name/namespace as the cr
// func newPodForCR(cr *litev1alpha1.FireLite) *corev1.Pod {
// 	labels := map[string]string{
// 		"app": cr.Name,
// 	}
// 	return &corev1.Pod{
// 		ObjectMeta: metav1.ObjectMeta{
// 			Name:      cr.Name + "-pod",
// 			Namespace: cr.Namespace,
// 			Labels:    labels,
// 		},
// 		Spec: corev1.PodSpec{
// 			Containers: []corev1.Container{
// 				{
// 					Name:    "busybox",
// 					Image:   "busybox",
// 					Command: []string{"sleep", "3600"},
// 				},
// 			},
// 		},
// 	}
// }

func labels(v *litev1alpha1.FireLite, tier string) map[string]string {
	// Fetches and sets labels

	return map[string]string{
		"app":         "firelite",
		"firelite_cr": v.Name,
		"tier":        tier,
	}
}

func (r *ReconcileFireLite) backendService(v *litev1alpha1.FireLite) *corev1.Service {
	// Build a Service and deploys

	labels := labels(v, "backend")

	s := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      v.Spec.NickName + "-backend-service",
			Namespace: v.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector: labels,
			Ports: []corev1.ServicePort{{
				Protocol:   corev1.ProtocolTCP,
				Port:       80,
				TargetPort: intstr.FromInt(int(v.Spec.Port)),
				// 30000-32767
				NodePort: 30000 + (v.Spec.Port % 1000),
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	controllerutil.SetControllerReference(v, s, r.scheme)
	return s
}

func (r *ReconcileFireLite) backendDeployment(v *litev1alpha1.FireLite) *appsv1.Deployment {
	// Build a Deployment

	labels := labels(v, "backend")
	size := v.Spec.Size
	port := v.Spec.Port
	nickName := v.Spec.NickName

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      nickName + "-pod",
			Namespace: v.Namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: &size,
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: labels,
				},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Image:           v.Spec.Image,
						ImagePullPolicy: corev1.PullAlways,
						Name:            nickName + "-pod",
						Env: []corev1.EnvVar{{
							Name:  "PORT",
							Value: fmt.Sprintf("%d", port),
						}},
						Ports: []corev1.ContainerPort{{
							ContainerPort: port,
							Name:          nickName,
							Protocol:      corev1.ProtocolTCP,
						}},
					}},
				},
			},
		},
	}

	controllerutil.SetControllerReference(v, dep, r.scheme)
	return dep
}

func (r *ReconcileFireLite) lineupService(request reconcile.Request,
	instance *litev1alpha1.FireLite,
	s *corev1.Service,
) (*reconcile.Result, error) {

	log.Info("harmonize Service")

	// See if service already exists and create if it doesn't
	found := &corev1.Service{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      s.Name,
		Namespace: instance.Namespace,
	}, found)

	log.Info("found", "value", found, "err", err)
	if err != nil && errors.IsNotFound(err) {
		res, err := r.createService(s)
		if err != nil {
			return res, err
		}
	} else if err != nil {
		// Error that isn't due to the service not existing
		log.Error(err, "Failed to get Service")
		return &reconcile.Result{}, err
	} else {
		// service exists lets check if we have to change it
		log.Info("service",
			"desired",
			fmt.Sprintf("%d", s.Spec.Ports[0].NodePort),
			"current",
			fmt.Sprintf("%d", found.Spec.Ports[0].NodePort),
		)
		if s.Spec.Ports[0].NodePort != found.Spec.Ports[0].NodePort {
			found.Spec.Ports[0].NodePort = s.Spec.Ports[0].NodePort
			res, err := r.updateService(found)
			if err != nil {
				return res, err
			}
		}
	}

	return nil, nil
}

func (r *ReconcileFireLite) createService(s *corev1.Service) (*reconcile.Result, error) {
	// Create the service
	log.Info("Creating a new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
	err := r.client.Create(context.TODO(), s)

	if err != nil {
		// Service creation failed
		log.Error(err, "Failed to create new Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		return &reconcile.Result{}, err
	}
	// Service creation was successful
	return nil, nil
}
func (r *ReconcileFireLite) updateService(s *corev1.Service) (*reconcile.Result, error) {
	// Create the service
	log.Info("Updating a Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
	err := r.client.Update(context.TODO(), s)

	if err != nil {
		// Service creation failed
		log.Error(err, "Failed to update Service", "Service.Namespace", s.Namespace, "Service.Name", s.Name)
		return &reconcile.Result{}, err
	}
	// Service updates was successfuly
	return nil, nil
}

func (r *ReconcileFireLite) lineupDeployment(request reconcile.Request,
	instance *litev1alpha1.FireLite,
	dep *appsv1.Deployment,
) (*reconcile.Result, error) {
	log.Info("lineup Deployment")
	// See if deployment already exists and create if it doesn't
	found := &appsv1.Deployment{}
	err := r.client.Get(context.TODO(), types.NamespacedName{
		Name:      dep.Name,
		Namespace: instance.Namespace,
	}, found)
	if err != nil && errors.IsNotFound(err) {
		res, err := r.createDeployment(dep)
		if err != nil {
			return res, err
		}
	} else if err != nil {
		// Error that isn't due to the deployment not existing
		log.Error(err, "Failed to get Deployment")
		return &reconcile.Result{}, err
	} else {
		replicas := instance.Spec.Size
		found.Spec.Replicas = &replicas
		found.Spec.Template.Spec.Containers[0].Ports[0].ContainerPort = instance.Spec.Port
		found.Spec.Template.Spec.Containers[0].Image = instance.Spec.Image
		found.Spec.Template.Spec.Containers[0].Env[0].Value = fmt.Sprintf("%d", instance.Spec.Port)
		res, err := r.updateDeployment(found)
		if err != nil {
			return res, err
		}
	}

	return nil, nil
}
func (r *ReconcileFireLite) createDeployment(
	dep *appsv1.Deployment,
) (*reconcile.Result, error) {

	// Create the deployment
	log.Info("Creating a new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
	err := r.client.Create(context.TODO(), dep)

	if err != nil {
		// Deployment failed
		log.Error(err, "Failed to create new Deployment", "Deployment.Namespace", dep.Namespace, "Deployment.Name", dep.Name)
		return &reconcile.Result{}, err
	}
	// Deployment was successful
	return nil, nil
}

func (r *ReconcileFireLite) updateDeployment(
	dep *appsv1.Deployment,
) (*reconcile.Result, error) {
	err := r.client.Update(context.TODO(), dep)
	if err != nil {
		return &reconcile.Result{}, err
	}
	return nil, nil
}
