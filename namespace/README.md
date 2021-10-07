# Namespace Kustomize Generator Plugin

It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to use Kubernetes Secrets
encrypted with [SOPS](https://github.com/mozilla/sops) as a generator.

## Getting Started

### Setup

To install this plugin on Kustomize, download the binary to Kustomize Plugin folder
with `apiVersion: incognia.com/v1alpha1` and `kind: Namespace`. Then make it executable.

#### Linux 64-bits

```bash
PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1/namespace
mkdir -p $PLACEMENT
PLUGIN=$PLACEMENT/Namespace
wget -O $PLUGIN https://github.com/inloco/namespace-kustomize-generator-plugins/releases/download/v0.0.2/namespace-linux-amd64
chmod +x $PLUGIN
```

#### macOS 64-bits

```bash
PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1/namespace
mkdir -p $PLACEMENT
PLUGIN=$PLACEMENT/Namespace
wget -O $PLUGIN https://github.com/inloco/iac-kustomize-generator-plugins/releases/download/v0.0.2/namespace-darwin-amd64
chmod +x $PLUGIN
```

#### Manual Build and Install for Other Systems and/or Architectures

```bash
git clone https://github.com/inloco/iac-kustomize-generator-plugins
cd iac-kustomize-generator-plugins
go get -d -v ./...
go build -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -tags netgo -v ./...
PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1/namespace
mkdir -p $PLACEMENT
mv ./namespace-kustomize-generator-plugins $PLACEMENT/Namespace
cd ..
rm -fR iac-kustomize-generator-plugins
```

### Example

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
    - security@0
  readWrite:
    - sre@0
    - infrastructure@0
```

Now we can specify `./namespace.yaml` as a generator on `kustomization.yaml`:

```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./namespace.yaml
```

## Notes

- Remember to use `--enable-alpha-plugins` flag when running `kustomize build`.
- This documentation assumes that you are familiar with [Kustomize](https://github.com/kubernetes-sigs/kustomize), read
  their documentation if necessary.
- To make the generator behave like a patch, you might want to set `kustomize.config.k8s.io/behavior` annotation
  to `"merge"`. The other internal annotations described
  on [Kustomize Plugins Guide](https://kubernetes-sigs.github.io/kustomize/guides/plugins/#generator-options) are also
  supported.
