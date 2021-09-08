FROM ubuntu:20.10

ARG COMMIT=""
ARG QUBES_VERSION=""
ARG BUILD_DATE=""

# Label according to  https://github.com/opencontainers/image-spec
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL org.opencontainers.image.revision=${COMMIT}
LABEL org.opencontainers.image.source="https://github.com/ConsenSys/qubernetes.git"
LABEL org.opencontainers.image.title="qubernetes"
LABEL org.opencontainers.image.version=${QUBES_VERSION}

RUN apt-get update

# set tzdata non-interactive https://serverfault.com/questions/949991/how-to-install-tzdata-on-a-ubuntu-docker-image
# for now need musl-dev for geneating account key from the private key
RUN DEBIAN_FRONTEND="noninteractive" TZ="America/New_York" apt-get install -y ruby-full golang-go git make musl-dev xxd wget
RUN gem install colorize

RUN mkdir -p /root/go/bin

ENV PATH=/root/go/bin:$PATH

RUN cd /root/go/bin && \
    wget https://artifacts.consensys.net/public/quorum-tools/raw/versions/v1.1.0/istanbul-tools_v1.1.0_linux_amd64.tar.gz &&  \
    tar -xvf istanbul-tools_v1.1.0_linux_amd64.tar.gz &&  \
    rm istanbul-tools_v1.1.0_linux_amd64.tar.gz

RUN git clone https://github.com/ethereum/go-ethereum.git /root/go/src/github.com/ethereum/go-ethereum && \
    cd /root/go/src/github.com/ethereum/go-ethereum && git checkout e9ba536d && make all && \
    cp /root/go/src/github.com/ethereum/go-ethereum/build/bin/ethkey /root/go/bin/ && \
    cp /root/go/src/github.com/ethereum/go-ethereum/build/bin/bootnode /root/go/bin/ && \
    cp /root/go/bin/* /usr/local/bin && \
    rm -r /root/go

RUN apt-get remove -y git golang-go wget make
# uninstall rake
RUN gem uninstall --no-executables -i /usr/share/rubygems-integration/all rake && rm /usr/bin/rake

WORKDIR /qubernetes
COPY . .

# set commit SHA and QUBES_VERSION as ENV vars in last layer
ENV COMMIT_SHA=${COMMIT}
ENV QUBES_VERSION=${QUBES_VERSION}
