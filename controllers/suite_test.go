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
	"go/build"
	"path/filepath"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/envtest"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"

	hasApplicationAPI "github.com/redhat-appstudio/application-api/api/v1alpha1"
	appstudioredhatcomv1alpha1 "github.com/redhat-appstudio/clone-controller/api/v1alpha1"
	integrationtestapi "github.com/redhat-appstudio/integration-service/api/v1beta1"
	corev1 "k8s.io/api/core/v1"
	k8sErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"

	ctrl "sigs.k8s.io/controller-runtime"
	//+kubebuilder:scaffold:imports
)

// These tests use Ginkgo (BDD-style Go testing framework). Refer to
// http://onsi.github.io/ginkgo/ to learn more about Ginkgo.

const (
	// timeout is used as a limit until condition become true
	// Usually used in Eventually statements
	timeout = time.Second * 15
	// ensureTimeout is used as a period of time during which the condition should not be changed
	// Usually used in Consistently statements
	ensureTimeout = time.Second * 4
	interval      = time.Millisecond * 250
)

var (
	cancel context.CancelFunc
	ctx    context.Context
)

var cfg *rest.Config
var k8sClient client.Client
var testEnv *envtest.Environment

func TestAPIs(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Controller Suite")
}

var _ = BeforeSuite(func() {
	logf.SetLogger(zap.New(zap.WriteTo(GinkgoWriter), zap.UseDevMode(true)))

	By("bootstrapping test environment")

	applicationApiDepVersion := "v0.0.0-20221220162402-c1e887791dac"
	integrationAPIDepVersion := "v0.0.0-20230724115413-9e84440dc538"

	testEnv = &envtest.Environment{
		CRDDirectoryPaths: []string{
			filepath.Join("..", "config", "crd", "bases"),
			filepath.Join(build.Default.GOPATH, "pkg", "mod", "github.com", "redhat-appstudio", "application-api@"+applicationApiDepVersion, "config", "crd", "bases"),

			// integration-service/api at main · redhat-appstudio/integration-service
			filepath.Join(build.Default.GOPATH, "pkg", "mod", "github.com", "redhat-appstudio", "integration-service@"+integrationAPIDepVersion, "config", "crd", "bases"),
		},
		ErrorIfCRDPathMissing: true,
	}

	ctx, cancel = context.WithCancel(context.TODO())

	var err error
	// cfg is defined in this file globally.
	cfg, err = testEnv.Start()
	Expect(err).NotTo(HaveOccurred())
	Expect(cfg).NotTo(BeNil())

	err = appstudioredhatcomv1alpha1.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = hasApplicationAPI.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	err = integrationtestapi.AddToScheme(scheme.Scheme)
	Expect(err).NotTo(HaveOccurred())

	//+kubebuilder:scaffold:scheme

	k8sClient, err = client.New(cfg, client.Options{Scheme: scheme.Scheme})
	Expect(err).NotTo(HaveOccurred())
	Expect(k8sClient).NotTo(BeNil())

	k8sManager, err := ctrl.NewManager(cfg, ctrl.Options{
		Scheme: scheme.Scheme,
	})
	Expect(err).ToNot(HaveOccurred())

	err = (&ApplicationCloneReconciler{
		Client: k8sManager.GetClient(),
		Scheme: k8sManager.GetScheme(),
	}).SetupWithManager(k8sManager)
	Expect(err).ToNot(HaveOccurred())

	go func() {
		defer GinkgoRecover()
		err = k8sManager.Start(ctx)
		Expect(err).ToNot(HaveOccurred(), "failed to run manager")
	}()

})

var _ = Describe("ApplicationClone controller", func() {

	Context("When creating an ApplicationClone", func() {
		It("Should create the Application, Components & IntegrationTestScenario", func() {
			By("By creating a new Application CR, relevant Component CRs and Tests")
			ctx := context.Background()
			createNamespace(ctx, "foo")

			Expect(k8sClient.Create(ctx, &hasApplicationAPI.Application{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "appfoo",
					Namespace: "foo",
				},
				Spec: hasApplicationAPI.ApplicationSpec{
					DisplayName: "",
				},
			})).NotTo(HaveOccurred())

			// Add 2 components

			Expect(k8sClient.Create(ctx, &hasApplicationAPI.Component{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "c1",
					Namespace: "foo",
				},
				Spec: hasApplicationAPI.ComponentSpec{
					ComponentName: "component-a",
					Application:   "appfoo",
					Source: hasApplicationAPI.ComponentSource{
						ComponentSourceUnion: hasApplicationAPI.ComponentSourceUnion{
							GitSource: &hasApplicationAPI.GitSource{
								URL:           "github.com/foo/bar",
								DockerfileURL: "",
							},
						},
					},
					ContainerImage: "github.com/foo/c1",
				},
			},
			)).NotTo(HaveOccurred())

			Expect(k8sClient.Create(ctx, &hasApplicationAPI.Component{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "c2",
					Namespace: "foo",
				},
				Spec: hasApplicationAPI.ComponentSpec{
					ComponentName: "component-b",
					Application:   "appfoo",
					Source: hasApplicationAPI.ComponentSource{
						ComponentSourceUnion: hasApplicationAPI.ComponentSourceUnion{
							GitSource: &hasApplicationAPI.GitSource{
								URL:           "github.com/foo/c2",
								DockerfileURL: "",
							},
						},
					},
					ContainerImage: "quay.io/foo/c2",
				},
			},
			)).NotTo(HaveOccurred())

			// Add two Integration tests

			Expect(k8sClient.Create(ctx, &integrationtestapi.IntegrationTestScenario{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "it1",
					Namespace: "foo",
				},
				Spec: integrationtestapi.IntegrationTestScenarioSpec{
					Application: "appfoo",
					ResolverRef: integrationtestapi.ResolverRef{
						Resolver: "quay.io/pipeline/location",
						Params:   []integrationtestapi.ResolverParameter{},
					},
					Params: []integrationtestapi.PipelineParameter{},
				},
			},
			)).NotTo(HaveOccurred())

			Expect(k8sClient.Create(ctx, &integrationtestapi.IntegrationTestScenario{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "it2",
					Namespace: "foo",
				},
				Spec: integrationtestapi.IntegrationTestScenarioSpec{
					Application: "appfoo",
					ResolverRef: integrationtestapi.ResolverRef{
						Resolver: "quay.io/pipeline/location",
						Params:   []integrationtestapi.ResolverParameter{},
					},
					Params: []integrationtestapi.PipelineParameter{},
				},
			},
			)).NotTo(HaveOccurred())

			// Now, let's get cloning

			createNamespace(ctx, "bar")

			Expect(k8sClient.Create(ctx, &appstudioredhatcomv1alpha1.ApplicationClone{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "appfoo",
					Namespace: "bar",
				},
				Spec: appstudioredhatcomv1alpha1.ApplicationCloneSpec{
					From: appstudioredhatcomv1alpha1.From{
						Name:      "appfoo",
						Namespace: "foo",
					},
					ComponentSources: []appstudioredhatcomv1alpha1.ComponentSource{
						{
							Name: "c1",
						},
					},
				},
			})).NotTo(HaveOccurred())

			clonedApplication := &hasApplicationAPI.Application{}

			Eventually(func() bool {

				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "appfoo",
					Namespace: "bar",
				}, clonedApplication)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(clonedApplication.Name).NotTo(BeEmpty())

			clonedComponent := &hasApplicationAPI.Component{}
			Eventually(func() bool {
				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "c1",
					Namespace: "bar",
				}, clonedComponent)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(clonedComponent.Spec.ContainerImage).To(BeEmpty())
			Expect(clonedComponent.Spec.Source.GitSource.URL).NotTo(BeEmpty())

			clonedComponent = &hasApplicationAPI.Component{}
			Eventually(func() bool {

				err := k8sClient.Get(ctx, types.NamespacedName{
					Name:      "c2",
					Namespace: "bar",
				}, clonedComponent)
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(clonedComponent.Spec.ContainerImage).NotTo(BeEmpty())
			Expect(clonedComponent.Spec.Source.ComponentSourceUnion.GitSource).To(BeNil())

			// ensure the Integration tests show up.

			// Setup the Integration Tests

			testsList := &integrationtestapi.IntegrationTestScenarioList{}

			Eventually(func() bool {
				err := k8sClient.List(ctx, testsList, &client.ListOptions{Namespace: "bar"})
				return err == nil
			}, timeout, interval).Should(BeTrue())

			Expect(testsList.Items).To(HaveLen(2))

		})

	})
})

var _ = AfterSuite(func() {
	var err error
	By("tearing down the test environment")
	testEnv.Stop()
	Expect(err).NotTo(HaveOccurred())
})

func createNamespace(ctx context.Context, name string) {
	namespace := corev1.Namespace{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Namespace",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name,
		},
	}

	if err := k8sClient.Create(ctx, &namespace); err != nil && !k8sErrors.IsAlreadyExists(err) {
		Fail(err.Error())
	}
}
