# Kustomize Generator Plugins

## Setup

To install all plugins, download the binaries to the Kustomize plugin folder and make them executable.

### Linux 64-bits and/or macOS 64-bits

```bash
VERSION=$(wget -qO- https://api.github.com/repos/inloco/kustomize-generator-plugins/releases/latest | jq -r '.tag_name')
wget -qO- "https://github.com/inloco/kustomize-generator-plugins/releases/download/${VERSION}/install.sh" | sh
```

### Manual Build and Install for Other Systems and/or Architectures

```bash
git clone https://github.com/inloco/kustomize-generator-plugins
cd kustomize-generator-plugins
make install
```

## Notes

- Remember to use `--enable-alpha-plugins` flag when running `kustomize build`.
- This documentation assumes that you are familiar with [Kustomize](https://github.com/kubernetes-sigs/kustomize), read their documentation if necessary.
- To make the generator behave like a patch, you might want to set `kustomize.config.k8s.io/behavior` annotation to `"merge"`. The other internal annotations described on [Kustomize Plugins Guide](https://kubernetes-sigs.github.io/kustomize/guides/plugins/#generator-options) are also supported.
