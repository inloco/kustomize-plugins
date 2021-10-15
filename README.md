# IaC Kustomize Generator Plugins

## Setup

To install all plugins, download the binaries to the Kustomize plugin folder and make them executable.

### Linux 64-bits and macOS 64-bits

```bash
wget -qO- https://github.com/inloco/iac-kustomize-generator-plugins/releases/download/v0.0.6/install.sh | sh
```

### Manual Build and Install for Other Systems and/or Architectures

```bash
git clone https://github.com/inloco/iac-kustomize-generator-plugins
cd iac-kustomize-generator-plugins
make install
```

## Notes

- Remember to use `--enable-alpha-plugins` flag when running `kustomize build`.
- This documentation assumes that you are familiar with [Kustomize](https://github.com/kubernetes-sigs/kustomize), read their documentation if necessary.
- To make the generator behave like a patch, you might want to set `kustomize.config.k8s.io/behavior` annotation to `"merge"`. The other internal annotations described on [Kustomize Plugins Guide](https://kubernetes-sigs.github.io/kustomize/guides/plugins/#generator-options) are also supported.
