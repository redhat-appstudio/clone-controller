/*
Copyright 2023.

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

package controllers

import (
	"context"
	"fmt"

	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	ctrllog "sigs.k8s.io/controller-runtime/pkg/log"

	appstudioredhatcomv1alpha1 "github.com/redhat-appstudio/clone-controller/api/v1alpha1"
	integrationtestapi "github.com/redhat-appstudio/integration-service/api/v1beta1"

	hasApplicationAPI "github.com/redhat-appstudio/application-api/api/v1alpha1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ApplicationCloneReconciler reconciles a ApplicationClone object
type ApplicationCloneReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=appstudio.redhat.com,resources=applicationclones,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=appstudio.redhat.com,resources=applicationclones/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=appstudio.redhat.com,resources=applicationclones/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the ApplicationClone object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *ApplicationCloneReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = ctrllog.FromContext(ctx)

	// TODO(user): your logic here

	log := ctrllog.FromContext(ctx).WithName("ApplicationClone")

	ctx = ctrllog.IntoContext(ctx, log)

	applicationClone := &appstudioredhatcomv1alpha1.ApplicationClone{}

	err := r.Client.Get(ctx, req.NamespacedName, applicationClone)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, fmt.Errorf("error reading resource: %w", err)
	}

	err = r.Client.Create(ctx, &hasApplicationAPI.Application{
		ObjectMeta: metav1.ObjectMeta{
			Name:      applicationClone.Spec.From.Name,
			Namespace: applicationClone.Namespace,
		},
		Spec: hasApplicationAPI.ApplicationSpec{
			DisplayName: "appfoo",
		},
	})

	if err != nil {
		return ctrl.Result{}, fmt.Errorf("error creating application %v", err)
	}

	log.Info("successfully created Application CR ", applicationClone.Spec.From.Namespace, applicationClone.Name)

	hasComponentList := &hasApplicationAPI.ComponentList{}
	err = r.Client.List(ctx, hasComponentList, &client.ListOptions{Namespace: applicationClone.Spec.From.Namespace})
	if err != nil {
		log.Error(err, "Error listing components")
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, fmt.Errorf("error reading resource: %w", err)
	}

	var componentToBeCloned hasApplicationAPI.ComponentList

	// Check if the Components belong to the relevant Application
	for _, c := range hasComponentList.Items {
		if c.Spec.Application != applicationClone.Spec.From.Name {
			continue
		}
		// take everything that's associated with the Application *in*
		log.Info("found Component", c.Namespace, c.Name)
		componentToBeCloned.Items = append(componentToBeCloned.Items, c)
	}

	// Create the Components

	for _, c := range componentToBeCloned.Items {
		// determine if this is the "source" component or the "image component"
		for _, sourceComponent := range applicationClone.Spec.ComponentSources {
			if c.Name == sourceComponent.Name {

				// Create a new Component without specifying the image.

				err = r.Client.Create(ctx, &hasApplicationAPI.Component{
					ObjectMeta: metav1.ObjectMeta{
						Name:      c.Name,
						Namespace: applicationClone.Namespace,
						Annotations: map[string]string{
							"skip-initial-checks":       "true",
							"image.redhat.com/generate": `{"visibility": "public"}`,
						},
					},
					Spec: hasApplicationAPI.ComponentSpec{
						Application:   applicationClone.Spec.From.Name,
						ComponentName: c.Spec.ComponentName,
						Source: hasApplicationAPI.ComponentSource{
							ComponentSourceUnion: hasApplicationAPI.ComponentSourceUnion{
								GitSource: &hasApplicationAPI.GitSource{
									URL:           c.Spec.Source.GitSource.URL,
									Context:       c.Spec.Source.GitSource.Revision,
									Revision:      c.Spec.Source.GitSource.Revision,
									DockerfileURL: c.Spec.Source.GitSource.DockerfileURL,
								},
							},
						},
						Replicas:   c.Spec.Replicas,
						Resources:  c.Spec.Resources,
						Env:        c.Spec.Env,
						TargetPort: c.Spec.TargetPort,
					},
				})

				if err != nil {
					log.Error(err, "error creating Component")
				} else {
					log.Info("created component from Source", c.Name, c.Namespace, "Source", c.Spec.Source.GitSource.URL)
				}
			} else {
				// Create a new Component with the image reference.

				r.Client.Create(ctx, &hasApplicationAPI.Component{
					ObjectMeta: metav1.ObjectMeta{
						Name:      c.Name,
						Namespace: applicationClone.Namespace,
						Annotations: map[string]string{
							"skip-initial-checks": "true",
						},
					},
					Spec: hasApplicationAPI.ComponentSpec{
						Application:    applicationClone.Spec.From.Name,
						ComponentName:  c.Spec.ComponentName,
						Replicas:       c.Spec.Replicas,
						Resources:      c.Spec.Resources,
						Env:            c.Spec.Env,
						TargetPort:     c.Spec.TargetPort,
						ContainerImage: c.Spec.ContainerImage,
					},
				})

				if err != nil {
					log.Error(err, "error creating Component")
				}
				log.Info("created component with Image Reference", c.Namespace, c.Name, "image", c.Spec.ContainerImage)
			}
		}
	}

	// Setup the Integration Tests
	testsList := &integrationtestapi.IntegrationTestScenarioList{}

	err = r.Client.List(ctx, testsList, &client.ListOptions{Namespace: applicationClone.Spec.From.Namespace})
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return ctrl.Result{}, fmt.Errorf("error reading resource: %w", err)
	}

	for _, integrationTest := range testsList.Items {
		if integrationTest.Spec.Application == applicationClone.Name {
			err = r.Client.Create(ctx, &integrationtestapi.IntegrationTestScenario{
				ObjectMeta: metav1.ObjectMeta{
					Name:      integrationTest.Name,
					Namespace: applicationClone.Namespace,
				},
				Spec: integrationtestapi.IntegrationTestScenarioSpec{
					Application: integrationTest.Spec.Application,
					ResolverRef: integrationTest.Spec.ResolverRef,
					Params:      integrationTest.Spec.Params,
					Environment: integrationTest.Spec.Environment,
					Contexts:    integrationTest.Spec.Contexts,
				},
			})
			if err != nil {
				log.Error(err, "error creating integrationtestscenario", "application", applicationClone, "test", integrationTest)
				return ctrl.Result{}, nil
			}
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *ApplicationCloneReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&appstudioredhatcomv1alpha1.ApplicationClone{}).
		Complete(r)
}
