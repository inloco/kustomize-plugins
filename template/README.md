# Template Transformer

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to template manifests by
using Go's [text/template](https://pkg.go.dev/text/template) library.

This plugin uses `([{` and `}])` as delimiters to avoid collision with other libraries.

## Using

A Template transformer can be defined as:

```yaml
apiVersion: incognia.com/v1alpha1
kind: Template
metadata:
  name: _
data:
  aws:
    account:
      name: example-aws-account-1234
      number: 1234
  cluster:
    name: blue
    owner: mario
```

Specify `./template.yaml` as a transform on `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
transformers:
  - ./template.yaml
resources:
  - deployment.yaml
```

If your `./deployment.yaml` file look like:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    spec:
      containers:
        - name: example
          command:
            - example-cmd
            - input-arg=([{ .aws.account.name }])
```

It'll be outputted as:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: example
spec:
  template:
    spec:
      containers:
        - name: example
          command:
            - example-cmd
            - input-arg=example-aws-account-1234
```
