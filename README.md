# IaC Kustomize Generator Plugins

## Setup

To install all plugins on Kustomize, download the binaries to Kustomize Plugin folder. Then make it executable.

#### Linux 64-bits

```bash
RELEASE_URL="https://github.com/inloco/iac-kustomize-generator-plugins/releases/download/v0.0.3"
PLACEMENT="${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1"
for KIND in Namespace Unnamespaced ClusterRoles
do
  KIND_LOWERCASE=$(echo ${KIND} | tr '[:upper:]' '[:lower:]')
  mkdir -p ${PLACEMENT}/${KIND_LOWERCASE}
  wget -O ${PLACEMENT}/${KIND_LOWERCASE}/${KIND} ${RELEASE_URL}/${KIND_LOWERCASE}-linux-amd64
  chmod +x ${PLACEMENT}/${KIND_LOWERCASE}/${KIND}
done
```

#### macOS 64-bits

```bash
RELEASE_URL="https://github.com/inloco/iac-kustomize-generator-plugins/releases/download/v0.0.3"
PLACEMENT="${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1"
for KIND in Namespace Unnamespaced ClusterRoles
do
  KIND_LOWERCASE=$(echo ${KIND} | tr '[:upper:]' '[:lower:]')
  mkdir -p ${PLACEMENT}/${KIND_LOWERCASE}
  wget -O ${PLACEMENT}/${KIND_LOWERCASE}/${KIND} ${RELEASE_URL}/${KIND_LOWERCASE}-darwin-amd64
  chmod +x ${PLACEMENT}/${KIND_LOWERCASE}/${KIND}
done
```

#### Manual Build and Install for Other Systems and/or Architectures

```bash
git clone https://github.com/inloco/iac-kustomize-generator-plugins
cd iac-kustomize-generator-plugins
make install
```
