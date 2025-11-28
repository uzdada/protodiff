#!/bin/bash
set -e

# ProtoDiff Development Environment Setup
# Sets up a local Kubernetes cluster and deploys sample services for testing

echo "================================="
echo "  ProtoDiff Dev Setup"
echo "================================="
echo ""

# Colors
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

# Check prerequisites
echo -e "${YELLOW}Checking prerequisites...${NC}"

if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl not found. Please install kubectl."
    exit 1
fi

if ! command -v docker &> /dev/null; then
    echo "Error: docker not found. Please install Docker."
    exit 1
fi

echo -e "${GREEN}✓ Prerequisites OK${NC}"
echo ""

# Step 1: Create kind cluster (optional)
if [[ "${CREATE_CLUSTER}" == "true" ]]; then
    echo -e "${YELLOW}Creating kind cluster...${NC}"

    if ! command -v kind &> /dev/null; then
        echo "Error: kind not found. Install with: go install sigs.k8s.io/kind@latest"
        exit 1
    fi

    cat <<EOF | kind create cluster --name protodiff-dev --config=-
kind: Cluster
apiVersion: kind.x-k8s.io/v1alpha4
nodes:
- role: control-plane
- role: worker
EOF

    echo -e "${GREEN}✓ Cluster created${NC}"
    echo ""
fi

# Step 2: Build and load image
echo -e "${YELLOW}Building ProtoDiff image...${NC}"
make docker-build

if [[ "${CREATE_CLUSTER}" == "true" ]]; then
    echo -e "${YELLOW}Loading image into kind cluster...${NC}"
    kind load docker-image protodiff:latest --name protodiff-dev
    echo -e "${GREEN}✓ Image loaded${NC}"
fi
echo ""

# Step 3: Deploy ProtoDiff
echo -e "${YELLOW}Deploying ProtoDiff...${NC}"
kubectl apply -f deploy/k8s/install.yaml

echo -e "${YELLOW}Waiting for ProtoDiff to be ready...${NC}"
kubectl wait --for=condition=ready pod \
    -l app.kubernetes.io/name=protodiff \
    -n protodiff-system \
    --timeout=60s

echo -e "${GREEN}✓ ProtoDiff deployed${NC}"
echo ""

# Step 4: Deploy sample gRPC service (optional)
if [[ "${DEPLOY_SAMPLES}" == "true" ]]; then
    echo -e "${YELLOW}Deploying sample services...${NC}"

    # Create a simple gRPC service pod for testing
    kubectl apply -f - <<EOF
apiVersion: v1
kind: Pod
metadata:
  name: sample-grpc-service
  labels:
    app: sample-service
    grpc-service: "true"
spec:
  containers:
  - name: grpc-server
    image: fullstorydev/grpcurl:latest
    command: ["/bin/sh"]
    args:
    - -c
    - |
      # Simple gRPC reflection server for testing
      echo "Starting sample gRPC service..."
      sleep infinity
    ports:
    - containerPort: 9090
      name: grpc
EOF

    echo -e "${GREEN}✓ Sample service deployed${NC}"
fi
echo ""

# Step 5: Show next steps
echo "================================="
echo -e "${GREEN}Development environment ready!${NC}"
echo "================================="
echo ""
echo "Next steps:"
echo ""
echo -e "${BLUE}1. Port-forward to dashboard:${NC}"
echo "   kubectl port-forward -n protodiff-system svc/protodiff 8080:80"
echo ""
echo -e "${BLUE}2. Open dashboard:${NC}"
echo "   http://localhost:8080"
echo ""
echo -e "${BLUE}3. View logs:${NC}"
echo "   make logs"
echo ""
echo -e "${BLUE}4. Configure mappings:${NC}"
echo "   kubectl edit configmap protodiff-mapping -n protodiff-system"
echo ""
echo -e "${BLUE}5. Deploy your gRPC services with label:${NC}"
echo "   grpc-service: \"true\""
echo ""
