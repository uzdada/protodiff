# BSR Integration Guide

[English](#english) | [한국어](#korean)

---

## English

### Overview

ProtoDiff supports two modes for BSR (Buf Schema Registry) integration:

1. **Mock Mode** - Uses hardcoded sample schemas (for testing)
2. **HTTP Mode** - Connects to real BSR API (for production)

### Configuration

#### Mock Mode (Default for Testing)

```bash
# Set environment variable
export USE_MOCK_BSR=true

# Or in Kubernetes Deployment
env:
  - name: USE_MOCK_BSR
    value: "true"
```

Mock mode uses predefined schemas and doesn't require BSR account.

#### HTTP Mode (Production)

```bash
# Do not set USE_MOCK_BSR or set to false
export USE_MOCK_BSR=false

# Set BSR token (required for private modules)
export BSR_TOKEN=your-token-here
```

### Getting BSR Token

1. Visit https://buf.build/settings/user
2. Click "Create Token"
3. Copy the generated token
4. Store securely

### Kubernetes Setup

#### Step 1: Create Secret

```bash
kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_BSR_TOKEN \
  -n protodiff-system
```

#### Step 2: Update Deployment

Edit `deploy/k8s/install.yaml` and add to the container env section:

```yaml
spec:
  template:
    spec:
      containers:
      - name: protodiff
        env:
        # Use real BSR client
        - name: USE_MOCK_BSR
          value: "false"
        # BSR authentication token
        - name: BSR_TOKEN
          valueFrom:
            secretKeyRef:
              name: bsr-token
              key: token
        # Optional: Custom BSR URL
        - name: BSR_URL
          value: "https://buf.build"
```

#### Step 3: Apply Changes

```bash
kubectl apply -f deploy/k8s/install.yaml
kubectl rollout restart deployment/protodiff -n protodiff-system
```

### Pushing Schemas to BSR

Before ProtoDiff can validate schemas, you need to push them to BSR:

#### Install Buf CLI

```bash
# macOS
brew install bufbuild/buf/buf

# Linux
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
```

#### Login to BSR

```bash
buf registry login
```

#### Push Schemas

```bash
# For grpc-server-go
cd /path/to/grpc-server-go
buf push proto

# For grpc-server-java
cd /path/to/grpc-server-java
buf push src/main/proto
```

### Verification

Check if ProtoDiff can fetch schemas:

```bash
# View logs
kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f

# Look for:
# "BSR client initialized (HTTP mode)"
# "Fetching schema from BSR: buf.build/yourname/greeter"
```

### Troubleshooting

#### Error: "BSR API returned status 401"

**Cause**: Invalid or missing BSR token

**Solution**:
```bash
# Verify secret exists
kubectl get secret bsr-token -n protodiff-system

# Recreate with correct token
kubectl delete secret bsr-token -n protodiff-system
kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_CORRECT_TOKEN \
  -n protodiff-system

# Restart deployment
kubectl rollout restart deployment/protodiff -n protodiff-system
```

#### Error: "schema not found for module"

**Cause**: Schema not pushed to BSR or wrong module name

**Solution**:
```bash
# Verify module exists on BSR
buf registry module list buf.build/yourname/yourmodule

# Push if missing
cd /path/to/proto
buf push .

# Check ConfigMap mapping
kubectl get configmap protodiff-mapping -n protodiff-system -o yaml
```

#### Error: "connection refused" or timeout

**Cause**: Network issues or BSR_URL misconfigured

**Solution**:
```bash
# Test network connectivity from pod
kubectl exec -it deployment/protodiff -n protodiff-system -- \
  wget -O- https://buf.build

# Verify BSR_URL environment variable
kubectl describe deployment protodiff -n protodiff-system | grep BSR_URL
```

### API Details

ProtoDiff uses the BSR FileDescriptorSet API:

- **Endpoint**: `https://buf.build/buf.reflect.v1beta1.FileDescriptorSetService/GetFileDescriptorSet`
- **Method**: POST
- **Content-Type**: application/json
- **Authentication**: Bearer token
- **Request**: `{"module": "buf.build/owner/repo"}`

### References

- [BSR Documentation](https://buf.build/docs/bsr/)
- [BSR API Access](https://docs.bufbuild.ru/bsr/apis/api-access/)
- [Buf CLI Guide](https://docs.buf.build/installation)

---

## Korean

### 개요

ProtoDiff는 BSR(Buf Schema Registry) 연동을 위한 두 가지 모드를 지원합니다:

1. **Mock 모드** - 하드코딩된 샘플 스키마 사용 (테스트용)
2. **HTTP 모드** - 실제 BSR API 연결 (프로덕션용)

### 설정

#### Mock 모드 (테스트 기본값)

```bash
# 환경 변수 설정
export USE_MOCK_BSR=true

# 또는 Kubernetes Deployment에서
env:
  - name: USE_MOCK_BSR
    value: "true"
```

Mock 모드는 미리 정의된 스키마를 사용하며 BSR 계정이 필요 없습니다.

#### HTTP 모드 (프로덕션)

```bash
# USE_MOCK_BSR를 설정하지 않거나 false로 설정
export USE_MOCK_BSR=false

# BSR 토큰 설정 (비공개 모듈에 필요)
export BSR_TOKEN=your-token-here
```

### BSR 토큰 얻기

1. https://buf.build/settings/user 방문
2. "Create Token" 클릭
3. 생성된 토큰 복사
4. 안전하게 보관

### Kubernetes 설정

#### 1단계: Secret 생성

```bash
kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_BSR_TOKEN \
  -n protodiff-system
```

#### 2단계: Deployment 업데이트

`deploy/k8s/install.yaml`을 편집하고 container env 섹션에 추가:

```yaml
spec:
  template:
    spec:
      containers:
      - name: protodiff
        env:
        # 실제 BSR 클라이언트 사용
        - name: USE_MOCK_BSR
          value: "false"
        # BSR 인증 토큰
        - name: BSR_TOKEN
          valueFrom:
            secretKeyRef:
              name: bsr-token
              key: token
        # 선택사항: 커스텀 BSR URL
        - name: BSR_URL
          value: "https://buf.build"
```

#### 3단계: 변경사항 적용

```bash
kubectl apply -f deploy/k8s/install.yaml
kubectl rollout restart deployment/protodiff -n protodiff-system
```

### BSR에 스키마 푸시

ProtoDiff가 스키마를 검증하려면 먼저 BSR에 푸시해야 합니다:

#### Buf CLI 설치

```bash
# macOS
brew install bufbuild/buf/buf

# Linux
curl -sSL "https://github.com/bufbuild/buf/releases/latest/download/buf-$(uname -s)-$(uname -m)" \
  -o /usr/local/bin/buf
chmod +x /usr/local/bin/buf
```

#### BSR 로그인

```bash
buf registry login
```

#### 스키마 푸시

```bash
# grpc-server-go의 경우
cd /path/to/grpc-server-go
buf push proto

# grpc-server-java의 경우
cd /path/to/grpc-server-java
buf push src/main/proto
```

### 검증

ProtoDiff가 스키마를 가져올 수 있는지 확인:

```bash
# 로그 확인
kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f

# 다음 메시지 확인:
# "BSR client initialized (HTTP mode)"
# "Fetching schema from BSR: buf.build/yourname/greeter"
```

### 문제 해결

#### 오류: "BSR API returned status 401"

**원인**: 잘못되었거나 누락된 BSR 토큰

**해결**:
```bash
# Secret 존재 확인
kubectl get secret bsr-token -n protodiff-system

# 올바른 토큰으로 재생성
kubectl delete secret bsr-token -n protodiff-system
kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_CORRECT_TOKEN \
  -n protodiff-system

# Deployment 재시작
kubectl rollout restart deployment/protodiff -n protodiff-system
```

#### 오류: "schema not found for module"

**원인**: BSR에 스키마가 푸시되지 않았거나 잘못된 모듈 이름

**해결**:
```bash
# BSR에 모듈 존재 확인
buf registry module list buf.build/yourname/yourmodule

# 누락된 경우 푸시
cd /path/to/proto
buf push .

# ConfigMap 매핑 확인
kubectl get configmap protodiff-mapping -n protodiff-system -o yaml
```

### API 세부사항

ProtoDiff는 BSR FileDescriptorSet API를 사용합니다:

- **엔드포인트**: `https://buf.build/buf.reflect.v1beta1.FileDescriptorSetService/GetFileDescriptorSet`
- **메서드**: POST
- **Content-Type**: application/json
- **인증**: Bearer 토큰
- **요청**: `{"module": "buf.build/owner/repo"}`

### 참고자료

- [BSR 문서](https://buf.build/docs/bsr/)
- [BSR API 접근](https://docs.bufbuild.ru/bsr/apis/api-access/)
- [Buf CLI 가이드](https://docs.buf.build/installation)
