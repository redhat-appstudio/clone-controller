# The ApplicationClone controller
A controller which clones an Application ( and all associated resources ) from one namespace to another.

*This is a work-in-progress proof-of-concept.*

## Description

A user creates the following CR in the namespace where she would 
like to have resources copied to:

```
apiVersion: appstudio.redhat.com/v1alpha1
kind: ApplicationClone
metadata:
  name: applicationclone-sample
  namepace: target-ns
spec:
  from:
    namespace: source-ns
    name: billing-app # name of the Application CR
  componentSources:
    - name: component-a
    - name: component-b
```

After verifying if the requesting actor is authorized to read resources from the namespace from which the `Application`` is
being cloned (TODO), the reconciler copies over the following resources to the namespace where the `ApplicationClone` CR was created.

* The `Application` CR,
* The `Component` CRs and 
* The `IntegrationTestScenario` CRs.
* The `Secrets`, where relevant. (TODO)

The `Components` listed in `.spec.componentSources` are copied over in the new namespace with the intent to be built from source into an image. The rest of the `Components` in the `Application` are imported using their image references.


Defining the intent to clone as a Kubernetes custom resources gives us the ability to store 'status' information associated with the the cloning in the `.status` resource.

```   
apiVersion: appstudio.redhat.com/v1alpha1
kind: ApplicationClone
metadata:
  name: applicationclone-sample
spec:
  from:
    namespace: source-ns
    name: billing-app
  componentSources:
    - name: component-a
    - name: component-b
status:
  lastAttempt: 14:23 APR 21 2023
  lastSuccessfulAttempt: 14:23 APR 21 2023
  resources:
    - Kind: Component
      Name: Component-a
    - Kind: Component
      Name: Component-b
    - Kind: Component
      Name: Component-c
    - Kind: Component
      Name: Component-d
    - Kind: IntegrationTestScenario
      Name: test-1
    - Kind: IntegrationTestScenario
      Name: test-2
    - Kind: Secret
      Name: pull-secret-1
      
```
## Scenarios

* Clone Application with two Components to be built from source.


```
apiVersion: appstudio.redhat.com/v1alpha1
kind: ApplicationClone
metadata:
  name: applicationclone-sample
  namespace: target-ns
spec:
  from:
    namespace: source-ns
    name: billing-app
  componentSources:
    - name: component-a
    - name: component-b
```


* Clone Application with all Components to be built from source

```
apiVersion: appstudio.redhat.com/v1alpha1
kind: ApplicationClone
metadata:
  name: applicationclone-sample
  namespace: target-ns
spec:
  from:
    namespace: source-ns
    name: billing-app
  componentSources:
    - name: *
```

## Development 
You’ll need a Kubernetes cluster to run against. You can use [KIND](https://sigs.k8s.io/kind) to get a local cluster for testing, or run against a remote cluster.
**Note:** Your controller will automatically use the current context in your kubeconfig file (i.e. whatever cluster `kubectl cluster-info` shows).

### Running on the cluster
1. Install Instances of Custom Resources:

```sh
kubectl apply -f config/samples/
```

2. Build and push your image to the location specified by `IMG`:

```sh
make docker-build docker-push IMG=<some-registry>/applicationclone:tag
```

3. Deploy the controller to the cluster with the image specified by `IMG`:

```sh
make deploy IMG=<some-registry>/applicationclone:tag
```

### Uninstall CRDs
To delete the CRDs from the cluster:

```sh
make uninstall
```

### Undeploy controller
UnDeploy the controller from the cluster:

```sh
make undeploy
```

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

### How it works
This project aims to follow the Kubernetes [Operator pattern](https://kubernetes.io/docs/concepts/extend-kubernetes/operator/).

It uses [Controllers](https://kubernetes.io/docs/concepts/architecture/controller/),
which provide a reconcile function responsible for synchronizing resources until the desired state is reached on the cluster.

### Test It Out
1. Install the CRDs into the cluster:

```sh
make install
```

2. Run your controller (this will run in the foreground, so switch to a new terminal if you want to leave it running):

```sh
make run
```

**NOTE:** You can also run this in one step by running: `make install run`

### Modifying the API definitions
If you are editing the API definitions, generate the manifests such as CRs or CRDs using:

```sh
make manifests
```

**NOTE:** Run `make --help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

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

