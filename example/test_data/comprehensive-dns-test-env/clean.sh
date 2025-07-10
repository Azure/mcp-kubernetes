#!/bin/bash
set -euo pipefail

echo "Cleaning up DNS test environment..."
echo "=================================="

# Delete all test namespaces (this will delete all resources within them)
echo "Deleting test namespaces..."
kubectl delete namespace dns-test --ignore-not-found=true
kubectl delete namespace app-backend --ignore-not-found=true
kubectl delete namespace secure-ns --ignore-not-found=true

# Restore original CoreDNS custom configuration (empty it)
echo "Restoring original CoreDNS custom configuration..."
kubectl patch configmap coredns-custom -n kube-system --type merge -p '{"data":{}}'

# Restart CoreDNS to apply the clean configuration
echo "Restarting CoreDNS to apply clean configuration..."
kubectl rollout restart deployment/coredns -n kube-system

echo "Waiting for CoreDNS rollout to complete..."
kubectl rollout status deployment/coredns -n kube-system --timeout=60s

# Clean up any leftover pods that might be stuck
echo "Cleaning up any remaining test pods..."
kubectl delete pod debug-with-policy -n app-backend --ignore-not-found=true

echo ""
echo "Cleanup completed!"
echo "=================="
echo "All test namespaces and resources have been removed."
echo "CoreDNS configuration has been restored to default."
echo ""
echo "Verify cleanup:"
echo "kubectl get namespaces | grep -E '(dns-test|app-backend|secure-ns)'"
echo "kubectl get configmap coredns-custom -n kube-system -o yaml"
