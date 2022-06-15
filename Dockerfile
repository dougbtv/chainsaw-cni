# This Dockerfile is used to build the image available on DockerHub
FROM golang:1.17.1 as build

# Add everything
ADD . /usr/src/swiss-army-knife-cni

RUN  cd /usr/src/swiss-army-knife-cni && \
     ./hack/build-go.sh

FROM centos:centos7
LABEL org.opencontainers.image.source https://github.com/dougbtv/swiss-army-knife-cni
COPY --from=build /usr/src/swiss-army-knife-cni/bin /usr/src/sak-cni/bin
COPY --from=build /usr/src/swiss-army-knife-cni/LICENSE /usr/src/sak-cni/LICENSE
WORKDIR /

ADD ./deployments/entrypoint.sh /
ENTRYPOINT ["/entrypoint.sh"]
