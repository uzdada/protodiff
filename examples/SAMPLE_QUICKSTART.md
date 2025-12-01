# gRPC Test Services Quick Start Guide

[English](#english) | [한국어](#korean)

---

## English

### Overview

This example demonstrates **two gRPC services communicating with each other** in a Kubernetes cluster:

- **Go Greeter Service** (Port 9090): A greeting service that provides personalized greetings
- **Java UserService** (Port 9091): A user management service with in-memory storage

**Communication Flow:**
```
Client → Go Greeter Service → Java UserService
         (SayHelloToUser)      (GetUser)
```

When you call `SayHelloToUser(user_id)` on the Go service, it:
1. Receives the user ID from the client
2. Calls Java UserService's `GetUser(user_id)` to fetch user details
3. Creates a personalized greeting with the user's information
4. Returns the greeting to the client

This setup is designed to work with **ProtoDiff** for monitoring schema drift between your gRPC services and the Buf Schema Registry.

### Architecture

```
┌─────────────────────────────────────────────────────────────┐
│  Kubernetes Cluster (namespace: grpc-test)                  │
│                                                              │
│  ┌──────────────────────────┐    ┌─────────────────────┐   │
│  │  grpc-server-go:9090     │───▶│ grpc-server-java    │   │
│  │  (Go Greeter Service)    │    │ :9091               │   │
│  │                          │    │ (Java UserService)  │   │
│  │  Services:               │    │                     │   │
│  │  - SayHello              │    │ Services:           │   │
│  │  - SayHelloAgain         │    │ - GetUser           │   │
│  │  - SayHelloToUser ───────┼───▶│ - CreateUser        │   │
│  │    (calls Java service)  │    │ - ListUsers         │   │
│  │                          │    │                     │   │
│  │  Label: grpc-service=true│    │ Label: grpc-service │   │
│  └──────────────────────────┘    │ =true               │   │
│                                   └─────────────────────┘   │
│                                                              │
│  Both services have gRPC Reflection enabled                 │
│  (required for ProtoDiff to discover schemas)               │
└─────────────────────────────────────────────────────────────┘

        ▲
        │
        │ Monitors schema drift
        │
┌───────┴────────────────────┐
│  ProtoDiff                 │
│  (protodiff-system ns)     │
│                            │
│  Dashboard: :18080         │
└────────────────────────────┘
```

### Prerequisites

- Kubernetes cluster (minikube, kind, or cloud provider)
- `kubectl` configured to access your cluster
- `grpcurl` for testing (optional but recommended)

**Install grpcurl:**
```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# Or download from: https://github.com/fullstorydev/grpcurl/releases
```

### Quick Start

#### 1. Deploy the Services

```bash
# Navigate to the examples directory
cd examples

# Apply the Kubernetes manifests
kubectl apply -f sample-grpc-service.yaml
```

This creates:
- Namespace `grpc-test`
- Two Deployments (grpc-server-go and grpc-server-java)
- Two Services (ClusterIP)

#### 2. Verify Deployment

```bash
# Check if pods are running
kubectl get pods -n grpc-test

# Expected output:
# NAME                                READY   STATUS    RESTARTS   AGE
# grpc-server-go-xxxxxxxxxx-xxxxx     1/1     Running   0          30s
# grpc-server-java-xxxxxxxxxx-xxxxx   1/1     Running   0          30s

# Check services
kubectl get svc -n grpc-test

# Check pods with grpc-service label (these are discovered by ProtoDiff)
kubectl get pods -n grpc-test -l grpc-service=true
```

#### 3. Test the Services

**Option A: Test Go Greeter Service (Standalone)**

```bash
# Port-forward Go service
kubectl port-forward -n grpc-test svc/grpc-server-go 9090:9090

# In another terminal, list available services
grpcurl -plaintext localhost:9090 list

# Call SayHello
grpcurl -plaintext -d '{"name": "World"}' localhost:9090 greeter.Greeter/SayHello

# Expected response:
# {
#   "message": "Hello World from Go server!"
# }

# Call SayHelloAgain
grpcurl -plaintext -d '{"name": "Alice"}' localhost:9090 greeter.Greeter/SayHelloAgain

# Expected response:
# {
#   "message": "Hello again Alice from Go server!"
# }
```

**Option B: Test Java UserService (Standalone)**

```bash
# Port-forward Java service
kubectl port-forward -n grpc-test svc/grpc-server-java 9091:9091

# In another terminal, list available services
grpcurl -plaintext localhost:9091 list

# Get user by ID (sample users: 1=admin, 2=user1, 3=user2)
grpcurl -plaintext -d '{"user_id": 1}' localhost:9091 user.UserService/GetUser

# Expected response:
# {
#   "userId": 1,
#   "username": "admin",
#   "email": "admin@example.com",
#   "createdAt": "1733024832123"
# }

# Create a new user
grpcurl -plaintext -d '{"username": "john", "email": "john@example.com"}' \
  localhost:9091 user.UserService/CreateUser

# List all users
grpcurl -plaintext -d '{"page_size": 10, "page_number": 1}' \
  localhost:9091 user.UserService/ListUsers
```

**Option C: Test Inter-Service Communication** ⭐

This is the main feature - Go service calling Java service!

```bash
# Port-forward Go service
kubectl port-forward -n grpc-test svc/grpc-server-go 9090:9090

# In another terminal, call SayHelloToUser
# This will make Go service call Java service internally
grpcurl -plaintext -d '{"user_id": 1}' localhost:9090 greeter.Greeter/SayHelloToUser

# Expected response (personalized greeting with user data from Java service):
# {
#   "message": "Hello admin (ID: 1, Email: admin@example.com)! Greetings from Go Greeter Service!"
# }

# Try with different user IDs
grpcurl -plaintext -d '{"user_id": 2}' localhost:9090 greeter.Greeter/SayHelloToUser
grpcurl -plaintext -d '{"user_id": 3}' localhost:9090 greeter.Greeter/SayHelloToUser

# Try with non-existent user (will return error)
grpcurl -plaintext -d '{"user_id": 999}' localhost:9090 greeter.Greeter/SayHelloToUser
```

#### 4. View Service Logs

```bash
# Go service logs
kubectl logs -n grpc-test -l app=grpc-server-go -f

# You'll see logs like:
# Go gRPC server listening at [::]:9090
# Received SayHelloToUser request: user_id=1
# Successfully greeted user: admin

# Java service logs
kubectl logs -n grpc-test -l app=grpc-server-java -f

# You'll see logs like:
# Java gRPC server started, listening on port 9091
# GetUser called for userId: 1
```

### Integration with ProtoDiff

These test services are designed to work seamlessly with ProtoDiff for schema monitoring. The schemas are already published to a **public BSR repository** at `buf.build/proto-diff-bsr/test-services`, so you can test ProtoDiff without setting up your own BSR account.

#### 1. Deploy ProtoDiff

Download the installation manifest:

```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
```

Edit the ConfigMap section to configure the test services:

```bash
vi install.yaml  # or use your preferred editor
```

Find the ConfigMap section (around line 69-71) and add:

```yaml
data:
  grpc-server-go: "buf.build/proto-diff-bsr/test-services"
  grpc-server-java: "buf.build/proto-diff-bsr/test-services"
```

Deploy ProtoDiff:

```bash
kubectl apply -f install.yaml
```

Verify deployment:

```bash
kubectl get pods -n protodiff-system
# Expected: protodiff pod running
```

**Note**: The schemas are already published at https://buf.build/proto-diff-bsr/test-services - you don't need to push anything!

#### 2. Verify ProtoDiff Discovery

Check that ProtoDiff discovered your test services:

```bash
# Check ProtoDiff logs
kubectl logs -n protodiff-system -l app=protodiff -f

# You should see logs like:
# Discovered gRPC service: grpc-server-go in namespace grpc-test
# Discovered gRPC service: grpc-server-java in namespace grpc-test
# Fetching schema for grpc-server-go...
# Comparing with BSR module: buf.build/proto-diff-bsr/test-services
```

#### 3. Access ProtoDiff Dashboard

```bash
kubectl port-forward -n protodiff-system svc/protodiff 18080:80
```

Open http://localhost:18080 in your browser. You should see:

- **grpc-server-go**
  - Status: 🟢 Green (if schema matches BSR)
  - BSR Module: `buf.build/proto-diff-bsr/test-services`
  - Services: `greeter.Greeter`

- **grpc-server-java**
  - Status: 🟢 Green (if schema matches BSR)
  - BSR Module: `buf.build/proto-diff-bsr/test-services`
  - Services: `user.UserService`

**Status Meanings:**
- 🟢 **Green (IN_SYNC)**: Live schema matches BSR - all good!
- 🔴 **Red (MISMATCH)**: Schema drift detected - update needed
- 🟡 **Yellow (UNKNOWN)**: Can't fetch schema or BSR module not found

#### 4. Understanding the Dashboard

The dashboard shows the current status of schema synchronization. For these test services, you should see:

- 🟢 **Green (IN_SYNC)**: The deployed service schemas match the BSR schemas
- Both services pointing to the same BSR module: `buf.build/proto-diff-bsr/test-services`

**What ProtoDiff is Checking:**

ProtoDiff continuously monitors your deployed gRPC services by:
1. Using gRPC reflection to fetch the live schemas from your running pods
2. Comparing them against the schemas stored in BSR
3. Alerting you when they drift apart

This ensures your documentation (BSR) stays synchronized with your actual deployments!

### Cleanup

```bash
# Delete the test services
kubectl delete -f sample-grpc-service.yaml

# This removes:
# - grpc-test namespace
# - All deployments, services, and pods
```

### Troubleshooting

#### Pods Not Starting

```bash
# Check pod events
kubectl describe pod -n grpc-test <pod-name>

# Common issues:
# - ImagePullBackOff: Check if images are accessible from Docker Hub
# - CrashLoopBackOff: Check logs with kubectl logs
# - ARM64/AMD64 compatibility: Images are now built for both architectures
```

#### Health Check Failures

If you see health check errors like "nc: not found", the images use tcpSocket probes instead of exec commands with netcat.

#### Connection Refused Between Services

```bash
# Verify service DNS resolution
kubectl run -it --rm debug --image=busybox --restart=Never -n grpc-test -- sh

# Inside the pod:
nslookup grpc-server-java.grpc-test.svc.cluster.local
nslookup grpc-server-go.grpc-test.svc.cluster.local

# Test connectivity
nc -zv grpc-server-java.grpc-test.svc.cluster.local 9091
```

#### gRPC Call Failures

```bash
# Check if gRPC reflection is enabled
grpcurl -plaintext localhost:9090 list

# If you see "Failed to list services", reflection might not be enabled
# Check the server logs for errors
```

### Next Steps

- **Monitor Schema Drift**: Use ProtoDiff to detect when your deployed services diverge from BSR
- **Add More Services**: Create additional gRPC services following the same pattern
- **Customize Protos**: Modify the proto definitions and redeploy to see ProtoDiff detect changes
- **Production Deployment**: Adapt these examples for your production environment

### Resources

- **Main Documentation**: [../README.md](../README.md)
- **Go Server Source**: See `grpc-server-go/` directory in parent folder
- **Java Server Source**: See `grpc-server-java/` directory in parent folder
- **ProtoDiff GitHub**: https://github.com/uzdada/protodiff
- **Buf Schema Registry**: https://buf.build

---

## Korean

### 개요

이 예제는 **Kubernetes 클러스터에서 서로 통신하는 두 개의 gRPC 서비스**를 보여줍니다:

- **Go Greeter Service** (포트 9090): 개인화된 인사말을 제공하는 서비스
- **Java UserService** (포트 9091): 인메모리 저장소를 사용하는 사용자 관리 서비스

**통신 흐름:**
```
클라이언트 → Go Greeter Service → Java UserService
            (SayHelloToUser)      (GetUser)
```

Go 서비스에서 `SayHelloToUser(user_id)`를 호출하면:
1. 클라이언트로부터 사용자 ID를 받습니다
2. Java UserService의 `GetUser(user_id)`를 호출하여 사용자 정보를 가져옵니다
3. 사용자 정보를 포함한 개인화된 인사말을 생성합니다
4. 클라이언트에게 인사말을 반환합니다

이 구성은 gRPC 서비스와 Buf Schema Registry 간의 스키마 드리프트를 모니터링하기 위한 **ProtoDiff**와 함께 작동하도록 설계되었습니다.

### 아키텍처

```
┌─────────────────────────────────────────────────────────────┐
│  Kubernetes 클러스터 (네임스페이스: grpc-test)               │
│                                                              │
│  ┌──────────────────────────┐    ┌─────────────────────┐   │
│  │  grpc-server-go:9090     │───▶│ grpc-server-java    │   │
│  │  (Go Greeter Service)    │    │ :9091               │   │
│  │                          │    │ (Java UserService)  │   │
│  │  서비스:                  │    │                     │   │
│  │  - SayHello              │    │ 서비스:              │   │
│  │  - SayHelloAgain         │    │ - GetUser           │   │
│  │  - SayHelloToUser ───────┼───▶│ - CreateUser        │   │
│  │    (Java 서비스 호출)     │    │ - ListUsers         │   │
│  │                          │    │                     │   │
│  │  레이블: grpc-service=true│   │ 레이블: grpc-service│   │
│  └──────────────────────────┘    │ =true               │   │
│                                   └─────────────────────┘   │
│                                                              │
│  두 서비스 모두 gRPC Reflection 활성화                       │
│  (ProtoDiff가 스키마를 발견하는 데 필요)                     │
└─────────────────────────────────────────────────────────────┘

        ▲
        │
        │ 스키마 드리프트 모니터링
        │
┌───────┴────────────────────┐
│  ProtoDiff                 │
│  (protodiff-system ns)     │
│                            │
│  대시보드: :18080           │
└────────────────────────────┘
```

### 사전 요구사항

- Kubernetes 클러스터 (minikube, kind, 또는 클라우드 제공자)
- `kubectl` 클러스터 접근 설정 완료
- `grpcurl` 테스트용 (선택사항이지만 권장)

**grpcurl 설치:**
```bash
# macOS
brew install grpcurl

# Linux
go install github.com/fullstorydev/grpcurl/cmd/grpcurl@latest

# 또는 다운로드: https://github.com/fullstorydev/grpcurl/releases
```

### 빠른 시작

#### 1. 서비스 배포

```bash
# examples 디렉토리로 이동
cd examples

# Kubernetes 매니페스트 적용
kubectl apply -f sample-grpc-service.yaml
```

다음이 생성됩니다:
- 네임스페이스 `grpc-test`
- 두 개의 Deployment (grpc-server-go와 grpc-server-java)
- 두 개의 Service (ClusterIP)

#### 2. 배포 확인

```bash
# Pod 실행 상태 확인
kubectl get pods -n grpc-test

# 예상 출력:
# NAME                                READY   STATUS    RESTARTS   AGE
# grpc-server-go-xxxxxxxxxx-xxxxx     1/1     Running   0          30s
# grpc-server-java-xxxxxxxxxx-xxxxx   1/1     Running   0          30s

# 서비스 확인
kubectl get svc -n grpc-test

# grpc-service 레이블이 있는 Pod 확인 (ProtoDiff가 발견하는 대상)
kubectl get pods -n grpc-test -l grpc-service=true
```

#### 3. 서비스 테스트

**옵션 A: Go Greeter Service 테스트 (단독)**

```bash
# Go 서비스 포트 포워딩
kubectl port-forward -n grpc-test svc/grpc-server-go 9090:9090

# 다른 터미널에서 사용 가능한 서비스 목록 확인
grpcurl -plaintext localhost:9090 list

# SayHello 호출
grpcurl -plaintext -d '{"name": "World"}' localhost:9090 greeter.Greeter/SayHello

# 예상 응답:
# {
#   "message": "Hello World from Go server!"
# }

# SayHelloAgain 호출
grpcurl -plaintext -d '{"name": "Alice"}' localhost:9090 greeter.Greeter/SayHelloAgain

# 예상 응답:
# {
#   "message": "Hello again Alice from Go server!"
# }
```

**옵션 B: Java UserService 테스트 (단독)**

```bash
# Java 서비스 포트 포워딩
kubectl port-forward -n grpc-test svc/grpc-server-java 9091:9091

# 다른 터미널에서 사용 가능한 서비스 목록 확인
grpcurl -plaintext localhost:9091 list

# ID로 사용자 조회 (샘플 사용자: 1=admin, 2=user1, 3=user2)
grpcurl -plaintext -d '{"user_id": 1}' localhost:9091 user.UserService/GetUser

# 예상 응답:
# {
#   "userId": 1,
#   "username": "admin",
#   "email": "admin@example.com",
#   "createdAt": "1733024832123"
# }

# 새 사용자 생성
grpcurl -plaintext -d '{"username": "john", "email": "john@example.com"}' \
  localhost:9091 user.UserService/CreateUser

# 모든 사용자 목록 조회
grpcurl -plaintext -d '{"page_size": 10, "page_number": 1}' \
  localhost:9091 user.UserService/ListUsers
```

**옵션 C: 서비스 간 통신 테스트** ⭐

이것이 핵심 기능입니다 - Go 서비스가 Java 서비스를 호출합니다!

```bash
# Go 서비스 포트 포워딩
kubectl port-forward -n grpc-test svc/grpc-server-go 9090:9090

# 다른 터미널에서 SayHelloToUser 호출
# Go 서비스가 내부적으로 Java 서비스를 호출합니다
grpcurl -plaintext -d '{"user_id": 1}' localhost:9090 greeter.Greeter/SayHelloToUser

# 예상 응답 (Java 서비스에서 가져온 사용자 데이터로 개인화된 인사말):
# {
#   "message": "Hello admin (ID: 1, Email: admin@example.com)! Greetings from Go Greeter Service!"
# }

# 다른 사용자 ID로 시도
grpcurl -plaintext -d '{"user_id": 2}' localhost:9090 greeter.Greeter/SayHelloToUser
grpcurl -plaintext -d '{"user_id": 3}' localhost:9090 greeter.Greeter/SayHelloToUser

# 존재하지 않는 사용자로 시도 (오류 반환)
grpcurl -plaintext -d '{"user_id": 999}' localhost:9090 greeter.Greeter/SayHelloToUser
```

#### 4. 서비스 로그 확인

```bash
# Go 서비스 로그
kubectl logs -n grpc-test -l app=grpc-server-go -f

# 다음과 같은 로그가 표시됩니다:
# Go gRPC server listening at [::]:9090
# Received SayHelloToUser request: user_id=1
# Successfully greeted user: admin

# Java 서비스 로그
kubectl logs -n grpc-test -l app=grpc-server-java -f

# 다음과 같은 로그가 표시됩니다:
# Java gRPC server started, listening on port 9091
# GetUser called for userId: 1
```

### ProtoDiff와 통합

이 테스트 서비스는 스키마 모니터링을 위해 ProtoDiff와 원활하게 작동하도록 설계되었습니다. 스키마는 이미 `buf.build/proto-diff-bsr/test-services`의 **퍼블릭 BSR 리포지토리**에 게시되어 있으므로, 별도의 BSR 계정 설정 없이 ProtoDiff를 테스트할 수 있습니다.

#### 1. ProtoDiff 배포

설치 매니페스트 다운로드:

```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
```

ConfigMap 섹션을 편집하여 테스트 서비스 설정:

```bash
vi install.yaml  # 또는 원하는 에디터 사용
```

ConfigMap 섹션(69-71번째 줄 근처)을 찾아 추가:

```yaml
data:
  grpc-server-go: "buf.build/proto-diff-bsr/test-services"
  grpc-server-java: "buf.build/proto-diff-bsr/test-services"
```

ProtoDiff 배포:

```bash
kubectl apply -f install.yaml
```

배포 확인:

```bash
kubectl get pods -n protodiff-system
# 예상: protodiff pod가 실행 중
```

**참고**: 스키마는 이미 https://buf.build/proto-diff-bsr/test-services 에 게시되어 있습니다 - 별도로 푸시할 필요가 없습니다!

#### 2. ProtoDiff 발견 확인

ProtoDiff가 테스트 서비스를 발견했는지 확인:

```bash
# ProtoDiff 로그 확인
kubectl logs -n protodiff-system -l app=protodiff -f

# 다음과 같은 로그가 표시되어야 합니다:
# Discovered gRPC service: grpc-server-go in namespace grpc-test
# Discovered gRPC service: grpc-server-java in namespace grpc-test
# Fetching schema for grpc-server-go...
# Comparing with BSR module: buf.build/proto-diff-bsr/test-services
```

#### 3. ProtoDiff 대시보드 접속

```bash
kubectl port-forward -n protodiff-system svc/protodiff 18080:80
```

브라우저에서 http://localhost:18080을 엽니다. 다음이 표시됩니다:

- **grpc-server-go**
  - 상태: 🟢 녹색 (스키마가 BSR과 일치)
  - BSR 모듈: `buf.build/proto-diff-bsr/test-services`
  - 서비스: `greeter.Greeter`

- **grpc-server-java**
  - 상태: 🟢 녹색 (스키마가 BSR과 일치)
  - BSR 모듈: `buf.build/proto-diff-bsr/test-services`
  - 서비스: `user.UserService`

**상태 의미:**
- 🟢 **녹색 (IN_SYNC)**: 라이브 스키마가 BSR과 일치 - 정상!
- 🔴 **빨강 (MISMATCH)**: 스키마 드리프트 감지 - 업데이트 필요
- 🟡 **노랑 (UNKNOWN)**: 스키마를 가져올 수 없거나 BSR 모듈을 찾을 수 없음

#### 4. 대시보드 이해하기

대시보드는 스키마 동기화의 현재 상태를 보여줍니다. 이 테스트 서비스의 경우 다음을 볼 수 있습니다:

- 🟢 **녹색 (IN_SYNC)**: 배포된 서비스 스키마가 BSR 스키마와 일치
- 두 서비스 모두 동일한 BSR 모듈을 가리킴: `buf.build/proto-diff-bsr/test-services`

**ProtoDiff가 확인하는 내용:**

ProtoDiff는 배포된 gRPC 서비스를 다음과 같이 지속적으로 모니터링합니다:
1. gRPC reflection을 사용하여 실행 중인 pod에서 라이브 스키마 가져오기
2. BSR에 저장된 스키마와 비교
3. 차이가 발생하면 알림

이를 통해 문서(BSR)가 실제 배포와 동기화된 상태를 유지합니다!

### 정리

```bash
# 테스트 서비스 삭제
kubectl delete -f sample-grpc-service.yaml

# 다음이 제거됩니다:
# - grpc-test 네임스페이스
# - 모든 deployment, service, pod
```

### 문제 해결

#### Pod가 시작되지 않음

```bash
# Pod 이벤트 확인
kubectl describe pod -n grpc-test <pod-name>

# 일반적인 문제:
# - ImagePullBackOff: Docker Hub에서 이미지에 접근할 수 있는지 확인
# - CrashLoopBackOff: kubectl logs로 로그 확인
```

#### 서비스 간 연결 거부

```bash
# 서비스 DNS 해석 확인
kubectl run -it --rm debug --image=busybox --restart=Never -n grpc-test -- sh

# Pod 내부에서:
nslookup grpc-server-java.grpc-test.svc.cluster.local
nslookup grpc-server-go.grpc-test.svc.cluster.local

# 연결 테스트
nc -zv grpc-server-java.grpc-test.svc.cluster.local 9091
```

#### gRPC 호출 실패

```bash
# gRPC reflection이 활성화되어 있는지 확인
grpcurl -plaintext localhost:9090 list

# "Failed to list services"가 표시되면 reflection이 활성화되지 않았을 수 있음
# 서버 로그에서 오류 확인
```

### 다음 단계

- **스키마 드리프트 모니터링**: ProtoDiff를 사용하여 배포된 서비스가 BSR과 다를 때 감지
- **더 많은 서비스 추가**: 동일한 패턴으로 추가 gRPC 서비스 생성
- **Proto 커스터마이징**: proto 정의를 수정하고 재배포하여 ProtoDiff가 변경 사항을 감지하는지 확인
- **프로덕션 배포**: 이 예제를 프로덕션 환경에 맞게 조정

### 리소스

- **메인 문서**: [../README.md](../README.md)
- **Go 서버 소스**: 상위 폴더의 `grpc-server-go/` 디렉토리 참조
- **Java 서버 소스**: 상위 폴더의 `grpc-server-java/` 디렉토리 참조
- **ProtoDiff GitHub**: https://github.com/uzdada/protodiff
- **Buf Schema Registry**: https://buf.build
