# mcp-kubernetes

The mcp-kubernetes is a Model Context Protocol (MCP) server that enables AI assistants to interact with Kubernetes clusters. It serves as a bridge between AI tools (like Claude, Cursor, and GitHub Copilot) and Kubernetes, translating natural language requests into Kubernetes operations and returning the results in a format the AI tools can understand.

It allows AI tools to:

- Query Kubernetes resources
- Execute kubectl commands
- Manage Kubernetes clusters through natural language interactions
- Diagnose and interpret the states of Kubernetes resources

## How it works

![](assets/mcp-kubernetes-server.png)

## How to install

### Docker

Get your kubeconfig file for your Kubernetes cluster and setup in the mcpServers (replace src path with your kubeconfig path):

```json
{
  "mcpServers": {
    "kubernetes": {
      "command": "docker",
      "args": [
        "run",
        "-i",
        "--rm",
        "--mount",
        "type=bind,src=/home/username/.kube/config,dst=/home/mcp/.kube/config",
        "ghcr.io/azure/mcp-kubernetes"
      ]
    }
  }
}
```

### Local

<details>

<summary>Install kubectl</summary>

Install [kubectl](https://kubernetes.io/docs/tasks/tools/) if it's not installed yet and add it to your PATH, e.g.

```bash
# For Linux
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl"

# For MacOS
curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/darwin/arm64/kubectl"
```

</details>

<details>
<summary>Install helm</summary>

Install [helm](https://helm.sh/docs/intro/install/) if it's not installed yet and add it to your PATH, e.g.

```bash
curl -sSL https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3 | bash
```

</details>

<br/>

Config your MCP servers in [Claude Desktop](https://claude.ai/download), [Cursor](https://www.cursor.com/), [ChatGPT Copilot](https://marketplace.visualstudio.com/items?itemName=feiskyer.chatgpt-copilot), [Github Copilot](https://github.com/features/copilot) and other supported AI clients, e.g.

```json
{
  "mcpServers": {
    "kubernetes": {
      "command": "<path of binary 'mcp-kubernetes'>",
      "args": ["--transport", "stdio"],
      "env": {
        "KUBECONFIG": "<your-kubeconfig-path>"
      }
    }
  }
}
```

### Options

Environment variables:

- `KUBECONFIG`: Path to your kubeconfig file, e.g. `/home/<username>/.kube/config`.

Command line arguments:

```sh
Usage of ./mcp-kubernetes:
      --access-level string       Access level (readonly, readwrite, or admin) (default "readonly")
      --additional-tools string   Comma-separated list of additional tools to support (kubectl is always enabled). Available: helm,cilium
      --allow-namespaces string   Comma-separated list of namespaces to allow (empty means all allowed)
      --host string               Host to listen for the server (only used with transport sse or streamable-http) (default "127.0.0.1")
      --port int                  Port to listen for the server (only used with transport sse or streamable-http) (default 8000)
      --timeout int               Timeout for command execution in seconds, default is 60s (default 60)
      --transport string          Transport mechanism to use (stdio, sse or streamable-http) (default "stdio")
```

### Access Levels

The `--access-level` flag controls what operations are allowed and which tools are available:

- **`readonly`** (default): Only read operations are allowed (get, describe, logs, etc.)
  - Available tools: 4 kubectl tools for viewing resources
- **`readwrite`**: Read and write operations are allowed (create, delete, apply, etc.)
  - Available tools: 6 kubectl tools for managing resources
- **`admin`**: All operations are allowed, including admin operations (cordon, drain, taint, etc.)
  - Available tools: All 7 kubectl tools including node management

Tools are filtered at registration time based on the access level, so AI assistants only see tools they can actually use.

Example configurations:

```json
// Read-only access (default)
{
  "mcpServers": {
    "kubernetes": {
      "command": "mcp-kubernetes"
    }
  }
}

// Read-write access
{
  "mcpServers": {
    "kubernetes": {
      "command": "mcp-kubernetes",
      "args": ["--access-level", "readwrite"]
    }
  }
}

// Admin access
{
  "mcpServers": {
    "kubernetes": {
      "command": "mcp-kubernetes",
      "args": ["--access-level", "admin"]
    }
  }
}
```

## Usage

Ask any questions about Kubernetes cluster in your AI client. The MCP tools make it easier for AI assistants to understand and use kubectl operations.

### Example Queries

```txt
What is the status of my Kubernetes cluster?

What is wrong with my nginx pod?

Show me all deployments in the production namespace

Scale my web deployment to 5 replicas

Check if I have permission to create pods
```

## Available Tools

The mcp-kubernetes server provides consolidated kubectl tools that group related operations together. Tools are automatically filtered based on your access level.

### Kubectl Tools

<details>
<summary><b>kubectl_resources</b> - Manage Kubernetes resources</summary>

**Available in**: readonly, readwrite, admin

Handles CRUD operations on Kubernetes resources. In readonly mode, only supports `get` and `describe` operations.

**Parameters:**

- `operation`: The operation to perform (get, describe, create, delete, apply, patch, replace)
- `resource`: The resource type (e.g., pods, deployments, services) or empty for file-based operations
- `args`: Additional arguments like resource names, namespaces, and flags

**Examples:**

```bash
# Get all pods
operation: "get"
resource: "pods"
args: "--all-namespaces"

# Apply a configuration
operation: "apply"
resource: ""
args: "-f deployment.yaml"
```

</details>

<details>
<summary><b>kubectl_workloads</b> - Manage workload deployments</summary>

**Available in**: readwrite, admin

Manages deployment lifecycle operations including scaling and rollouts.

**Parameters:**

- `operation`: The operation to perform (run, expose, scale, autoscale, rollout)
- `resource`: For rollout operations, the subcommand (status, history, undo, restart, pause, resume)
- `args`: Additional arguments

**Examples:**

```bash
# Scale a deployment
operation: "scale"
resource: "deployment"
args: "nginx --replicas=3"

# Check rollout status
operation: "rollout"
resource: "status"
args: "deployment/nginx"
```

</details>

<details>
<summary><b>kubectl_metadata</b> - Manage resource metadata</summary>

**Available in**: readwrite, admin

Updates labels, annotations, and other metadata on resources.

**Parameters:**

- `operation`: The operation to perform (label, annotate, set)
- `resource`: The resource type
- `args`: Resource name and metadata changes

**Examples:**

```bash
# Add a label
operation: "label"
resource: "pods"
args: "nginx-pod app=web"

# Set image
operation: "set"
resource: "image"
args: "deployment/nginx nginx=nginx:latest"
```

</details>

<details>
<summary><b>kubectl_diagnostics</b> - Debug and monitor resources</summary>

**Available in**: readonly, readwrite, admin

Provides debugging and monitoring capabilities.

**Parameters:**

- `operation`: The operation to perform (logs, events, top, exec, cp)
- `resource`: The resource type or specific resource
- `args`: Additional arguments

**Examples:**

```bash
# View logs
operation: "logs"
resource: ""
args: "nginx-pod -f"

# Execute command in pod
operation: "exec"
resource: ""
args: "nginx-pod -- ls /app"
```

</details>

<details>
<summary><b>kubectl_cluster</b> - View cluster information</summary>

**Available in**: readonly, readwrite, admin

Provides cluster-level information and API discovery.

**Parameters:**

- `operation`: The operation to perform (cluster-info, api-resources, api-versions, explain)
- `resource`: For explain operation, the resource to document
- `args`: Additional flags

**Examples:**

```bash
# Get cluster info
operation: "cluster-info"
resource: ""
args: ""

# Explain pod spec
operation: "explain"
resource: "pod.spec"
args: "--recursive"
```

</details>

<details>
<summary><b>kubectl_nodes</b> - Manage cluster nodes</summary>

**Available in**: admin only

Manages node-level operations for cluster maintenance.

**Parameters:**

- `operation`: The operation to perform (cordon, uncordon, drain, taint)
- `resource`: Usually 'node' or 'nodes'
- `args`: Node names and operation-specific flags

**Examples:**

```bash
# Drain a node
operation: "drain"
resource: "node"
args: "worker-1 --ignore-daemonsets"

# Add a taint
operation: "taint"
resource: "nodes"
args: "worker-1 key=value:NoSchedule"
```

</details>

<details>
<summary><b>kubectl_config</b> - Configuration and security</summary>

**Available in**: readonly, readwrite, admin

Handles configuration validation and security operations. In readonly mode, only supports `diff` and `auth can-i`.

**Parameters:**

- `operation`: The operation to perform (diff, auth, certificate)
- `resource`: Subcommand for auth/certificate operations
- `args`: Operation-specific arguments

**Examples:**

```bash
# Check permissions
operation: "auth"
resource: "can-i"
args: "create pods"

# Approve certificate
operation: "certificate"
resource: "approve"
args: "csr-name"
```

</details>

### Additional Tools

<details>
<summary><b>helm</b> - Helm package manager</summary>

**Available when**: `--additional-tools=helm` is specified

Run Helm commands for managing Kubernetes applications.

**Parameters:**

- `command`: The helm command to execute

**Example:**

```bash
command: "list --all-namespaces"
```

</details>

<details>
<summary><b>cilium</b> - Cilium CNI commands</summary>

**Available when**: `--additional-tools=cilium` is specified

Run Cilium commands for network policies and observability.

**Parameters:**

- `command`: The cilium command to execute

**Example:**

```bash
command: "status --brief"
```

</details>

## Development

How to inspect MCP server requests and responses:

```sh
npx @modelcontextprotocol/inspector <path of binary 'mcp-kubernetes'>
```

## Contributing

This project welcomes contributions and suggestions. Most contributions require you to agree to a Contributor License Agreement (CLA) declaring that you have the right to, and actually do, grant us the rights to use your contribution. For details, visit https://cla.opensource.microsoft.com.

When you submit a pull request, a CLA bot will automatically determine whether you need to provide a CLA and decorate the PR appropriately (e.g., status check, comment). Simply follow the instructions provided by the bot. You will only need to do this once across all repos using our CLA.

This project has adopted the [Microsoft Open Source Code of Conduct](https://opensource.microsoft.com/codeofconduct/). For more information see the [Code of Conduct FAQ](https://opensource.microsoft.com/codeofconduct/faq/) or contact [opencode@microsoft.com](mailto:opencode@microsoft.com) with any additional questions or comments.

## Trademarks

This project may contain trademarks or logos for projects, products, or services. Authorized use of Microsoft trademarks or logos is subject to and must follow [Microsoft's Trademark & Brand Guidelines](https://www.microsoft.com/en-us/legal/intellectualproperty/trademarks/usage/general). Use of Microsoft trademarks or logos in modified versions of this project must not cause confusion or imply Microsoft sponsorship. Any use of third-party trademarks or logos are subject to those third-party's policies.
