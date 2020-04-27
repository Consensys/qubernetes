FROM ubuntu:16.04

ARG COMMIT=""
ARG QUBES_VERSION=""
ARG BUILD_DATE=""

# Label according to  https://github.com/opencontainers/image-spec
LABEL org.opencontainers.image.created=${BUILD_DATE}
LABEL org.opencontainers.image.revision=${COMMIT}
LABEL org.opencontainers.image.source="https://github.com/jpmorganchase/qubernetes.git"
LABEL org.opencontainers.image.title="qubernetes"
LABEL org.opencontainers.image.version=${QUBES_VERSION}

RUN apt-get update \
    && apt-get --no-install-recommends install -y apt-utils curl wget git tree ne software-properties-common apt-transport-https ca-certificates build-essential

RUN echo "deb [signed-by=/usr/share/keyrings/cloud.google.gpg] https://packages.cloud.google.com/apt cloud-sdk main" | tee -a /etc/apt/sources.list.d/google-cloud-sdk.list
RUN curl https://packages.cloud.google.com/apt/doc/apt-key.gpg | apt-key --keyring /usr/share/keyrings/cloud.google.gpg add -
RUN add-apt-repository ppa:longsleep/golang-backports
RUN add-apt-repository -y ppa:ethereum/ethereum
RUN curl -sL https://deb.nodesource.com/setup_10.x | bash -

# RUN apt-get update # done by node script above
RUN apt-get --no-install-recommends  install -y nodejs ruby haskell-stack golang-go google-cloud-sdk kubectl ethereum libdb-dev libleveldb-dev libsodium-dev zlib1g-dev libtinfo-dev
RUN gem install colorize
RUN npm install web3
ENV PATH=/root/go/bin:$PATH
RUN go get github.com/getamis/istanbul-tools/cmd/istanbul
RUN cd /usr/bin && curl -L https://github.com/jpmorganchase/constellation/releases/download/v0.3.2/constellation-0.3.2-ubuntu1604.tar.xz | tar -xJ --strip=1

WORKDIR /qubernetes
COPY . .

# set commit SHA and QUBES_VERSION as ENV vars in last layer
ENV COMMIT_SHA=${COMMIT}
ENV QUBES_VERSION=${QUBES_VERSION}
