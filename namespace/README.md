# Namespace Kustomize Generator Plugin
It is a plugin for [Kustomize](https://github.com/kubernetes-sigs/kustomize) that allows you to use Kubernetes Secrets encrypted with [SOPS](https://github.com/mozilla/sops) as a generator.

## Getting Started

### Install
To install this plugin on Kustomize, download the binary to Kustomize Plugin folder with `apiVersion: incognia.com/v1` and `kind: Namespace`. Then make it executable.

#### Linux 64-bits
```bash
PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1/sops

mkdir -p $PLACEMENT

PLUGIN=$PLACEMENT/SOPS

wget -O $PLUGIN https://github.com/inloco/sops-kustomize-generator-plugin/releases/download/v1.1.1/plugin-linux-amd64

chmod +x $PLUGIN
```

#### macOS 64-bits
```bash
PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1/sops

mkdir -p $PLACEMENT

PLUGIN=$PLACEMENT/SOPS

wget -O $PLUGIN https://github.com/inloco/sops-kustomize-generator-plugin/releases/download/v1.1.1/plugin-darwin-amd64

chmod +x $PLUGIN
```

#### Manual Build and Install for Other Systems and/or Architectures
```bash
git clone https://github.com/inloco/sops-kustomize-generator-plugin

cd sops-kustomize-generator-plugin

go get -d -v ./...

go build -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -tags netgo -v ./...

PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1/namespace

mkdir -p $PLACEMENT

mv ./sops-kustomize-generator-plugin $PLACEMENT/Namespace

cd ..

rm -fR sops-kustomize-generator-plugin
```

### Using

We can start with a regular Kubernetes Secret in its YAML format.
```yaml
apiVersion: v1
kind: Namespace
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
```

To convert it to a file that will be processed by the plugin, we replace `apiVersion: v1` with `apiVersion: incognia.com/v1` and `kind: Secret` with `kind: Namespace`.
```yaml
apiVersion: incognia.com/v1
kind: Namespace
metadata:
  name: mysecret
type: Opaque
data:
  username: YWRtaW4=
  password: MWYyZDFlMmU2N2Rm
```

Finally we encrypt it using Namespace with the following command:
```bash
sops --encrypt --encrypted-regex '^(data|stringData)$' --in-place ./secret.yaml
```

Now we can specify `./secret.yaml` as a generator on `kustomization.yaml`:
```yaml
apiVersion: kustomize.config.k8s.io/v1beta1
kind: Kustomization
generators:
  - ./secret.yaml
```

## Notes
- Remember to use `--enable-alpha-plugins` flag when running `kustomize build`.
- You may need to use environment variables, such as `AWS_PROFILE`, to configure Namespace decryption when running Kustomize.
- Integrity checks are disabled on Namespace decryption, this is done to prevent integrity failures due to Kustomize sortting the keys of original YAML file.
- This documentation assumes that you are familiar with [Kustomize](https://github.com/kubernetes-sigs/kustomize) and [Namespace](https://github.com/mozilla/sops), read their documentation if necessary.
- To make the generator behave like a patch, you might want to set `kustomize.config.k8s.io/behavior` annotation to `"merge"`. The other internal annotations described on [Kustomize Plugins Guide](https://kubernetes-sigs.github.io/kustomize/guides/plugins/#generator-options) are also supported.
