# ProtoDiff - Project Overview

[English](#english) | [한국어](#korean)

---

## English

### What is ProtoDiff?

ProtoDiff is a Kubernetes-native monitoring agent that detects gRPC schema drift between running services and the Buf Schema Registry (BSR).

### Project Statistics

- Total Files: 25+
- Go Code: 900+ lines
- Documentation: 7 files
- Architecture: Hexagonal Architecture
- Layout: Standard Go Project Layout

### Core Components

#### 1. Domain Layer (`internal/core/`)

**Models** (`domain/models.go`)
- `ScanResult`: Validation results for pods
- `DiffStatus`: SYNC, MISMATCH, UNKNOWN states
- `SchemaDescriptor`: Protocol buffer schema representation

**Storage** (`store/store.go`)
- Thread-safe in-memory storage using `sync.RWMutex`
- Stores scan results by namespace/podName key

#### 2. Adapters Layer (`internal/adapters/`)

**Kubernetes** (`k8s/client.go`)
- Pod discovery via label selector
- ConfigMap loading for service mappings

**gRPC** (`grpc/reflection.go`)
- gRPC server reflection client using `jhump/protoreflect`

**BSR** (`bsr/client.go` + `bsr/mock.go`)
- Interface-based BSR client
- Mock implementation for testing

**Web** (`web/server.go`)
- HTTP server with HTML dashboard
- Bootstrap-based UI
- Template embedding via `go:embed`

#### 3. Scanner (`internal/scanner/`)

Orchestrates the validation workflow:
1. Load ConfigMap mappings
2. Discover gRPC pods
3. Resolve BSR module names
4. Fetch live and truth schemas
5. Compare and detect drift
6. Update in-memory store

### Project Structure

```
protodiff/
├── cmd/protodiff/main.go          # Application entrypoint
├── internal/
│   ├── core/
│   │   ├── domain/                # Business models
│   │   └── store/                 # Thread-safe storage
│   ├── adapters/
│   │   ├── k8s/                   # Kubernetes client
│   │   ├── grpc/                  # gRPC reflection
│   │   ├── bsr/                   # BSR client
│   │   └── web/                   # HTTP server
│   └── scanner/                   # Validation orchestrator
├── deploy/k8s/install.yaml        # All-in-one manifest
├── Dockerfile                     # Multi-stage build
├── Makefile                       # Build automation
└── go.mod                         # Go dependencies
```

### Kubernetes Resources

The `deploy/k8s/install.yaml` contains:

1. **Namespace**: `protodiff-system`
2. **ServiceAccount**: Identity for pod
3. **ClusterRole**: Permissions (pods, configmaps: get/list/watch)
4. **ClusterRoleBinding**: Links ServiceAccount to ClusterRole
5. **ConfigMap**: Service-to-BSR mappings
6. **Deployment**: Application deployment (1 replica)
7. **Service**: ClusterIP on port 80

### Data Flow

```
Startup:
  Initialize Store → K8s Client → gRPC Client → BSR Client
  → Web Server (goroutine) → Scanner (goroutine)

Scan Cycle (every 30s):
  Load ConfigMap → Discover Pods → For each pod:
    Resolve BSR Module → Fetch Live Schema (gRPC Reflection)
    → Fetch Truth Schema (BSR) → Compare → Update Store

Dashboard Request:
  User Browser → kubectl port-forward → Service:80
  → Pod:8080 → Read Store → Render Template → HTML
```

### Security Features

**Container Security**:
- Non-root user (UID 65532)
- Read-only root filesystem
- No privilege escalation
- All capabilities dropped

**RBAC**:
- Minimal permissions (only get/list/watch)
- No write access to any resources
- Cluster-scoped for multi-namespace support

### Build Information

**Binary**:
- Size: 50MB (uncompressed)
- Platform: Cross-platform (linux/amd64, linux/arm64)
- Dependencies: Statically linked

**Dependencies**:
- github.com/jhump/protoreflect v1.15.6
- google.golang.org/grpc v1.62.0
- k8s.io/client-go v0.29.0

### Development Workflow

```bash
# Setup
make deps

# Build
make build

# Test
make test

# Docker
make docker-build

# Deploy
make deploy

# Access
make port-forward
```

### Configuration

**Environment Variables**:
- `CONFIGMAP_NAMESPACE`: ConfigMap namespace (default: protodiff-system)
- `CONFIGMAP_NAME`: ConfigMap name (default: protodiff-mapping)
- `DEFAULT_BSR_TEMPLATE`: Template for BSR module resolution
- `WEB_ADDR`: Web server address (default: :8080)
- `SCAN_INTERVAL`: Time between scans (default: 30s)

**ConfigMap Format**:
```yaml
data:
  service-name: "buf.build/org/module"
```

### Documentation

- **README.md**: Main documentation
- **QUICKSTART.md**: 5-minute setup guide
- **ARCHITECTURE.md**: Detailed architecture
- **CONTRIBUTING.md**: Contribution guidelines
- **BUILD_SUCCESS.md**: Build instructions
- **LICENSE**: Apache 2.0

### Future Enhancements

- Real BSR API client (currently mock)
- Persistent storage option (Redis/PostgreSQL)
- Detailed schema diff display
- Webhook notifications
- Prometheus metrics
- Multi-cluster support
- ConfigMap live-reload

---

## Korean

### ProtoDiff란?

ProtoDiff는 실행 중인 서비스와 Buf Schema Registry(BSR) 간의 gRPC 스키마 드리프트를 감지하는 Kubernetes 네이티브 모니터링 에이전트입니다.

### 프로젝트 통계

- 총 파일 수: 25+
- Go 코드: 900+ 줄
- 문서: 7개
- 아키텍처: 헥사고날 아키텍처
- 레이아웃: 표준 Go 프로젝트 레이아웃

### 핵심 컴포넌트

#### 1. 도메인 레이어 (`internal/core/`)

**모델** (`domain/models.go`)
- `ScanResult`: Pod에 대한 검증 결과
- `DiffStatus`: SYNC, MISMATCH, UNKNOWN 상태
- `SchemaDescriptor`: Protocol buffer 스키마 표현

**저장소** (`store/store.go`)
- `sync.RWMutex`를 사용한 Thread-safe 인메모리 저장소
- namespace/podName 키로 스캔 결과 저장

#### 2. 어댑터 레이어 (`internal/adapters/`)

**Kubernetes** (`k8s/client.go`)
- 레이블 셀렉터를 통한 Pod 발견
- 서비스 매핑을 위한 ConfigMap 로딩

**gRPC** (`grpc/reflection.go`)
- `jhump/protoreflect`를 사용한 gRPC 서버 리플렉션 클라이언트

**BSR** (`bsr/client.go` + `bsr/mock.go`)
- 인터페이스 기반 BSR 클라이언트
- 테스트용 Mock 구현

**Web** (`web/server.go`)
- HTML 대시보드가 있는 HTTP 서버
- Bootstrap 기반 UI
- `go:embed`를 통한 템플릿 임베딩

#### 3. 스캐너 (`internal/scanner/`)

검증 워크플로우 오케스트레이션:
1. ConfigMap 매핑 로드
2. gRPC Pod 발견
3. BSR 모듈 이름 해석
4. 라이브 및 진실 스키마 가져오기
5. 비교 및 드리프트 감지
6. 인메모리 저장소 업데이트

### 프로젝트 구조

```
protodiff/
├── cmd/protodiff/main.go          # 애플리케이션 엔트리포인트
├── internal/
│   ├── core/
│   │   ├── domain/                # 비즈니스 모델
│   │   └── store/                 # Thread-safe 저장소
│   ├── adapters/
│   │   ├── k8s/                   # Kubernetes 클라이언트
│   │   ├── grpc/                  # gRPC 리플렉션
│   │   ├── bsr/                   # BSR 클라이언트
│   │   └── web/                   # HTTP 서버
│   └── scanner/                   # 검증 오케스트레이터
├── deploy/k8s/install.yaml        # All-in-one 매니페스트
├── Dockerfile                     # 멀티 스테이지 빌드
├── Makefile                       # 빌드 자동화
└── go.mod                         # Go 의존성
```

### Kubernetes 리소스

`deploy/k8s/install.yaml`에 포함된 내용:

1. **Namespace**: `protodiff-system`
2. **ServiceAccount**: Pod의 ID
3. **ClusterRole**: 권한 (pods, configmaps: get/list/watch)
4. **ClusterRoleBinding**: ServiceAccount와 ClusterRole 연결
5. **ConfigMap**: 서비스-BSR 매핑
6. **Deployment**: 애플리케이션 배포 (1 replica)
7. **Service**: 포트 80의 ClusterIP

### 데이터 플로우

```
시작:
  Store 초기화 → K8s Client → gRPC Client → BSR Client
  → Web Server (고루틴) → Scanner (고루틴)

스캔 사이클 (30초마다):
  ConfigMap 로드 → Pod 발견 → 각 Pod에 대해:
    BSR 모듈 해석 → Live 스키마 가져오기 (gRPC Reflection)
    → Truth 스키마 가져오기 (BSR) → 비교 → Store 업데이트

대시보드 요청:
  사용자 브라우저 → kubectl port-forward → Service:80
  → Pod:8080 → Store 읽기 → 템플릿 렌더링 → HTML
```

### 보안 기능

**컨테이너 보안**:
- Non-root 사용자 (UID 65532)
- 읽기 전용 루트 파일시스템
- 권한 상승 불가
- 모든 capability 제거

**RBAC**:
- 최소 권한 (get/list/watch만)
- 모든 리소스에 대한 쓰기 권한 없음
- 멀티 네임스페이스 지원을 위한 클러스터 범위

### 빌드 정보

**바이너리**:
- 크기: 50MB (비압축)
- 플랫폼: 크로스 플랫폼 (linux/amd64, linux/arm64)
- 의존성: 정적 링크

**의존성**:
- github.com/jhump/protoreflect v1.15.6
- google.golang.org/grpc v1.62.0
- k8s.io/client-go v0.29.0

### 개발 워크플로우

```bash
# 설정
make deps

# 빌드
make build

# 테스트
make test

# Docker
make docker-build

# 배포
make deploy

# 접속
make port-forward
```

### 설정

**환경 변수**:
- `CONFIGMAP_NAMESPACE`: ConfigMap 네임스페이스 (기본값: protodiff-system)
- `CONFIGMAP_NAME`: ConfigMap 이름 (기본값: protodiff-mapping)
- `DEFAULT_BSR_TEMPLATE`: BSR 모듈 해석 템플릿
- `WEB_ADDR`: 웹 서버 주소 (기본값: :8080)
- `SCAN_INTERVAL`: 스캔 간격 (기본값: 30s)

**ConfigMap 형식**:
```yaml
data:
  service-name: "buf.build/org/module"
```

### 문서

- **README.md**: 메인 문서
- **QUICKSTART.md**: 5분 설정 가이드
- **ARCHITECTURE.md**: 상세 아키텍처
- **CONTRIBUTING.md**: 기여 가이드라인
- **BUILD_SUCCESS.md**: 빌드 지침
- **LICENSE**: Apache 2.0

### 향후 개선사항

- 실제 BSR API 클라이언트 (현재 mock)
- 영구 저장소 옵션 (Redis/PostgreSQL)
- 상세 스키마 차이 표시
- Webhook 알림
- Prometheus 메트릭
- 멀티 클러스터 지원
- ConfigMap 실시간 리로드
