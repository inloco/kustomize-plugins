#!/bin/sh

set -e

VERSION=v0.8.3
RELEASE_URL=https://github.com/inloco/kustomize-plugins/releases/download/${VERSION}

TEMP_DIC=$(mktemp -d)
for KIND in ArgoCDProject ClusterRoles KustomizeBuild Namespace Template Unnamespaced
do
	KIND_LOWERCASE=$(echo ${KIND} | tr '[:upper:]' '[:lower:]')
	wget -P $TEMP_DIC ${RELEASE_URL}/${KIND_LOWERCASE}-darwin-amd64
    wget -P $TEMP_DIC ${RELEASE_URL}/${KIND_LOWERCASE}-linux-amd64
done

aws s3 cp $TEMP_DIC s3://kustomize-plugins.incognia.tech/inloco/kustomize-plugins/${VERSION} --recursive

rm -r $TEMP_DIC