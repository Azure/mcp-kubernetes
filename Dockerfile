# Build stage
FROM golang:1.24-alpine AS builder
ARG VERSION
ARG GIT_COMMIT
ARG BUILD_DATE
ARG GIT_TREE_STATE
# Set working directory
WORKDIR /app

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags "-X github.com/Azure/mcp-kubernetes/pkg/version.GitVersion=${VERSION} -X github.com/Azure/mcp-kubernetes/pkg/version.GitCommit=${GIT_COMMIT} -X github.com/Azure/mcp-kubernetes/pkg/version.BuildMetadata=${BUILD_DATE} -X github.com/Azure/mcp-kubernetes/pkg/version.GitTreeState=${GIT_TREE_STATE}" -o mcp-kubernetes ./cmd/mcp-kubernetes

# Runtime stage
FROM alpine:3.19

# Install required packages for kubectl and helm
RUN apk add --no-cache curl bash openssl ca-certificates git

# Create the mcp user and group
RUN addgroup -S mcp && \
    adduser -S -G mcp -h /home/mcp mcp && \
    mkdir -p /home/mcp/.kube && \
    chown -R mcp:mcp /home/mcp

# Install kubectl
RUN curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && \
    chmod +x kubectl && \
    mv kubectl /usr/local/bin/kubectl

# Install helm
RUN curl -fsSL -o get_helm.sh https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 && \
    chmod 700 get_helm.sh && \
    VERIFY_CHECKSUM=false ./get_helm.sh && \
    rm get_helm.sh

# Install cilium
RUN CILIUM_CLI_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/cilium-cli/main/stable.txt) && \
    CLI_ARCH=amd64 && \
    if [ "$(uname -m)" = "aarch64" ]; then CLI_ARCH=arm64; fi && \
    curl -L --fail --remote-name-all https://github.com/cilium/cilium-cli/releases/download/${CILIUM_CLI_VERSION}/cilium-linux-${CLI_ARCH}.tar.gz && \
    tar xzf cilium-linux-${CLI_ARCH}.tar.gz -C /usr/local/bin && \
    rm cilium-linux-${CLI_ARCH}.tar.gz

RUN HUBBLE_VERSION=$(curl -s https://raw.githubusercontent.com/cilium/hubble/master/stable.txt) && \
    HUBBLE_ARCH=amd64 && \
    if [ "$(uname -m)" = "aarch64" ]; then HUBBLE_ARCH=arm64; fi && \
    curl -L --fail --remote-name-all https://github.com/cilium/hubble/releases/download/$HUBBLE_VERSION/hubble-linux-${HUBBLE_ARCH}.tar.gz{,.sha256sum} && \
    sha256sum -c hubble-linux-${HUBBLE_ARCH}.tar.gz.sha256sum && \
    tar xzvf hubble-linux-${HUBBLE_ARCH}.tar.gz -C /usr/local/bin && \
    rm hubble-linux-${HUBBLE_ARCH}.tar.gz hubble-linux-${HUBBLE_ARCH}.tar.gz.sha256sum

# Copy binary from builder
COPY --from=builder /app/mcp-kubernetes /usr/local/bin/mcp-kubernetes

# Set working directory
WORKDIR /home/mcp

# Switch to non-root user
USER mcp

# Set environment variables
ENV HOME=/home/mcp \
    KUBECONFIG=/home/mcp/.kube/config

# Command to run
ENTRYPOINT ["/usr/local/bin/mcp-kubernetes"]
CMD ["--transport", "stdio"]
