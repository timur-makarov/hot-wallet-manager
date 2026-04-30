# syntax=docker/dockerfile:1.7
# https://github.com/trustwallet/wallet-core/blob/master/Dockerfile

FROM --platform=linux/amd64 ubuntu:22.04 AS wallet-core-build

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates wget curl git unzip xz-utils \
    build-essential libtool autoconf pkg-config \
    ninja-build ruby-full \
    clang-14 llvm-14 libc++-14-dev libc++abi-14-dev \
    cmake libboost-all-dev ccache \
    && rm -rf /var/lib/apt/lists/*

ENV CC=/usr/bin/clang-14 \
    CXX=/usr/bin/clang++-14
RUN ln -sf /usr/bin/clang-14 /usr/bin/clang \
    && ln -sf /usr/bin/clang++-14 /usr/bin/clang++

ARG RUST_TOOLCHAIN=nightly-2025-12-11
RUN curl -fsSL https://sh.rustup.rs | sh -s -- -y --default-toolchain none --no-modify-path
ENV PATH="/root/.cargo/bin:${PATH}"
RUN rustup toolchain install ${RUST_TOOLCHAIN} \
    && rustup default ${RUST_TOOLCHAIN} \
    && cargo install --force cbindgen --locked

ARG WALLET_CORE_REF=4.6.6
RUN git clone --depth=1 --branch=${WALLET_CORE_REF} \
    https://github.com/trustwallet/wallet-core.git /wallet-core
WORKDIR /wallet-core

RUN tools/install-dependencies
RUN tools/generate-files native
RUN cmake -H. -Bbuild -DCMAKE_BUILD_TYPE=Release \
    && make -Cbuild -j"$(nproc)" TrustWalletCore

FROM --platform=linux/amd64 ubuntu:22.04 AS go-build

ARG GO_VERSION=1.26.0
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates wget tar git \
    build-essential pkg-config \
    clang-14 protobuf-compiler \
    && rm -rf /var/lib/apt/lists/*

ADD https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz /tmp/go.tgz
RUN rm -rf /usr/local/go \
    && tar -C /usr/local -xzf /tmp/go.tgz \
    && rm -f /tmp/go.tgz

COPY --from=wallet-core-build /wallet-core/include              /wallet-core/include
COPY --from=wallet-core-build /wallet-core/build                /wallet-core/build
COPY --from=wallet-core-build /wallet-core/rust/target/release  /wallet-core/rust/target/release

ENV PATH="/usr/local/go/bin:${PATH}" \
    CC=/usr/bin/clang-14 \
    CXX=/usr/bin/clang++-14 \
    CGO_ENABLED=1 \
    GO111MODULE=on \
    CGO_CFLAGS="-I/wallet-core/include" \
    CGO_LDFLAGS="-L/wallet-core/build -L/wallet-core/build/local/lib -L/wallet-core/build/trezor-crypto -L/wallet-core/rust/target/release -lTrustWalletCore -lwallet_core_rs -lprotobuf -lTrezorCrypto -lstdc++ -lm -lpthread -ldl"

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN make api_gen

RUN make proto_gen

RUN go test ./...

RUN go build -o sheepy-tt-go-wallet ./cmd

FROM --platform=linux/amd64 ubuntu:22.04

WORKDIR /app

ENV DEBIAN_FRONTEND=noninteractive
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    && rm -rf /var/lib/apt/lists/*

COPY --from=go-build /app/sheepy-tt-go-wallet /app/sheepy-tt-go-wallet

ENV WALLET_CONFIG=/app/config.json

EXPOSE 8000

CMD ["./sheepy-tt-go-wallet"]
