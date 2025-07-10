#!/bin/bash
set -euo pipefail

echo "Deploying DNS test environment with intentional DNS issues..."
echo "================================================================"

# Create namespaces
echo "Creating namespaces..."
kubectl apply -f namespace.yaml

# Deploy applications
echo "Deploying web application..."
kubectl apply -f web-app.yaml

echo "Deploying backend application..."
kubectl apply -f backend-app.yaml

echo "Deploying secure application (with DNS misconfiguration)..."
kubectl apply -f secure-app.yaml

# Apply restrictive network policy
echo "Applying restrictive network policy (will block DNS)..."
kubectl apply -f restrictive-network-policy.yaml

# Deploy failing workload pods
echo "Deploying workload pods that will fail due to DNS issues..."
kubectl apply -f failing-workloads.yaml

# Apply the problematic CoreDNS custom configuration
echo "Applying problematic CoreDNS custom configuration..."
kubectl apply -f coredns-custom-config.yaml

# Restart CoreDNS pods to pick up the new configuration
echo "Restarting CoreDNS pods to apply new configuration..."
kubectl rollout restart deployment/coredns -n kube-system

echo "Waiting for CoreDNS rollout to complete..."
kubectl rollout status deployment/coredns -n kube-system --timeout=60s

echo "Waiting for pods to be ready..."
sleep 10

echo "Checking pod status..."
kubectl get pods -n dns-test
kubectl get pods -n app-backend  
kubectl get pods -n secure-ns

echo ""
echo "DNS Test Environment deployed!"
echo "==============================================="
echo "Known issues introduced:"
echo "1. secure-ns pods have incorrect DNS configuration (wrong nameserver)"
echo "2. app-backend namespace has NetworkPolicy blocking DNS traffic"
echo "3. CoreDNS custom config applied with invalid upstream DNS (192.168.999.999)"
echo ""
echo "Observe failing workloads:"
echo "kubectl get pods -n dns-test -w"
echo "kubectl get pods -n app-backend -w" 
echo "kubectl get pods -n secure-ns -w"
echo ""
echo "Check pod logs to see DNS failures:"
echo "kubectl logs -f deployment/failing-worker -n app-backend"
echo "kubectl logs -f deployment/dns-dependent-app -n secure-ns"
echo "kubectl logs -f deployment/health-check-app -n dns-test"
