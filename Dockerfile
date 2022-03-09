FROM golang:1.16-alpine
WORKDIR /go/src/github.com/inloco/iac-kustomize-plugins
COPY . .
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64 GO111MODULE=on
RUN apk add git && \
    go get -d -v ./... && \
    go install -a -installsuffix cgo -ldflags '-extldflags "-static" -s -w' -tags netgo -v ./...

FROM alpine:3.14
COPY --from=0 /go/bin/sops-kustomize-generator-plugin /root/.config/kustomize/plugin/incognia.com/v1alpha1/clusterroles/ClusterRoles
COPY --from=0 /go/bin/sops-kustomize-generator-plugin /root/.config/kustomize/plugin/incognia.com/v1alpha1/namespace/Namespace
COPY --from=0 /go/bin/sops-kustomize-generator-plugin /root/.config/kustomize/plugin/incognia.com/v1alpha1/unnamespace/Unnamespace
