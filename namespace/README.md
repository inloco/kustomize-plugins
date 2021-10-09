# Namespace Kustomize Generator Plugin

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to generate a Namespace
with its access control definitions.

## Using

We can start with a regular Kubernetes Namespace in its YAML format.

```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: my-namespace
```

To convert it to a file that will be processed by the plugin, we replace `apiVersion: v1`
with `apiVersion: incognia.com/v1alpha1`.

By doing this, you'll have access to the `accessControl` attribute. In it, you can define which groups will
have `read-only` and `read-write` access to the namespace.

```yaml
apiVersion: incognia.com/v1alpha1
kind: Namespace
metadata:
  name: my-namespace
accessControl:
  readOnly:
    - security:eng-0
  readWrite:
    - sre:eng-0
    - infrastructure:eng-0
```

Now we can specify `./namespace.yaml` as a generator on `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./namespace.yaml
```
