# This Dockerfile is used to build the image available on DockerHub
FROM golang:1.17.1 as build

# Add everything
ADD . /usr/src/chainsaw-cni

RUN  cd /usr/src/chainsaw-cni && \
     ./hack/build-go.sh

FROM dougbtv/chainsaw-baseimage:latest
LABEL org.opencontainers.image.source https://github.com/dougbtv/chainsaw-cni
COPY --from=build /usr/src/chainsaw-cni/bin /usr/src/chainsaw-cni/bin
COPY --from=build /usr/src/chainsaw-cni/LICENSE /usr/src/chainsaw-cni/LICENSE
WORKDIR /

ADD ./deployments/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
