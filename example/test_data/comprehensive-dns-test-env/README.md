# DNS Test Environment

This directory contains YAML files to create a Kubernetes test environment with intentional DNS configuration issues for troubleshooting practice.

## Files Description

- **namespace.yaml**: Creates three namespaces (dns-test, app-backend, secure-ns)
- **web-app.yaml**: Web application in dns-test namespace
- **backend-app.yaml**: Backend application in app-backend namespace  
- **secure-app.yaml**: Application with incorrect DNS policy in secure-ns namespace
- **restrictive-network-policy.yaml**: NetworkPolicy that blocks DNS traffic
- **coredns-custom-config.yaml**: Custom CoreDNS config with problematic settings
- **failing-workloads.yaml**: Workload applications that crash due to DNS resolution failures
- **clean.sh**: Script to completely remove the test environment and restore cluster to original state

## DNS Issues Introduced

1. **Incorrect DNS Policy**: Pods in `secure-ns` have wrong nameserver configuration
   - ⚠️ **Critical**: Using external DNS `8.8.8.8` instead of cluster DNS `10.0.0.10`
   - ⚠️ **Subtle**: Search domains set to `production.local` and `company.internal` instead of Kubernetes cluster domains
   - ⚠️ **Configuration**: `ndots:2` instead of standard `ndots:5` for Kubernetes

2. **NetworkPolicy Blocking DNS**: `app-backend` namespace blocks port 53 traffic
3. **CoreDNS Misconfiguration**: Custom config with invalid upstream DNS servers

## Deployment

```bash
chmod +x deploy.sh
./deploy.sh
```

## Observing DNS Issues

The workload pods will crash or fail health checks due to DNS problems:

```bash
# Watch for failing pods
kubectl get pods --all-namespaces -w

# Check pod status and restart counts
kubectl get pods -n app-backend
kubectl get pods -n secure-ns
kubectl get pods -n dns-test

# View logs to see DNS resolution errors
kubectl logs deployment/failing-worker -n app-backend
kubectl logs deployment/dns-dependent-app -n secure-ns
kubectl logs deployment/health-check-app -n dns-test
```

## Expected Failures

1. **failing-worker** (app-backend): Will crash due to NetworkPolicy blocking DNS
2. **dns-dependent-app** (secure-ns): Will crash due to wrong DNS configuration
3. **health-check-app** (dns-test): Health checks may fail depending on CoreDNS issues

## Troubleshooting Hints

🔍 **Key Investigation Points:**
- Check `/etc/resolv.conf` in failing pods - look for incorrect nameserver IPs
- Notice search domains that don't match Kubernetes patterns (`*.svc.cluster.local`)
- Verify `ndots` setting - Kubernetes typically uses `ndots:5`
- The DNS issues may look like valid corporate network configuration at first glance

## Cleanup

To completely remove the test environment and restore your cluster:

```bash
chmod +x clean.sh
./clean.sh
```

This will:
- Delete all test namespaces (dns-test, app-backend, secure-ns)
- Restore CoreDNS custom configuration to default (empty)
- Restart CoreDNS pods to apply clean configuration
- Remove any leftover test resources

## Manual Cleanup (Alternative)

```bash
kubectl delete namespace dns-test app-backend secure-ns
kubectl patch configmap coredns-custom -n kube-system --type merge -p '{"data":{}}'
kubectl rollout restart deployment/coredns -n kube-system
```

## Testing with MCP and AI Assistants

This environment is designed to test the mcp-kubernetes server's ability to help AI assistants diagnose complex DNS issues in Kubernetes clusters.

### Prerequisites for MCP Testing

1. **mcp-kubernetes server** built and available
2. **VS Code with GitHub Copilot** or another MCP-compatible AI assistant
3. **Kubernetes cluster** with kubectl configured

### MCP Server Setup

1. **Configure MCP Server**: Create or update `.vscode/mcp.json`:

```json
{
    "servers": {
        "mcp-kubernetes": {
            "type": "stdio",
            "command": "/path/to/your/mcp-kubernetes",
            "args": []
        }
    }
}
```


### MCP Testing Workflow

#### Step 1: Deploy the Test Environment

```bash
chmod +x deploy.sh
./deploy.sh
```

#### Step 2: Verify Issues are Present

```bash
# Check for failing pods
kubectl get pods --all-namespaces

# Observe the specific failures
kubectl get pods -n app-backend -o wide
kubectl get pods -n secure-ns -o wide
kubectl get pods -n dns-test -o wide
```

#### Step 3: Engage AI Assistant for Diagnosis

Start a conversation with your AI assistant using prompts like:

**Initial Discovery:**
- "I have pods failing in my Kubernetes cluster. Can you help me diagnose what's wrong?"
- "Some applications are having DNS resolution issues. What should I check?"

**Specific Investigation:**
- "Can you check the DNS configuration in the secure-ns namespace?"
- "Why are pods in app-backend namespace failing to resolve DNS?"
- "There seem to be network connectivity issues between namespaces. Can you investigate?"

#### Step 4: Validate AI Diagnosis

The AI assistant should identify these issues:

1. **DNS Policy Misconfiguration** (secure-ns):
   - Wrong nameserver IP (8.8.8.8 instead of cluster DNS)
   - Incorrect search domains (production.local, company.internal)
   - Wrong ndots setting (2 instead of 5)

2. **Network Policy Blocking DNS** (app-backend):
   - NetworkPolicy preventing access to port 53
   - Pods unable to reach CoreDNS

3. **CoreDNS Issues**:
   - Custom configuration with invalid upstream servers
   - Potential CoreDNS pod failures

#### Step 5: Test AI-Suggested Fixes

Ask the AI assistant to help fix the issues:

- "How can I fix the DNS configuration in secure-ns?"
- "What changes are needed to allow DNS traffic in app-backend?"
- "Can you help me restore the CoreDNS configuration?"

The AI should provide kubectl commands to:
- Remove or modify the restrictive network policy
- Fix the DNS policy in pod specifications
- Restore CoreDNS configuration

#### Step 7: Verify Fixes

After applying suggested fixes:
```bash
kubectl get pods --all-namespaces
kubectl logs deployment/failing-worker -n app-backend
kubectl logs deployment/dns-dependent-app -n secure-ns
```
