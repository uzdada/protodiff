#!/bin/bash
set -e

# ProtoDiff Build Script
# Builds the binary, Docker image, and optionally deploys to Kubernetes

VERSION="${VERSION:-latest}"
REGISTRY="${REGISTRY:-protodiff}"
IMAGE_NAME="${REGISTRY}:${VERSION}"

echo "================================="
echo "  ProtoDiff Build Script"
echo "================================="
echo "Version: ${VERSION}"
echo "Image: ${IMAGE_NAME}"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Download dependencies
echo -e "${YELLOW}[1/5] Downloading dependencies...${NC}"
go mod download
go mod tidy
echo -e "${GREEN}✓ Dependencies downloaded${NC}"
echo ""

# Step 2: Run tests
echo -e "${YELLOW}[2/5] Running tests...${NC}"
go test -v ./...
echo -e "${GREEN}✓ Tests passed${NC}"
echo ""

# Step 3: Build binary
echo -e "${YELLOW}[3/5] Building binary...${NC}"
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-w -s -X main.Version=${VERSION}" \
    -o bin/protodiff \
    ./cmd/protodiff
echo -e "${GREEN}✓ Binary built: bin/protodiff${NC}"
echo ""

# Step 4: Build Docker image
echo -e "${YELLOW}[4/5] Building Docker image...${NC}"
docker build -t "${IMAGE_NAME}" .
echo -e "${GREEN}✓ Docker image built: ${IMAGE_NAME}${NC}"
echo ""

# Step 5: Optional deployment
if [[ "${DEPLOY}" == "true" ]]; then
    echo -e "${YELLOW}[5/5] Deploying to Kubernetes...${NC}"
    kubectl apply -f deploy/k8s/install.yaml
    echo -e "${GREEN}✓ Deployed to Kubernetes${NC}"
else
    echo -e "${YELLOW}[5/5] Skipping deployment (set DEPLOY=true to deploy)${NC}"
fi

echo ""
echo "================================="
echo -e "${GREEN}Build completed successfully!${NC}"
echo "================================="
echo ""
echo "Next steps:"
echo "  1. Push image: docker push ${IMAGE_NAME}"
echo "  2. Deploy: kubectl apply -f deploy/k8s/install.yaml"
echo "  3. Port-forward: kubectl port-forward -n protodiff-system svc/protodiff 18080:80"
echo "  4. Open: http://localhost:18080"
echo ""
