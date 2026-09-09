# k8s-model-operator

Model operator will control synchronziations between Model kind and k8s resources needed to maintain it.

Model kind is a definition that will choose a model and a inference manager to be implemented in a k8s clusters in pursue of a self owned AI infrastructure.

Creation of this project is based on kubebuilder.

## Description

Model definition should be kept simple, to let different resources of the kind to read models from centralized volumes.

Current definition covers i.e. this example:

```bash
apiVersion: llmmodel.host-llm.io/v1alpha1
kind: Model
metadata:
  labels:
    app.kubernetes.io/name: k8s-model-operator
    app.kubernetes.io/managed-by: kustomize
  name: model-sample
  namespace: default
spec:
  provider: Ollama
  llmModel:
    name: qwen3
    version: 0.6b
  maxReplicas: 2
  minReplicas: 1
  targetNamespace: default
  persistentVolumeClaimRef: models-pvc
  mountPath: /data/models
  modelPath: qwen3-0.0b
```

The deinition chooses an `Ollama` provider/manager, using `qwen3:06b` model stored at a volume reachable by PersistentVolueClaim `models-pvc` and will be accesible by Ollama pods in `/data/models/qweb3-0.0b` path

## Easy install steps

Clone this repo and run:

`make lint && make manifests && make build && make test && make build-installer`

if you use a local kind cluster, you can disponibilize the image by running:

`docker build -t k8s-model-operator:latest . && kind load docker-image k8s-model-operator:latest --name k8s-models`

Then run `kubectl apply -f dist/install.yaml`

Please check that the image name is the one used by the controller

--- 
## TO DO

* We need to make the docker image available
* Create HPA and Service resources when synchronizing models
* Include correct images for llama.cpp and lamafiles

--- 
# KubeBuilder documentation
## Getting Started

### Prerequisites
- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### To Deploy on the cluster
**Build and push your image to the location specified by `IMG`:**

```sh
make docker-build docker-push IMG=<some-registry>/k8s-model-operator:tag
```

**NOTE:** This image ought to be published in the personal registry you specified.
And it is required to have access to pull the image from the working environment.
Make sure you have the proper permission to the registry if the above commands don’t work.

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with the image specified by `IMG`:**

```sh
make deploy IMG=<some-registry>/k8s-model-operator:tag
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**
You can apply the samples (examples) from the config/sample:

```sh
kubectl apply -k config/samples/
```

>**NOTE**: Ensure that the samples has default values to test it out.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/k8s-model-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/k8s-model-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Contributing
// TODO(user): Add detailed information on how you would like others to contribute to this project

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2026.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

