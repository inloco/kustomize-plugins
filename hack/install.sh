#!/bin/sh
set -e

OS_ARCH=$(uname -m | sed 's/x86_64/amd64/g')
OS_NAME=$(uname -s | tr '[:upper:]' '[:lower:]')
PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1
RELEASE_URL=https://github.com/inloco/iac-kustomize-generator-plugins/releases/download/v0.0.0

for KIND in ArgoCDProject ClusterRoles KustomizeBuild Namespace Unnamespaced
do
	KIND_LOWERCASE=$(echo ${KIND} | tr '[:upper:]' '[:lower:]')
	mkdir -p ${PLACEMENT}/${KIND_LOWERCASE}
	wget -O ${PLACEMENT}/${KIND_LOWERCASE}/${KIND} ${RELEASE_URL}/${KIND_LOWERCASE}-${OS_NAME}-amd64
	chmod +x ${PLACEMENT}/${KIND_LOWERCASE}/${KIND}
done
