# Unnamespaced Kustomize Generator Plugin

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to generate
ClusterRoleBindings to `unnamedspaced-ro` and `unnamedspaced-rw` ClusterRoles.

## Getting Started

### Example

The plugin's manifest is pretty simple, it only has the `accessControl` attribute. In it, you can define which groups
will have `read-only` and `read-write` access to all unnamespaced resources.

```yaml
apiVersion: incognia.com/v1alpha1
kind: Unnamespaced
accessControl:
  readOnly:
    - security:eng-0
  readWrite:
    - sre:eng-0
    - infrastructure:eng-0
```

Now we can specify `./unnamespaced.yaml` as a generator on `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./unnamespaced.yaml
```

## Notes

- Remember to use `--enable-alpha-plugins` flag when running `kustomize build`.
- This documentation assumes that you are familiar with [Kustomize](https://github.com/kubernetes-sigs/kustomize), read
  their documentation if necessary.
- To make the generator behave like a patch, you might want to set `kustomize.config.k8s.io/behavior` annotation
  to `"merge"`. The other internal annotations described
  on [Kustomize Plugins Guide](https://kubernetes-sigs.github.io/kustomize/guides/plugins/#generator-options) are also
  supported.
