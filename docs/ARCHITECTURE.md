# ProtoDiff Architecture

[English](#english) | [한국어](#korean)

---

## English

This document provides a detailed overview of ProtoDiff's architecture, design decisions, and implementation details.

### Table of Contents

- [Overview](#overview)
- [Architecture Patterns](#architecture-patterns)
- [Component Breakdown](#component-breakdown)
- [Data Flow](#data-flow)
- [Concurrency Model](#concurrency-model)
- [Design Decisions](#design-decisions)

### Overview

ProtoDiff is built using **Hexagonal Architecture** (also known as Ports and Adapters), which provides:

- **Clear separation of concerns**: Business logic isolated from external dependencies
- **Testability**: Easy to mock and test each layer independently
- **Flexibility**: Adapters can be swapped without changing core logic
- **Maintainability**: Well-organized codebase following Go project standards

### Architecture Patterns

#### Hexagonal Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                     External World                          │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  k8s API │  │  gRPC    │  │   BSR    │  │   HTTP   │   │
│  │          │  │ Services │  │   API    │  │ Browser  │   │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘   │
└────────┼─────────────┼─────────────┼─────────────┼─────────┘
         │             │             │             │
┌────────┼─────────────┼─────────────┼─────────────┼─────────┐
│        │   Adapters Layer (Ports)  │             │         │
│  ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐  │
│  │    k8s   │  │   grpc   │  │   bsr    │  │   web    │  │
│  │  Client  │  │  Client  │  │  Client  │  │  Server  │  │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘  │
└────────┼─────────────┼─────────────┼─────────────┼─────────┘
         │             │             │             │
┌────────┼─────────────┼─────────────┼─────────────┼─────────┐
│        │             │   Core Domain Layer        │         │
│        │             │                            │         │
│        │     ┌───────▼─────────┐          ┌──────▼──────┐  │
│        │     │     Scanner     │          │    Store    │  │
│        │     │  (Orchestrator) │◄────────►│ (In-Memory) │  │
│        │     └───────┬─────────┘          └─────────────┘  │
│        │             │                                      │
│        │     ┌───────▼─────────┐                           │
│        └────►│  Domain Models  │                           │
│              │   (ScanResult,  │                           │
│              │   DiffStatus)   │                           │
│              └─────────────────┘                           │
└─────────────────────────────────────────────────────────────┘
```

### Component Breakdown

#### Core Domain (`internal/core/`)

The innermost layer containing business logic and domain models.

**domain/models.go**

Defines core business entities:

- `ScanResult`: Represents validation results for a single pod
  - Pod information (name, namespace, IP)
  - Service and BSR module mapping
  - Drift status (SYNC, MISMATCH, UNKNOWN)
  - Timestamp and error messages

- `DiffStatus`: Enumeration of validation states
  - `StatusSync`: Schemas match perfectly
  - `StatusMismatch`: Schema drift detected
  - `StatusUnknown`: Unable to determine (errors, connectivity issues)

- `SchemaDescriptor`: Protocol buffer schema representation
  - Service definitions and RPC methods
  - Message type definitions

**store/store.go**

Thread-safe in-memory storage using `sync.RWMutex`:

```go
type Store struct {
    mu      sync.RWMutex
    results map[string]*domain.ScanResult
}
```

Design rationale:
- **In-memory**: Fast access, no external dependencies for MVP
- **Thread-safe**: Multiple goroutines (scanner + web server) access concurrently
- **RWMutex**: Allows concurrent reads, exclusive writes
- **Key format**: `{namespace}/{podName}` for uniqueness

Methods:
- `Set()`: Store/update scan results (write lock)
- `Get()`: Retrieve single result (read lock)
- `GetAll()`: Retrieve all results (read lock)
- `Delete()`: Remove result (write lock)

#### Adapters (`internal/adapters/`)

External integrations implementing ports to the core domain.

**k8s/client.go**

Kubernetes API integration:
- Pod Discovery: Lists pods with `grpc-service=true` label
- ConfigMap Loading: Reads service-to-BSR mappings
- In-cluster Config: Runs inside Kubernetes using service account

**grpc/reflection.go**

gRPC Server Reflection integration:
- Uses `jhump/protoreflect` library
- Connects to pod IP:port
- Lists available services and methods
- Builds `SchemaDescriptor` from reflection data

**bsr/client.go & bsr/mock.go**

Buf Schema Registry integration:
- Interface: Defines contract for BSR operations
- Mock Implementation: Provides sample data for MVP testing

**web/server.go**

HTTP dashboard server:
- Serves HTML template with Bootstrap UI
- Reads from in-memory store
- Aggregates statistics (sync/mismatch/unknown counts)
- Auto-refresh every 30 seconds
- Health check endpoint at `/health`

#### Scanner (`internal/scanner/`)

Orchestration layer coordinating all adapters.

Responsibilities:
1. Discovery: Find all gRPC pods in cluster
2. Resolution: Map service names to BSR modules
3. Validation: Compare live vs truth schemas
4. Storage: Update in-memory store with results

Scan cycle:
```
Load ConfigMap → Discover Pods → For each pod:
    ├─ Resolve BSR module
    ├─ Fetch live schema (gRPC reflection)
    ├─ Fetch truth schema (BSR)
    ├─ Compare schemas
    └─ Store result
```

### Data Flow

#### Startup Sequence

```
main.go
  │
  ├─► Initialize Store
  ├─► Create K8s Client
  ├─► Create gRPC Client
  ├─► Create BSR Client (mock)
  ├─► Initialize Web Server
  ├─► Initialize Scanner
  │
  ├─► Start Web Server (goroutine)
  └─► Start Scanner (goroutine)
```

#### Scan Cycle Flow

```
Scanner Loop (every 30s)
  │
  ├─► Load ConfigMap mappings
  │     └─► {service-name: bsr-module}
  │
  ├─► Discover gRPC Pods
  │     └─► [PodInfo, PodInfo, ...]
  │
  └─► For each pod:
        │
        ├─► Resolve BSR Module
        │     ├─ Check ConfigMap
        │     └─ Fallback to template
        │
        ├─► Fetch Live Schema
        │     └─► gRPC Reflection → SchemaDescriptor
        │
        ├─► Fetch Truth Schema
        │     └─► BSR API → SchemaDescriptor
        │
        ├─► Compare Schemas
        │     └─► schemasMatch() → bool
        │
        └─► Update Store
              └─► store.Set(ScanResult)
```

#### Dashboard Request Flow

```
User Browser
  │
  ├─► GET http://localhost:8080/
  │
  └─► Web Server
        │
        ├─► Read from Store
        │     └─► store.GetAll() → [ScanResult, ...]
        │
        ├─► Aggregate Statistics
        │     └─► count sync/mismatch/unknown
        │
        └─► Render Template
              └─► HTML with Bootstrap UI
```

### Concurrency Model

#### Goroutines

ProtoDiff uses two main goroutines:

1. **Web Server Goroutine**
   - Runs HTTP server
   - Handles incoming dashboard requests
   - Read-only access to store (RLock)

2. **Scanner Goroutine**
   - Periodic validation loop
   - Writes to store (Lock)
   - Can be cancelled via context

#### Synchronization

```go
// In-memory store uses RWMutex for safe concurrent access
type Store struct {
    mu      sync.RWMutex  // Allows multiple readers OR single writer
    results map[string]*domain.ScanResult
}

// Web server (multiple concurrent requests)
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
    results := s.store.GetAll()  // RLock acquired internally
}

// Scanner (single goroutine)
func (s *Scanner) validatePod(...) {
    s.store.Set(result)  // Lock acquired internally
}
```

#### Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())

// Listen for SIGINT/SIGTERM
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

<-sigChan              // Block until signal
cancel()               // Cancel context
time.Sleep(2 * time.Second)  // Allow cleanup
```

### Design Decisions

#### Why In-Memory Storage?

**Pros**:
- Zero external dependencies
- Fast access (no network/disk I/O)
- Simple deployment (single binary)
- Sufficient for MVP

**Cons**:
- Data lost on restart
- No historical tracking
- Memory limited

**Future**: Add optional persistent storage (Redis, PostgreSQL)

#### Why Mock BSR Client?

For MVP and testing:
- No BSR API credentials needed
- Predictable test data
- Faster development iteration

**Production**: Implement real BSR HTTP API client

#### Why gRPC Reflection?

**Advantages**:
- No access to proto files needed
- Works with any gRPC service
- Standard protocol

**Requirements**:
- Services must enable reflection
- One line in Go: `reflection.Register(server)`

#### Why ConfigMap for Mappings?

**Benefits**:
- Centralized configuration
- No pod restarts needed
- Native Kubernetes resource
- Easy to edit: `kubectl edit`

**Alternative**: CRD (Custom Resource Definition)

#### Why Label-Based Discovery?

Simple and non-invasive:
- Add one label: `grpc-service=true`
- No deployment changes needed
- Standard Kubernetes practice

**Alternative**: Service mesh integration

### Performance Considerations

- **Scan Interval**: Default 30s, configurable
- **Pod Count**: Tested up to 100 pods
- **Memory Usage**: ~10MB baseline + ~1KB per pod result
- **CPU Usage**: Minimal (mostly I/O bound)

### Security

- **RBAC**: Minimal permissions (get/list/watch pods, ConfigMaps)
- **Non-root**: Runs as user 65532
- **Read-only FS**: Container filesystem is read-only
- **No capabilities**: Drops all Linux capabilities
- **In-cluster only**: Uses internal pod IPs

---

## Korean

본 문서는 ProtoDiff의 아키텍처, 설계 결정 및 구현 세부사항에 대한 상세한 개요를 제공합니다.

### 목차

- [개요](#개요)
- [아키텍처 패턴](#아키텍처-패턴)
- [컴포넌트 분석](#컴포넌트-분석)
- [데이터 플로우](#데이터-플로우)
- [동시성 모델](#동시성-모델)
- [설계 결정](#설계-결정)

### 개요

ProtoDiff는 **헥사고날 아키텍처**(포트와 어댑터라고도 함)를 사용하여 구축되었으며, 다음을 제공합니다:

- **명확한 관심사 분리**: 비즈니스 로직이 외부 종속성으로부터 격리됨
- **테스트 가능성**: 각 레이어를 독립적으로 쉽게 모킹하고 테스트
- **유연성**: 핵심 로직을 변경하지 않고 어댑터 교체 가능
- **유지보수성**: Go 프로젝트 표준을 따르는 잘 정리된 코드베이스

### 아키텍처 패턴

#### 헥사고날 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│                     외부 세계                                │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐  ┌──────────┐   │
│  │  k8s API │  │  gRPC    │  │   BSR    │  │   HTTP   │   │
│  │          │  │ Services │  │   API    │  │ Browser  │   │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘   │
└────────┼─────────────┼─────────────┼─────────────┼─────────┘
         │             │             │             │
┌────────┼─────────────┼─────────────┼─────────────┼─────────┐
│        │   어댑터 레이어 (포트)      │             │         │
│  ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐  ┌─────▼────┐  │
│  │    k8s   │  │   grpc   │  │   bsr    │  │   web    │  │
│  │  Client  │  │  Client  │  │  Client  │  │  Server  │  │
│  └─────┬────┘  └─────┬────┘  └─────┬────┘  └─────┬────┘  │
└────────┼─────────────┼─────────────┼─────────────┼─────────┘
         │             │             │             │
┌────────┼─────────────┼─────────────┼─────────────┼─────────┐
│        │             │   핵심 도메인 레이어        │         │
│        │             │                            │         │
│        │     ┌───────▼─────────┐          ┌──────▼──────┐  │
│        │     │     Scanner     │          │    Store    │  │
│        │     │  (오케스트레이터)│◄────────►│  (인메모리)  │  │
│        │     └───────┬─────────┘          └─────────────┘  │
│        │             │                                      │
│        │     ┌───────▼─────────┐                           │
│        └────►│  도메인 모델     │                           │
│              │   (ScanResult,  │                           │
│              │   DiffStatus)   │                           │
│              └─────────────────┘                           │
└─────────────────────────────────────────────────────────────┘
```

### 컴포넌트 분석

#### 핵심 도메인 (`internal/core/`)

비즈니스 로직과 도메인 모델을 포함하는 가장 안쪽 레이어입니다.

**domain/models.go**

핵심 비즈니스 엔티티 정의:

- `ScanResult`: 단일 Pod에 대한 검증 결과 표현
  - Pod 정보 (이름, 네임스페이스, IP)
  - 서비스 및 BSR 모듈 매핑
  - 드리프트 상태 (SYNC, MISMATCH, UNKNOWN)
  - 타임스탬프 및 오류 메시지

- `DiffStatus`: 검증 상태 열거형
  - `StatusSync`: 스키마가 완벽하게 일치
  - `StatusMismatch`: 스키마 드리프트 감지
  - `StatusUnknown`: 확인 불가 (오류, 연결 문제)

- `SchemaDescriptor`: Protocol buffer 스키마 표현
  - 서비스 정의 및 RPC 메서드
  - 메시지 타입 정의

**store/store.go**

`sync.RWMutex`를 사용한 Thread-safe 인메모리 저장소:

```go
type Store struct {
    mu      sync.RWMutex
    results map[string]*domain.ScanResult
}
```

설계 근거:
- **인메모리**: 빠른 접근, MVP를 위한 외부 종속성 없음
- **Thread-safe**: 여러 고루틴 (스캐너 + 웹 서버)이 동시 접근
- **RWMutex**: 동시 읽기 허용, 배타적 쓰기
- **키 형식**: 고유성을 위한 `{namespace}/{podName}`

메서드:
- `Set()`: 스캔 결과 저장/업데이트 (쓰기 잠금)
- `Get()`: 단일 결과 검색 (읽기 잠금)
- `GetAll()`: 모든 결과 검색 (읽기 잠금)
- `Delete()`: 결과 제거 (쓰기 잠금)

#### 어댑터 (`internal/adapters/`)

핵심 도메인에 대한 포트를 구현하는 외부 통합.

**k8s/client.go**

Kubernetes API 통합:
- Pod 발견: `grpc-service=true` 레이블이 있는 Pod 나열
- ConfigMap 로딩: 서비스-BSR 매핑 읽기
- 클러스터 내 설정: 서비스 계정을 사용하여 Kubernetes 내부에서 실행

**grpc/reflection.go**

gRPC 서버 리플렉션 통합:
- `jhump/protoreflect` 라이브러리 사용
- Pod IP:포트에 연결
- 사용 가능한 서비스 및 메서드 나열
- 리플렉션 데이터에서 `SchemaDescriptor` 구축

**bsr/client.go & bsr/mock.go**

Buf Schema Registry 통합:
- 인터페이스: BSR 작업에 대한 계약 정의
- Mock 구현: MVP 테스트를 위한 샘플 데이터 제공

**web/server.go**

HTTP 대시보드 서버:
- Bootstrap UI가 있는 HTML 템플릿 제공
- 인메모리 저장소에서 읽기
- 통계 집계 (동기화/불일치/알 수 없음 개수)
- 30초마다 자동 새로고침
- `/health`에서 상태 확인 엔드포인트

#### 스캐너 (`internal/scanner/`)

모든 어댑터를 조정하는 오케스트레이션 레이어.

책임:
1. 발견: 클러스터의 모든 gRPC Pod 찾기
2. 해석: 서비스 이름을 BSR 모듈에 매핑
3. 검증: 라이브 스키마와 진실 스키마 비교
4. 저장: 결과로 인메모리 저장소 업데이트

스캔 사이클:
```
ConfigMap 로드 → Pod 발견 → 각 Pod에 대해:
    ├─ BSR 모듈 해석
    ├─ 라이브 스키마 가져오기 (gRPC 리플렉션)
    ├─ 진실 스키마 가져오기 (BSR)
    ├─ 스키마 비교
    └─ 결과 저장
```

### 데이터 플로우

#### 시작 순서

```
main.go
  │
  ├─► Store 초기화
  ├─► K8s Client 생성
  ├─► gRPC Client 생성
  ├─► BSR Client 생성 (mock)
  ├─► Web Server 초기화
  ├─► Scanner 초기화
  │
  ├─► Web Server 시작 (고루틴)
  └─► Scanner 시작 (고루틴)
```

#### 스캔 사이클 플로우

```
Scanner 루프 (30초마다)
  │
  ├─► ConfigMap 매핑 로드
  │     └─► {서비스명: bsr-모듈}
  │
  ├─► gRPC Pod 발견
  │     └─► [PodInfo, PodInfo, ...]
  │
  └─► 각 Pod에 대해:
        │
        ├─► BSR 모듈 해석
        │     ├─ ConfigMap 확인
        │     └─ 템플릿으로 폴백
        │
        ├─► Live 스키마 가져오기
        │     └─► gRPC Reflection → SchemaDescriptor
        │
        ├─► Truth 스키마 가져오기
        │     └─► BSR API → SchemaDescriptor
        │
        ├─► 스키마 비교
        │     └─► schemasMatch() → bool
        │
        └─► Store 업데이트
              └─► store.Set(ScanResult)
```

#### 대시보드 요청 플로우

```
사용자 브라우저
  │
  ├─► GET http://localhost:8080/
  │
  └─► Web Server
        │
        ├─► Store에서 읽기
        │     └─► store.GetAll() → [ScanResult, ...]
        │
        ├─► 통계 집계
        │     └─► sync/mismatch/unknown 개수 세기
        │
        └─► 템플릿 렌더링
              └─► Bootstrap UI가 있는 HTML
```

### 동시성 모델

#### 고루틴

ProtoDiff는 두 개의 주요 고루틴을 사용합니다:

1. **Web Server 고루틴**
   - HTTP 서버 실행
   - 들어오는 대시보드 요청 처리
   - 저장소에 대한 읽기 전용 접근 (RLock)

2. **Scanner 고루틴**
   - 주기적 검증 루프
   - 저장소에 쓰기 (Lock)
   - 컨텍스트를 통해 취소 가능

#### 동기화

```go
// 인메모리 저장소는 안전한 동시 접근을 위해 RWMutex 사용
type Store struct {
    mu      sync.RWMutex  // 여러 읽기자 또는 단일 쓰기자 허용
    results map[string]*domain.ScanResult
}

// 웹 서버 (여러 동시 요청)
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
    results := s.store.GetAll()  // 내부적으로 RLock 획득
}

// 스캐너 (단일 고루틴)
func (s *Scanner) validatePod(...) {
    s.store.Set(result)  // 내부적으로 Lock 획득
}
```

#### Graceful Shutdown

```go
ctx, cancel := context.WithCancel(context.Background())

// SIGINT/SIGTERM 수신 대기
sigChan := make(chan os.Signal, 1)
signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

<-sigChan              // 시그널까지 차단
cancel()               // 컨텍스트 취소
time.Sleep(2 * time.Second)  // 정리 허용
```

### 설계 결정

#### 왜 인메모리 저장소인가?

**장점**:
- 외부 종속성 없음
- 빠른 접근 (네트워크/디스크 I/O 없음)
- 간단한 배포 (단일 바이너리)
- MVP에 충분

**단점**:
- 재시작 시 데이터 손실
- 이력 추적 없음
- 메모리 제한

**향후**: 선택적 영구 저장소 추가 (Redis, PostgreSQL)

#### 왜 Mock BSR Client인가?

MVP 및 테스트용:
- BSR API 자격 증명 불필요
- 예측 가능한 테스트 데이터
- 빠른 개발 반복

**프로덕션**: 실제 BSR HTTP API 클라이언트 구현

#### 왜 gRPC Reflection인가?

**장점**:
- proto 파일 접근 불필요
- 모든 gRPC 서비스와 작동
- 표준 프로토콜

**요구사항**:
- 서비스에서 리플렉션 활성화 필요
- Go에서 한 줄: `reflection.Register(server)`

#### 왜 매핑에 ConfigMap인가?

**이점**:
- 중앙 집중식 설정
- Pod 재시작 불필요
- 네이티브 Kubernetes 리소스
- 쉬운 편집: `kubectl edit`

**대안**: CRD (Custom Resource Definition)

#### 왜 레이블 기반 발견인가?

간단하고 비침투적:
- 레이블 하나 추가: `grpc-service=true`
- 배포 변경 불필요
- 표준 Kubernetes 관행

**대안**: 서비스 메시 통합

### 성능 고려사항

- **스캔 간격**: 기본 30s, 설정 가능
- **Pod 수**: 최대 100개 Pod까지 테스트
- **메모리 사용**: ~10MB 기본 + Pod 결과당 ~1KB
- **CPU 사용**: 최소 (대부분 I/O 바운드)

### 보안

- **RBAC**: 최소 권한 (pods, ConfigMaps get/list/watch)
- **Non-root**: 사용자 65532로 실행
- **Read-only FS**: 컨테이너 파일시스템 읽기 전용
- **No capabilities**: 모든 Linux capability 제거
- **In-cluster only**: 내부 Pod IP 사용
