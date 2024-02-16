FROM python:3.9.18
ARG USER=user
ENV GOLANG_VERSION=1.20.14

RUN useradd -s /bin/bash -m ${USER}
RUN apt update && \
    apt install -y wget sudo gzip vim
RUN echo "${USER} ALL=NOPASSWD: ALL" >> /etc/sudoers

WORKDIR /workspace
# Go Install
ENV GOROOT=/usr/local/go
ENV GOPATH=/workspace
RUN wget -O /workspace/go${GOLANG_VERSION}.linux-amd64.tar.gz https://go.dev/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    tar -xzvf /workspace/go${GOLANG_VERSION}.linux-amd64.tar.gz && \
    mv /workspace/go /usr/local/ && \
    chmod -R 777 ${GOPATH} && chmod -R 777 ${GOROOT}

USER ${USER}
ENV PATH=${GOPATH}/bin:/usr/local/go/bin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/home/${USER}/.local/bin:/usr/local/bin

