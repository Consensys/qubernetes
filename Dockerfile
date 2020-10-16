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
RUN DEBIAN_FRONTEND="noninteractive" TZ="America/New_York" apt-get install -y ruby-full golang-go git make musl-dev
RUN gem install colorize

RUN go get github.com/getamis/istanbul-tools/cmd/istanbul
ENV PATH=/root/go/bin:$PATH

RUN go get github.com/getamis/istanbul-tools/cmd/istanbul && git clone https://github.com/ethereum/go-ethereum.git /root/go/src/github.com/ethereum/go-ethereum && \
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
