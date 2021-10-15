#!/bin/sh

OS_NAME_LOWERCASE=$(uname -s | tr '[:upper:]' '[:lower:]')
PLACEMENT=${XDG_CONFIG_HOME:-$HOME/.config}/kustomize/plugin/incognia.com/v1alpha1
RELEASE_URL=https://github.com/inloco/iac-kustomize-generator-plugins/releases/download/v0.0.6

for KIND in Namespace Unnamespaced ClusterRoles
do
  KIND_LOWERCASE=$(echo ${KIND} | tr '[:upper:]' '[:lower:]')
  mkdir -p ${PLACEMENT}/${KIND_LOWERCASE}
  wget -O ${PLACEMENT}/${KIND_LOWERCASE}/${KIND} ${RELEASE_URL}/${KIND_LOWERCASE}-${OS_NAME_LOWERCASE}-amd64
  chmod +x ${PLACEMENT}/${KIND_LOWERCASE}/${KIND}
done
