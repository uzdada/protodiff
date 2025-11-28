# ProtoDiff - Build Success

[English](#english) | [한국어](#korean)

---

## English

### Project Status

Complete and functional Kubernetes-native gRPC schema drift monitoring agent successfully built.

### Build Results

```
Binary Location: bin/protodiff
File Size: 50MB
Type: Mach-O 64-bit executable (macOS)
Build Date: 2025-11-28
```

### Issues Resolved

The following issues were addressed during the build process:

#### 1. go.sum File Generation

**Problem**: Empty go.sum file causing errors
```bash
malformed go.sum: wrong number of fields
```

**Solution**:
```bash
go mod download  # Download dependencies
go mod tidy      # Auto-generate go.sum
```

#### 2. go:embed Path Correction

**Problem**: Invalid embed path
```
//go:embed ../../../web/templates/index.html  # Incorrect
```

**Solution**:
```
//go:embed templates/index.html  # Correct
```
- Moved template to `internal/adapters/web/templates/`
- Adjusted relative path based on package location

#### 3. gRPC API Compatibility

**Problem**: gRPC version compatibility issue
```go
grpc.NewClient()  // Latest API, not supported in current version
```

**Solution**:
```go
grpc.Dial()  // Use stable API
grpcreflect.NewClientV1Alpha()  # Explicit version
```

#### 4. Unused Import Removal

**Problem**: Unused domain import
```go
"github.com/uzdada/protodiff/internal/core/domain"  # Unused
```

**Solution**: Removed from k8s/client.go

### Build Instructions

#### Quick Build
```bash
cd /path/to/proto-diff
go build -o bin/protodiff ./cmd/protodiff
```

#### Using Makefile
```bash
make build
```

#### Optimized Build (Smaller Binary)
```bash
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/protodiff ./cmd/protodiff
```

Flags explained:
- `CGO_ENABLED=0`: Pure Go binary (no dependencies)
- `-w`: Remove DWARF debug info
- `-s`: Remove symbol table

### Docker Image Build

#### Local Build
```bash
docker build -t protodiff:latest .
```

#### Multi-platform Build
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t protodiff:latest \
  .
```

### Running Tests

#### All Tests
```bash
go test -v ./...
```

#### With Coverage
```bash
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### Specific Package
```bash
go test -v ./internal/core/store
go test -v ./internal/scanner
```

### Dependency Information

#### Main Dependencies

```
Direct Dependencies:
- github.com/jhump/protoreflect v1.15.6  (gRPC Reflection)
- google.golang.org/grpc v1.62.0         (gRPC Client)
- k8s.io/api v0.29.0                     (Kubernetes API Types)
- k8s.io/apimachinery v0.29.0            (Kubernetes Meta Types)
- k8s.io/client-go v0.29.0               (Kubernetes Client)

Indirect Dependencies: 40+ packages
```

#### Dependency Management
```bash
# View dependency tree
go mod graph | grep "github.com/uzdada/protodiff"

# Update dependencies
go get -u ./...
go mod tidy
```

### Build Verification

#### Binary Information
```bash
file bin/protodiff
# Output: Mach-O 64-bit executable x86_64

ls -lh bin/protodiff
# Output: 50M
```

### Next Steps

#### 1. Local Testing
```bash
# Requires Kubernetes (minikube, kind, Docker Desktop, etc.)

# 1. Verify cluster
kubectl cluster-info

# 2. Deploy
kubectl apply -f deploy/k8s/install.yaml

# 3. Check logs
kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f

# 4. Access dashboard
kubectl port-forward -n protodiff-system svc/protodiff 8080:80
open http://localhost:8080
```

#### 2. Docker Image Testing
```bash
# Build image
docker build -t protodiff:test .

# Note: Requires Kubernetes API access, works only inside cluster
```

#### 3. Production Deployment

##### a) Push to Image Registry
```bash
# Docker Hub
docker tag protodiff:latest uzdada/protodiff:v1.0.0
docker push uzdada/protodiff:v1.0.0

# GitHub Container Registry
docker tag protodiff:latest ghcr.io/uzdada/protodiff:v1.0.0
docker push ghcr.io/uzdada/protodiff:v1.0.0
```

##### b) Update Deployment
```yaml
# deploy/k8s/install.yaml
spec:
  template:
    spec:
      containers:
        - name: protodiff
          image: uzdada/protodiff:v1.0.0  # Specify version
          imagePullPolicy: Always                # Production recommended
```

##### c) Deploy and Monitor
```bash
kubectl apply -f deploy/k8s/install.yaml
kubectl rollout status deployment/protodiff -n protodiff-system
```

### Troubleshooting

#### Build Errors

**Symptom**: `undefined: grpc.XXX`
```bash
# Solution: Re-download dependencies
go mod download
go mod tidy
go clean -modcache
```

**Symptom**: `go.sum: checksum mismatch`
```bash
# Solution: Regenerate go.sum
rm go.sum
go mod download
go mod tidy
```

**Symptom**: `cannot find package`
```bash
# Solution: Verify Go module mode
export GO111MODULE=on
go mod download
```

#### Runtime Errors

**Symptom**: `permission denied` (Kubernetes API)
```bash
# Check RBAC permissions
kubectl describe clusterrole protodiff
kubectl describe clusterrolebinding protodiff
```

**Symptom**: `connection refused` (gRPC Reflection)
```bash
# Check pod-to-pod networking
kubectl exec -it <pod> -- nc -zv <target-pod-ip> 9090

# Verify gRPC Reflection is enabled
grpcurl -plaintext <pod-ip>:9090 list
```

### Performance Optimization

#### Reduce Binary Size

```bash
# UPX compression (optional)
upx --best --lzma bin/protodiff
# 50MB → ~15MB
```

#### Reduce Memory Usage
```yaml
# deploy/k8s/install.yaml
resources:
  requests:
    memory: 64Mi   # Test minimum value
  limits:
    memory: 256Mi  # Safe limit
```

#### Adjust Scan Interval
```yaml
env:
  - name: SCAN_INTERVAL
    value: "60s"  # 30s → 60s (reduce load)
```

### Checklist

Project completion status:

- [x] Go code builds successfully
- [x] Dependencies cleaned up (go.sum)
- [x] Docker image builds
- [x] Kubernetes manifests created
- [x] Documentation complete
- [ ] Unit tests written
- [ ] Integration tests
- [ ] E2E tests
- [ ] CI/CD pipeline configured
- [ ] Production deployment

---

## Korean

### 프로젝트 상태

완전하고 기능적인 Kubernetes 네이티브 gRPC 스키마 드리프트 모니터링 에이전트가 성공적으로 빌드되었습니다.

### 빌드 결과

```
바이너리 위치: bin/protodiff
파일 크기: 50MB
타입: Mach-O 64-bit executable (macOS)
빌드 일시: 2025-11-28
```

### 해결된 문제

빌드 과정에서 다음 문제들이 해결되었습니다:

#### 1. go.sum 파일 생성

**문제**: 빈 go.sum 파일로 인한 오류
```bash
malformed go.sum: wrong number of fields
```

**해결**:
```bash
go mod download  # 의존성 다운로드
go mod tidy      # go.sum 자동 생성
```

#### 2. go:embed 경로 수정

**문제**: 잘못된 embed 경로
```
//go:embed ../../../web/templates/index.html  # 잘못됨
```

**해결**:
```
//go:embed templates/index.html  # 올바름
```
- 템플릿을 `internal/adapters/web/templates/`로 이동
- 패키지 위치 기준으로 상대 경로 조정

#### 3. gRPC API 호환성

**문제**: gRPC 버전 호환성 이슈
```go
grpc.NewClient()  // 최신 API, 현재 버전에서 미지원
```

**해결**:
```go
grpc.Dial()  // 안정적인 API 사용
grpcreflect.NewClientV1Alpha()  # 명시적 버전 지정
```

#### 4. 불필요한 import 제거

**문제**: 사용되지 않는 domain import
```go
"github.com/uzdada/protodiff/internal/core/domain"  # 미사용
```

**해결**: k8s/client.go에서 제거

### 빌드 방법

#### 빠른 빌드
```bash
cd /path/to/proto-diff
go build -o bin/protodiff ./cmd/protodiff
```

#### Makefile 사용
```bash
make build
```

#### 최적화된 빌드 (작은 바이너리)
```bash
CGO_ENABLED=0 go build -ldflags="-w -s" -o bin/protodiff ./cmd/protodiff
```

플래그 설명:
- `CGO_ENABLED=0`: 순수 Go 바이너리 (의존성 없음)
- `-w`: DWARF 디버그 정보 제거
- `-s`: 심볼 테이블 제거

### Docker 이미지 빌드

#### 로컬 빌드
```bash
docker build -t protodiff:latest .
```

#### 멀티 플랫폼 빌드
```bash
docker buildx build \
  --platform linux/amd64,linux/arm64 \
  -t protodiff:latest \
  .
```

### 테스트 실행

#### 모든 테스트
```bash
go test -v ./...
```

#### 커버리지 포함
```bash
go test -cover -coverprofile=coverage.out ./...
go tool cover -html=coverage.out
```

#### 특정 패키지
```bash
go test -v ./internal/core/store
go test -v ./internal/scanner
```

### 의존성 정보

#### 주요 의존성

```
직접 의존성:
- github.com/jhump/protoreflect v1.15.6  (gRPC Reflection)
- google.golang.org/grpc v1.62.0         (gRPC 클라이언트)
- k8s.io/api v0.29.0                     (Kubernetes API 타입)
- k8s.io/apimachinery v0.29.0            (Kubernetes 메타타입)
- k8s.io/client-go v0.29.0               (Kubernetes 클라이언트)

간접 의존성: 40+ 패키지
```

#### 의존성 관리
```bash
# 의존성 트리 확인
go mod graph | grep "github.com/uzdada/protodiff"

# 의존성 업데이트
go get -u ./...
go mod tidy
```

### 다음 단계

#### 1. 로컬 테스트
```bash
# Kubernetes 필요 (minikube, kind, Docker Desktop 등)

# 1. 클러스터 확인
kubectl cluster-info

# 2. 배포
kubectl apply -f deploy/k8s/install.yaml

# 3. 로그 확인
kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f

# 4. 대시보드 접속
kubectl port-forward -n protodiff-system svc/protodiff 8080:80
open http://localhost:8080
```

### 문제 해결

#### 빌드 오류

**증상**: `undefined: grpc.XXX`
```bash
# 해결: 의존성 재다운로드
go mod download
go mod tidy
go clean -modcache
```

**증상**: `go.sum: checksum mismatch`
```bash
# 해결: go.sum 재생성
rm go.sum
go mod download
go mod tidy
```

### 체크리스트

프로젝트 완성도 확인:

- [x] Go 코드 빌드 성공
- [x] 의존성 정리 완료 (go.sum)
- [x] Docker 이미지 빌드 가능
- [x] Kubernetes 매니페스트 작성
- [x] 문서 작성 완료
- [ ] 단위 테스트 작성
- [ ] 통합 테스트
- [ ] E2E 테스트
- [ ] CI/CD 파이프라인 구성
- [ ] Production 배포
