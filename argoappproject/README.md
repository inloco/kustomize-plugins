# Argo AppProject Kustomize Generator Plugin

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to generate an Argo's
AppProject with its access control definitions.

## Using

The plugin's manifest is pretty simple. It extends ArgoCD's AppProjects by adding the `accessControl`
attribute. In it, you can define which groups will have `read-only` and `read-sync` access to all applications within
the project.

```yaml
apiVersion: incognia.com/v1alpha1
kind: AppProject
metadata:
  name: employees
accessControl:
  readOnly:
    - sre:eng-1
  readSync:
    - sre:eng-0
```

Now we can specify `./unnamespaced.yaml` as a generator on `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./unnamespaced.yaml
```
