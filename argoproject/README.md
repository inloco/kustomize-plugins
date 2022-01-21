# Argo AppProject Kustomize Generator Plugin

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to generate an Argo's
AppProject and Applications with its access control definitions.

## Using

The plugin's manifest is pretty simple. It defines the following attributes:

- `spec.accessControl`: allows role access management. In it, you can define which groups will have `read-only`
  and `read-sync`
  access to all applications within the project.

- `spec.destination`: allows your applications' Cluster and Namespace destination.

- `spec.appProjectTemplate`: allows any additional fields for the argoproj.io AppProject.

- `spec.applicationTemplates`: allows multiple argoproj.io Application to be defined, since one project can contain
  multiple applications.

An ArgoProject can be defined as:

```yaml
# employees.argoProject.yaml

apiVersion: incognia.com/v1alpha1
kind: ArgoProject
metadata:
  name: employees
spec:
  accessControl:
    readOnly:
      - sre:eng-1
    readSync:
      - sre:eng-0
  destination:
    name: GlobalStaging-Product
    namespace: employees
  appProjectTemplate:
    spec:
      clusterResourceBlacklist:
        - group: ''
          kind: Secret
  applicationTemplates:
    - metadata:
        name: employees
      spec:
        source:
          repoURL: https://github.com/inloco/employees.git
          targetRevision: argocd-stag
          path: ./k8s/overlays/global-staging/
```

Now we can specify `./employees.argoProject.yaml` as a generator in `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./employees.argoProject.yaml
```
