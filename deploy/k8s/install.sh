#!/usr/bin/env bash

set -e

echo "====================================="
echo "  ProtoDiff Installation Script"
echo "====================================="
echo ""

# Check if kubectl is installed
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed. Please install kubectl first."
    exit 1
fi

# Check if kubectl can access the cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: Cannot connect to Kubernetes cluster. Please check your kubeconfig."
    exit 1
fi

echo "Step 1: BSR Token Configuration"
echo "--------------------------------"
echo "Get your BSR token from: https://buf.build/settings/user"
echo ""
read -sp "Enter your BSR token: " BSR_TOKEN
echo ""

if [ -z "$BSR_TOKEN" ]; then
    echo "Error: BSR token cannot be empty"
    exit 1
fi

echo ""
echo "Step 2: Service Mapping Configuration"
echo "--------------------------------------"
echo "Map your gRPC service names to BSR modules."
echo "Format: service-name=buf.build/org/module"
echo "Example: user-service=buf.build/acme/user"
echo ""
echo "Enter service mappings (one per line, empty line to finish):"
echo ""

declare -a MAPPINGS=()
while true; do
    read -p "Service mapping (or press Enter to finish): " MAPPING
    if [ -z "$MAPPING" ]; then
        break
    fi
    MAPPINGS+=("$MAPPING")
done

if [ ${#MAPPINGS[@]} -eq 0 ]; then
    echo ""
    echo "Warning: No service mappings provided."
    echo "You can add them later by editing the ConfigMap:"
    echo "  kubectl edit configmap protodiff-mapping -n protodiff-system"
    echo ""
fi

echo ""
echo "Step 3: Deploying ProtoDiff"
echo "----------------------------"

# Apply the main installation manifest
echo "Applying Kubernetes manifests..."
kubectl apply -f https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml

# Wait a moment for namespace to be created
sleep 2

# Create the BSR token secret
echo "Creating BSR token secret..."
kubectl create secret generic bsr-token \
  --from-literal=token="$BSR_TOKEN" \
  -n protodiff-system \
  --dry-run=client -o yaml | kubectl apply -f -

# Update ConfigMap with service mappings if provided
if [ ${#MAPPINGS[@]} -gt 0 ]; then
    echo "Updating service mappings..."

    # Build the ConfigMap data section
    DATA_SECTION=""
    for MAPPING in "${MAPPINGS[@]}"; do
        # Split by '=' to get service name and BSR module
        SERVICE=$(echo "$MAPPING" | cut -d'=' -f1 | xargs)
        MODULE=$(echo "$MAPPING" | cut -d'=' -f2- | xargs)
        DATA_SECTION="${DATA_SECTION}  ${SERVICE}: \"${MODULE}\"\n"
    done

    # Create ConfigMap with mappings
    cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ConfigMap
metadata:
  name: protodiff-mapping
  namespace: protodiff-system
  labels:
    app.kubernetes.io/name: protodiff
    app.kubernetes.io/component: config
data:
$(echo -e "$DATA_SECTION")
EOF
fi

echo ""
echo "✓ ProtoDiff has been successfully installed!"
echo ""
echo "Next steps:"
echo "  1. Label your gRPC pods with 'grpc-service=true'"
echo "  2. Access the dashboard:"
echo "       kubectl port-forward -n protodiff-system svc/protodiff 8080:80"
echo "  3. Open http://localhost:8080 in your browser"
echo ""
echo "To add more service mappings later:"
echo "  kubectl edit configmap protodiff-mapping -n protodiff-system"
echo ""
