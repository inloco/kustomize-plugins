# ClusterRoles Kustomize Generator Plugin

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that dynamically generates read-only and
read-write ClusterRules for namespaced and unnamespaced resources using the K8s Discovery API.

## Using

Create `./clusterroles.yaml`.

```yaml
apiVersion: incognia.com/v1alpha1
kind: ClusterRoles
```

Specify `./clusterroles.yaml` as a generator on `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./clusterroles.yaml
```

Build the KRM resourced using Kustomize while connected to the target cluster.

```yaml
kustomize build --enable-alpha-plugins
```

The generated output will contain four ClusterRoles. `namespaced-ro` and `namespaced-rw` must be used with RoleBindings.
`unnamespaced-ro` and `unnamespaced-rw` must be used with ClusterRoleBindings.

```yaml
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: namespaced-ro
rules:
  ...
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: namespaced-rw
rules:
  ...
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: unnamespaced-ro
rules:
  ...
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: unnamespaced-rw
rules:
  ...
```
