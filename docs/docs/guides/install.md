---
title: Installing Vidra
sidebar_position: 1
---
import Admonition from '@theme/Admonition';

# User Guide
## Prerequisites
Before installing the Vidra Operator, make sure you have access to a Kubernetes cluster (version 1.26.0 or higher) and that your `kubectl` tool is configured to communicate with it. For setup instructions, see the [Kubernetes installation guide](https://kubernetes.io/docs/setup/).

<Admonition type="note" title="Note">
For a development cluster, we recommend using [minikube](https://minikube.sigs.k8s.io/docs/start/) or [kind](https://kind.sigs.k8s.io/docs/user/quick-start/). Refer to the [Cluster Setup](../topics/deploy.md) guide for more details.
</Admonition>

## Install the Vidra Operator using Helm
Helm is a Kubernetes package manager that simplifies installing and managing applications like Vidra using charts.

<Admonition type="note" title="Note">
Helm is not required to run the Vidra Operator, but it simplifies the installation process. Helm version 3.0.0 or higher is recommended. See the [Helm installation guide](https://helm.sh/docs/intro/install/).
</Admonition>

### Helm Installation
To install the Vidra Operator using Helm OCI, use the following command:

```sh
helm install vidra-operator oci://ghcr.io/infrahub-operator/vidra/helm-charts/vidra-operator --namespace vidra-system --create-namespace
```
Or, if you prefer to use a Helm repository:

```sh
helm repo add infrahub-operator https://infrahub-operator.github.io/vidra
helm repo update
helm install vidra infrahub-operator/vidra-operator --namespace vidra-system --create-namespace
```
<Admonition type="note" title="Note">
The Helm repository is also released under https://github.com/infrahub-operator/vidra/releases/download/vX.Y.Z if you want to install an older version.
</Admonition>

Wait for the Vidra Operator to be installed (this might take a few seconds):

```sh
kubectl get pods -n vidra-system -w
```

---

## Install the Vidra Operator using OLM
OLM (Operator Lifecycle Manager) is a Kubernetes project that helps you manage the lifecycle of operators running on your cluster. It provides a way to install, update, and manage operators in a Kubernetes-native way.

<Admonition type="warning" title="Warning">
OLM is not required to run the Vidra Operator, but it simplifies the installation process. OLM version 0.32.0 or higher is recommended. See the [installation guide](https://operator-framework.github.io/operator-lifecycle-manager/docs/installation/).

**Alternative OLM installation:**

```sh
curl -sL https://github.com/operator-framework/operator-lifecycle-manager/releases/download/v0.32.0/install.sh | bash -s v0.32.0
```
</Admonition>

### Operator Lifecycle Manager (OLM) Installation

Install the Vidra Operator by creating a CatalogSource and Subscription:

```sh
kubectl apply -f https://raw.githubusercontent.com/Infrahub-Operator/Vidra/main/install/catalogsource.yaml -f https://raw.githubusercontent.com/Infrahub-Operator/Vidra/main/install/subscription.yaml
```
<Admonition type="note" title="Note">
It is also released under https://github.com/infrahub-operator/vidra/releases/download/vX.Y.Z if you want to install an older version.
</Admonition>

Wait for the Vidra Operator to be installed (this might take a few seconds):

```sh
kubectl get csv -n operators -w
```

---

## Uninstall Vidra Operator
<Admonition type="warning" title="Warning">
Before uninstalling the Vidra Operator, ensure that you have removed all `VidraResource` resources; otherwise, the finalizer could block the uninstall.
</Admonition>
To completely remove the Vidra Operator and its resources, follow these steps:

### Helm Uninstall
```sh
helm uninstall vidra-operator --namespace vidra-system
kubectl delete namespace vidra-system
```

### OLM Uninstall

#### 1. Delete the Subscription and ClusterServiceVersion (CSV)

```sh
kubectl delete subscriptions.operators.coreos.com -n operators vidra-operator-subscription 
kubectl delete csv -n operators $(kubectl get csv -n operators -o jsonpath="{.items[?(@.metadata.labels['operators.coreos.com/vidra.operators')].metadata.name}")
```

#### 2. Delete the CRDs

```sh
kubectl delete crd infrahubsyncs.infrahub.operators.com vidraresources.infrahub.operators.com
```

#### 3. Delete the Operator

```sh
kubectl delete operator vidra-operator.operators
```

#### 4. Delete the CatalogSource

```sh
kubectl delete catalogsource -n operators vidra-catalog
```
