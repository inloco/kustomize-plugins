# KustomizeBuild Kustomize Generator Plugin

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to run multiple
kustomizations at once. It receives a list of directories as input, which uses the same syntax as `.gitignore`.

## Using

A KustomizeBuild can be defined as:

```yaml
# kustomizeBuild.yaml

apiVersion: incognia.com/v1alpha1
kind: KustomizeBuild
metadata:
  name: _
spec:
  directories:
    - base: git
      globs:
        - projects/**/argocd/**/production-product/
        - '!projects/**/argocd/**/production-product/**'
    - base: pwd
      globs:
        - ../../projects/**/argocd/**/staging-product/
        - '!../../projects/**/argocd/**/staging-product/**'
```

Now we can specify `./kustomizeBuild.yaml` as a generator on `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./kustomizeBuild.yaml
```
