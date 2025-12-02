#!/usr/bin/env bash

set -e

echo "=================================================="
echo "  ProtoDiff Complete Demo Setup"
echo "=================================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Track background processes
PIDS=()

# Cleanup function
cleanup() {
    echo ""
    echo "${YELLOW}Cleaning up background processes...${NC}"
    for pid in "${PIDS[@]}"; do
        if kill -0 "$pid" 2>/dev/null; then
            kill "$pid" 2>/dev/null || true
        fi
    done
    echo "${GREEN}Cleanup complete!${NC}"
    exit 0
}

# Set trap for cleanup on exit
trap cleanup SIGINT SIGTERM EXIT

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "${RED}Error: kubectl is not installed or not in PATH${NC}"
    exit 1
fi

# Check if cluster is accessible
if ! kubectl cluster-info &> /dev/null; then
    echo "${RED}Error: Cannot connect to Kubernetes cluster${NC}"
    exit 1
fi

echo "${GREEN}✓${NC} kubectl is available and cluster is accessible"
echo ""

# Step 1: Deploy test gRPC services
echo "=================================================="
echo "Step 1: Deploying test gRPC services"
echo "=================================================="
echo ""

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
SAMPLE_YAML="${SCRIPT_DIR}/sample-grpc-service.yaml"
INSTALL_YAML="${SCRIPT_DIR}/../deploy/k8s/install.yaml"

if [ ! -f "$SAMPLE_YAML" ]; then
    echo "${RED}Error: sample-grpc-service.yaml not found at ${SAMPLE_YAML}${NC}"
    exit 1
fi

echo "Applying sample-grpc-service.yaml..."
kubectl apply -f "$SAMPLE_YAML"
echo "${GREEN}✓${NC} Test services deployed"
echo ""

# Step 2: Deploy ProtoDiff
echo "=================================================="
echo "Step 2: Deploying ProtoDiff"
echo "=================================================="
echo ""

if [ ! -f "$INSTALL_YAML" ]; then
    echo "${RED}Error: install.yaml not found at ${INSTALL_YAML}${NC}"
    exit 1
fi

echo "Applying install.yaml..."
kubectl apply -f "$INSTALL_YAML"
echo "${GREEN}✓${NC} ProtoDiff deployed"
echo ""

# Step 3: Wait for pods to be ready
echo "=================================================="
echo "Step 3: Waiting for pods to be ready"
echo "=================================================="
echo ""

echo "Waiting for grpc-test namespace pods..."
kubectl wait --for=condition=ready pod -l app=grpc-server-go -n grpc-test --timeout=120s || {
    echo "${RED}Error: grpc-server-go pod not ready${NC}"
    kubectl get pods -n grpc-test
    exit 1
}
echo "${GREEN}✓${NC} grpc-server-go is ready"

kubectl wait --for=condition=ready pod -l app=grpc-server-java -n grpc-test --timeout=120s || {
    echo "${RED}Error: grpc-server-java pod not ready${NC}"
    kubectl get pods -n grpc-test
    exit 1
}
echo "${GREEN}✓${NC} grpc-server-java is ready"

echo ""
echo "Waiting for protodiff-system namespace pods..."
kubectl wait --for=condition=ready pod -l app.kubernetes.io/name=protodiff -n protodiff-system --timeout=120s || {
    echo "${RED}Error: protodiff pod not ready${NC}"
    kubectl get pods -n protodiff-system
    exit 1
}
echo "${GREEN}✓${NC} ProtoDiff is ready"
echo ""

# Step 4: Display pod status
echo "=================================================="
echo "Step 4: Pod Status"
echo "=================================================="
echo ""

echo "Test Services (grpc-test namespace):"
kubectl get pods -n grpc-test -o wide
echo ""

echo "ProtoDiff (protodiff-system namespace):"
kubectl get pods -n protodiff-system -o wide
echo ""

# Step 5: Port forwarding
echo "=================================================="
echo "Step 5: Setting up port forwarding"
echo "=================================================="
echo ""

echo "Starting port-forward for ProtoDiff dashboard (port 18080)..."
kubectl port-forward -n protodiff-system svc/protodiff 18080:80 > /dev/null 2>&1 &
PID1=$!
PIDS+=($PID1)
sleep 2

if kill -0 "$PID1" 2>/dev/null; then
    echo "${GREEN}✓${NC} ProtoDiff dashboard port-forward active on http://localhost:18080"
else
    echo "${YELLOW}Warning: Port-forward may have failed${NC}"
fi

echo ""
echo "Starting port-forward for grpc-server-go (port 9090)..."
kubectl port-forward -n grpc-test svc/grpc-server-go 9090:9090 > /dev/null 2>&1 &
PID2=$!
PIDS+=($PID2)
sleep 1

if kill -0 "$PID2" 2>/dev/null; then
    echo "${GREEN}✓${NC} grpc-server-go port-forward active on localhost:9090"
else
    echo "${YELLOW}Warning: Port-forward may have failed${NC}"
fi

echo ""
echo "Starting port-forward for grpc-server-java (port 9091)..."
kubectl port-forward -n grpc-test svc/grpc-server-java 9091:9091 > /dev/null 2>&1 &
PID3=$!
PIDS+=($PID3)
sleep 1

if kill -0 "$PID3" 2>/dev/null; then
    echo "${GREEN}✓${NC} grpc-server-java port-forward active on localhost:9091"
else
    echo "${YELLOW}Warning: Port-forward may have failed${NC}"
fi

echo ""

# Step 6: Summary
echo "=================================================="
echo "Demo Setup Complete! 🎉"
echo "=================================================="
echo ""
echo "${GREEN}Services are running:${NC}"
echo "  • ProtoDiff Dashboard: http://localhost:18080"
echo "  • Go gRPC Server: localhost:9090"
echo "  • Java gRPC Server: localhost:9091"
echo ""
echo "${YELLOW}Test gRPC services with grpcurl:${NC}"
echo "  grpcurl -plaintext localhost:9090 list"
echo "  grpcurl -plaintext -d '{\"name\": \"World\"}' localhost:9090 greeter.Greeter/SayHello"
echo "  grpcurl -plaintext -d '{\"user_id\": 1}' localhost:9090 greeter.Greeter/SayHelloToUser"
echo ""
echo "  grpcurl -plaintext localhost:9091 list"
echo "  grpcurl -plaintext -d '{\"user_id\": 1}' localhost:9091 user.UserService/GetUser"
echo ""
echo "${YELLOW}View ProtoDiff logs:${NC}"
echo "  kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f"
echo ""
echo "${RED}Press Ctrl+C to stop all port-forwards and exit${NC}"
echo ""

# Try to open browser (optional, might not work in all environments)
if command -v open &> /dev/null; then
    echo "Opening dashboard in browser..."
    sleep 2
    open http://localhost:18080 2>/dev/null || true
elif command -v xdg-open &> /dev/null; then
    echo "Opening dashboard in browser..."
    sleep 2
    xdg-open http://localhost:18080 2>/dev/null || true
fi

# Keep script running until interrupted
echo "Monitoring services... (Ctrl+C to exit)"
while true; do
    sleep 1
done
