# ProtoDiff Quick Start Guide

[English](#english) | [한국어](#korean)

---

## English

This guide will help you get ProtoDiff up and running in 5 minutes.

### Prerequisites

- Kubernetes cluster (v1.25+)
- kubectl configured and working
- At least one gRPC service running with reflection enabled
- **BSR Token** from https://buf.build/settings/user

### Step 1: Configure BSR Token

Before deploying, get your BSR token and configure it:

```bash
# Download the install manifest
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml

# Edit and replace YOUR_BSR_TOKEN_HERE with your actual token
# Find the Secret named 'bsr-token' (around line 73-86)
vi deploy/k8s/install.yaml
```

Look for this section:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: bsr-token
stringData:
  token: "YOUR_BSR_TOKEN_HERE"  # <-- Replace this
```

To get your BSR token:
1. Visit https://buf.build/settings/user
2. Click "Create Token"
3. Copy the generated token
4. Paste it into the install.yaml

### Step 2: Deploy ProtoDiff

```bash
kubectl apply -f deploy/k8s/install.yaml
```

This creates:
- `protodiff-system` namespace
- ServiceAccount with necessary permissions
- ConfigMap for service mappings
- ProtoDiff deployment and service

Verify the deployment:

```bash
kubectl get pods -n protodiff-system
```

Expected output:
```
NAME                        READY   STATUS    RESTARTS   AGE
protodiff-xxxxxxxxx-xxxxx   1/1     Running   0          30s
```

### Step 3: Configure Your Services

#### Option A: Using ConfigMap (Recommended)

Edit the ConfigMap to map your service names to BSR modules:

```bash
kubectl edit configmap protodiff-mapping -n protodiff-system
```

Add your services:

```yaml
data:
  my-user-service: "buf.build/myorg/user"
  my-order-service: "buf.build/myorg/order"
```

#### Option B: Using Template

Set an environment variable for automatic mapping:

```bash
kubectl set env deployment/protodiff \
  -n protodiff-system \
  DEFAULT_BSR_TEMPLATE="buf.build/myorg/{service}"
```

This automatically maps:
- `user-service` → `buf.build/myorg/user-service`
- `order-service` → `buf.build/myorg/order-service`

### Step 4: Label Your gRPC Pods

Add the `grpc-service=true` label to your gRPC service deployments:

```bash
kubectl patch deployment my-user-service \
  -p '{"spec":{"template":{"metadata":{"labels":{"grpc-service":"true"}}}}}'
```

Or edit your deployment YAML:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-user-service
spec:
  template:
    metadata:
      labels:
        app: my-user-service
        grpc-service: "true"
```

### Step 5: Access the Dashboard

Port-forward to the ProtoDiff service:

```bash
kubectl port-forward -n protodiff-system svc/protodiff 8080:80
```

Open your browser to: http://localhost:8080

You should see the ProtoDiff dashboard with your services.

### Step 5: Verify It's Working

Check the ProtoDiff logs:

```bash
kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f
```

Expected log output:

```
Starting ProtoDiff - gRPC Schema Drift Monitor
Configuration loaded:
  ConfigMap: protodiff-system/protodiff-mapping
  BSR Template: buf.build/myorg/{service}
  Web Address: :8080
  Scan Interval: 30s
Kubernetes client initialized
gRPC reflection client initialized
BSR client initialized (mock mode)
Starting web server on :8080
Starting scanner with interval: 30s
Starting scan cycle...
Discovered 3 gRPC pods
Validated default/my-user-service-abc123: SYNC
Validated default/my-order-service-def456: SYNC
Scan cycle completed. Results stored: 3
```

### Common Issues

#### No services discovered

**Symptom**: Dashboard shows "No gRPC services discovered yet"

**Solutions**:
1. Verify pods have `grpc-service=true` label:
   ```bash
   kubectl get pods -l grpc-service=true --all-namespaces
   ```
2. Check if pods are running:
   ```bash
   kubectl get pods
   ```

#### Status shows UNKNOWN

**Symptom**: Services show yellow "UNKNOWN" status

**Solutions**:
1. Verify gRPC reflection is enabled in your service
2. Check logs for connection errors:
   ```bash
   kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff
   ```

#### No BSR mapping found

**Symptom**: Message says "No BSR module mapping found"

**Solutions**:
1. Add service to ConfigMap:
   ```bash
   kubectl edit configmap protodiff-mapping -n protodiff-system
   ```
2. Or set `DEFAULT_BSR_TEMPLATE` environment variable

### Next Steps

- Read the full documentation: [README.md](../README.md)
- Configure advanced mappings: [examples/configmap-advanced.yaml](../examples/configmap-advanced.yaml)
- See example service configuration: [examples/sample-grpc-service.yaml](../examples/sample-grpc-service.yaml)
- Learn about contributing: [CONTRIBUTING.md](../CONTRIBUTING.md)

### Cleanup

To remove ProtoDiff from your cluster:

```bash
kubectl delete -f deploy/k8s/install.yaml
```

To remove labels from your services:

```bash
kubectl patch deployment my-user-service \
  -p '{"spec":{"template":{"metadata":{"labels":{"grpc-service":null}}}}}'
```

---

## Korean

본 가이드는 5분 안에 ProtoDiff를 시작하고 실행하는 데 도움을 드립니다.

### 사전 요구사항

- Kubernetes 클러스터 (v1.25+)
- kubectl 설정 및 작동 확인
- 리플렉션이 활성화된 최소 하나의 gRPC 서비스 실행 중
- **BSR 토큰** (https://buf.build/settings/user 에서 발급)

### 단계 1: BSR 토큰 설정

배포 전에 BSR 토큰을 발급받고 설정합니다:

```bash
# install manifest 다운로드
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml

# YOUR_BSR_TOKEN_HERE를 실제 토큰으로 교체
# 'bsr-token'이라는 이름의 Secret을 찾으세요 (약 73-86번째 줄)
vi deploy/k8s/install.yaml
```

다음 섹션을 찾으세요:

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: bsr-token
stringData:
  token: "YOUR_BSR_TOKEN_HERE"  # <-- 여기를 교체
```

BSR 토큰 발급 방법:
1. https://buf.build/settings/user 방문
2. "Create Token" 클릭
3. 생성된 토큰 복사
4. install.yaml에 붙여넣기

### 단계 2: ProtoDiff 배포

```bash
kubectl apply -f deploy/k8s/install.yaml
```

다음이 생성됩니다:
- `protodiff-system` 네임스페이스
- 필요한 권한을 가진 ServiceAccount
- 서비스 매핑을 위한 ConfigMap
- ProtoDiff 디플로이먼트 및 서비스

배포 확인:

```bash
kubectl get pods -n protodiff-system
```

예상 출력:
```
NAME                        READY   STATUS    RESTARTS   AGE
protodiff-xxxxxxxxx-xxxxx   1/1     Running   0          30s
```

### 단계 3: 서비스 설정

#### 옵션 A: ConfigMap 사용 (권장)

ConfigMap을 편집하여 서비스 이름을 BSR 모듈에 매핑:

```bash
kubectl edit configmap protodiff-mapping -n protodiff-system
```

서비스 추가:

```yaml
data:
  my-user-service: "buf.build/myorg/user"
  my-order-service: "buf.build/myorg/order"
```

#### 옵션 B: 템플릿 사용

자동 매핑을 위한 환경 변수 설정:

```bash
kubectl set env deployment/protodiff \
  -n protodiff-system \
  DEFAULT_BSR_TEMPLATE="buf.build/myorg/{service}"
```

자동으로 매핑됩니다:
- `user-service` → `buf.build/myorg/user-service`
- `order-service` → `buf.build/myorg/order-service`

### 단계 4: gRPC Pod에 레이블 추가

gRPC 서비스 디플로이먼트에 `grpc-service=true` 레이블 추가:

```bash
kubectl patch deployment my-user-service \
  -p '{"spec":{"template":{"metadata":{"labels":{"grpc-service":"true"}}}}}'
```

또는 디플로이먼트 YAML 편집:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: my-user-service
spec:
  template:
    metadata:
      labels:
        app: my-user-service
        grpc-service: "true"
```

### 단계 5: 대시보드 접속

ProtoDiff 서비스로 포트 포워드:

```bash
kubectl port-forward -n protodiff-system svc/protodiff 8080:80
```

브라우저에서 http://localhost:8080 열기

서비스가 포함된 ProtoDiff 대시보드가 표시됩니다.

### 단계 5: 작동 확인

ProtoDiff 로그 확인:

```bash
kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f
```

예상 로그 출력:

```
Starting ProtoDiff - gRPC Schema Drift Monitor
Configuration loaded:
  ConfigMap: protodiff-system/protodiff-mapping
  BSR Template: buf.build/myorg/{service}
  Web Address: :8080
  Scan Interval: 30s
Kubernetes client initialized
gRPC reflection client initialized
BSR client initialized (mock mode)
Starting web server on :8080
Starting scanner with interval: 30s
Starting scan cycle...
Discovered 3 gRPC pods
Validated default/my-user-service-abc123: SYNC
Validated default/my-order-service-def456: SYNC
Scan cycle completed. Results stored: 3
```

### 일반적인 문제

#### 서비스가 발견되지 않음

**증상**: 대시보드에 "No gRPC services discovered yet" 표시

**해결 방법**:
1. Pod에 `grpc-service=true` 레이블이 있는지 확인:
   ```bash
   kubectl get pods -l grpc-service=true --all-namespaces
   ```
2. Pod가 실행 중인지 확인:
   ```bash
   kubectl get pods
   ```

#### 상태가 UNKNOWN으로 표시

**증상**: 서비스가 노란색 "UNKNOWN" 상태로 표시

**해결 방법**:
1. 서비스에서 gRPC 리플렉션이 활성화되어 있는지 확인
2. 연결 오류에 대한 로그 확인:
   ```bash
   kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff
   ```

#### BSR 매핑을 찾을 수 없음

**증상**: "No BSR module mapping found" 메시지

**해결 방법**:
1. ConfigMap에 서비스 추가:
   ```bash
   kubectl edit configmap protodiff-mapping -n protodiff-system
   ```
2. 또는 `DEFAULT_BSR_TEMPLATE` 환경 변수 설정

### 다음 단계

- 전체 문서 읽기: [README.md](../README.md)
- 고급 매핑 설정: [examples/configmap-advanced.yaml](../examples/configmap-advanced.yaml)
- 예제 서비스 설정 보기: [examples/sample-grpc-service.yaml](../examples/sample-grpc-service.yaml)
- 기여에 대해 알아보기: [CONTRIBUTING.md](../CONTRIBUTING.md)

### 정리

클러스터에서 ProtoDiff 제거:

```bash
kubectl delete -f deploy/k8s/install.yaml
```

서비스에서 레이블 제거:

```bash
kubectl patch deployment my-user-service \
  -p '{"spec":{"template":{"metadata":{"labels":{"grpc-service":null}}}}}'
```
