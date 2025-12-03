# ProtoDiff Demo - Quick Start Guide

[English](#english) | [한국어](#korean)

---

## English

### 🎯 What You'll Get

Experience **ProtoDiff in action** with just one command! This demo sets up:

- **Two sample gRPC services** (Go + Java) that communicate with each other
- **ProtoDiff monitoring** to track schema drift in real-time
- **Live dashboard** showing schema validation status

**Ready to see it work?** You're 60 seconds away from a running demo.

### Communication Flow
```
Client → Go Greeter Service → Java UserService
         (SayHelloToUser)      (GetUser)
```

The Go service fetches user data from Java service to create personalized greetings - a perfect example of microservice communication that ProtoDiff can monitor!

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

- **Kubernetes cluster** (minikube, kind, Docker Desktop, or any cloud provider)
- **kubectl** configured and connected to your cluster

That's it! The demo script handles everything else.

### 🚀 Quick Start (One Command!)

Clone the repository and run the automated demo:

```bash
git clone https://github.com/uzdada/protodiff.git
cd protodiff/examples
./demo.sh
```

**What happens automatically:**
1. ✅ Deploys two gRPC test services (Go + Java)
2. ✅ Deploys ProtoDiff monitoring agent
3. ✅ Waits for all pods to be ready
4. ✅ Sets up port-forwarding
5. ✅ Opens the dashboard in your browser

In about 60 seconds, you'll see the **ProtoDiff dashboard** showing real-time schema validation!

**Dashboard URL:** http://localhost:18080

Press `Ctrl+C` when you're done to clean up all port-forwards.

### 🧪 Try It Out!

The demo script already set up port-forwarding for you. Here are some quick tests:

**Test inter-service communication** (Go → Java):

```bash
# The Go service will call Java service to fetch user data
grpcurl -plaintext -d '{"user_id": 1}' localhost:9090 greeter.Greeter/SayHelloToUser

# Response: "Hello admin (ID: 1, Email: admin@example.com)! Greetings from Go Greeter Service!"
```

**Simple greeting:**

```bash
grpcurl -plaintext -d '{"name": "World"}' localhost:9090 greeter.Greeter/SayHello
# Response: "Hello World from Go server!"
```

**Get user directly from Java service:**

```bash
grpcurl -plaintext -d '{"user_id": 1}' localhost:9091 user.UserService/GetUser
# Returns user details: admin@example.com
```

> **Note:** If `grpcurl` is not installed, skip these tests and just explore the dashboard!

### 📊 Understanding the Dashboard

Open **http://localhost:18080** (the demo script should have opened it automatically).

You'll see both test services with their schema validation status:

**Expected Dashboard View:**

| Service | Status | BSR Module | What It Means |
|---------|--------|------------|---------------|
| **grpc-server-go** | 🟢 IN_SYNC | `buf.build/proto-diff-bsr/test-services` | Schema matches! |
| **grpc-server-java** | 🟢 IN_SYNC | `buf.build/proto-diff-bsr/test-services` | Schema matches! |

**Status Indicators:**
- 🟢 **Green (IN_SYNC)**: Your deployed service matches the BSR schema - perfect!
- 🔴 **Red (MISMATCH)**: Uh-oh! Schema drift detected - time to sync
- 🟡 **Yellow (UNKNOWN)**: Can't verify (check if service is running)

**What's Happening Behind the Scenes:**

ProtoDiff is continuously (every 30 seconds):
1. Using gRPC reflection to fetch live schemas from your running pods
2. Comparing them against schemas in Buf Schema Registry
3. Alerting you immediately when they diverge

This keeps your documentation (BSR) perfectly synced with your actual deployments - **no more "the docs are outdated" moments**!

> **Fun fact:** The test schemas are already published at https://buf.build/proto-diff-bsr/test-services as a public BSR module, so you can try this demo without any BSR account setup!

### 🧹 Cleanup

Press `Ctrl+C` in the terminal where `demo.sh` is running - it automatically cleans up all port-forwards!

To completely remove the demo:

```bash
kubectl delete namespace grpc-test
kubectl delete namespace protodiff-system
```

### 🚀 What's Next?

Now that you've seen ProtoDiff in action, here's how to use it with **your own services**:

1. **Deploy ProtoDiff** to your cluster ([see main README](../README.md))
2. **Configure service mappings** in the ConfigMap to point to your BSR modules
3. **Watch the magic happen** - ProtoDiff will automatically discover and monitor your gRPC services

**Want to experiment more?**
- Try modifying the proto files and redeploying to see schema drift detection
- Add your own gRPC services following the same pattern
- Explore the detailed logs: `kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f`

### 💡 Why This Matters

Schema drift is a **silent killer** in microservices:
- Deploy a new service version without updating BSR → clients break
- Update proto files but forget to push to BSR → documentation is wrong
- Services drift apart over time → integration nightmares

ProtoDiff solves this by **continuously validating** that your running services match your source of truth. Think of it as a **smoke detector for schema drift** - catching problems before they become fires.

### 📚 Learn More

- **Main Documentation**: [../README.md](../README.md) - Production deployment guide
- **ProtoDiff GitHub**: https://github.com/uzdada/protodiff - Star us if this helps!
- **Buf Schema Registry**: https://buf.build - Where your schemas live
- **Architecture Deep Dive**: [../docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)

---

**Found this useful?** ⭐ Star the repo and share with your team!

---

## Korean

### 🎯 무엇을 경험하게 되나요?

단 하나의 명령어로 **ProtoDiff를 직접 체험**해보세요! 이 데모는 다음을 자동으로 설정합니다:

- **두 개의 샘플 gRPC 서비스** (Go + Java)가 서로 통신
- **실시간 스키마 드리프트 추적**을 위한 ProtoDiff 모니터링
- **스키마 검증 상태**를 보여주는 라이브 대시보드

**지금 바로 시작할 준비 되셨나요?** 60초면 실행 중인 데모를 확인할 수 있어요.

### 통신 흐름
```
클라이언트 → Go Greeter Service → Java UserService
            (SayHelloToUser)      (GetUser)
```

Go 서비스가 Java 서비스에서 사용자 데이터를 가져와서 개인화된 인사말을 만들어요 - ProtoDiff가 모니터링할 수 있는 완벽한 마이크로서비스 통신 예제입니다!

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

- **Kubernetes 클러스터** (minikube, kind, Docker Desktop, 또는 클라우드 서비스 아무거나)
- **kubectl** 설정 완료 및 클러스터 연결됨

이게 전부예요! 나머지는 데모 스크립트가 알아서 처리합니다.

### 🚀 빠른 시작 (명령어 하나로!)

저장소를 클론하고 자동 데모를 실행하세요:

```bash
git clone https://github.com/uzdada/protodiff.git
cd protodiff/examples
./demo.sh
```

**자동으로 진행되는 작업:**
1. ✅ 두 개의 gRPC 테스트 서비스 배포 (Go + Java)
2. ✅ ProtoDiff 모니터링 에이전트 배포
3. ✅ 모든 Pod가 준비될 때까지 대기
4. ✅ 포트 포워딩 설정
5. ✅ 브라우저에서 대시보드 자동 오픈

약 60초 후면 **실시간 스키마 검증**을 보여주는 ProtoDiff 대시보드를 확인할 수 있어요!

**대시보드 URL:** http://localhost:18080

작업이 끝나면 `Ctrl+C`를 눌러서 모든 포트 포워딩을 정리하세요.

### 🧪 직접 테스트해보세요!

데모 스크립트가 이미 포트 포워딩을 설정해뒀어요. 간단한 테스트를 해볼까요:

**서비스 간 통신 테스트** (Go → Java):

```bash
# Go 서비스가 Java 서비스를 호출해서 사용자 데이터를 가져옵니다
grpcurl -plaintext -d '{"user_id": 1}' localhost:9090 greeter.Greeter/SayHelloToUser

# 응답: "Hello admin (ID: 1, Email: admin@example.com)! Greetings from Go Greeter Service!"
```

**간단한 인사말:**

```bash
grpcurl -plaintext -d '{"name": "World"}' localhost:9090 greeter.Greeter/SayHello
# 응답: "Hello World from Go server!"
```

**Java 서비스에서 직접 사용자 정보 가져오기:**

```bash
grpcurl -plaintext -d '{"user_id": 1}' localhost:9091 user.UserService/GetUser
# 사용자 상세 정보 반환: admin@example.com
```

> **참고:** `grpcurl`이 설치되어 있지 않다면 이 테스트는 건너뛰고 대시보드만 둘러보세요!

### 📊 대시보드 살펴보기

**http://localhost:18080**을 열어보세요 (데모 스크립트가 자동으로 열어줬을 거예요).

두 테스트 서비스의 스키마 검증 상태를 확인할 수 있습니다:

**예상되는 대시보드 화면:**

| 서비스 | 상태 | BSR 모듈 | 의미 |
|---------|--------|------------|------|
| **grpc-server-go** | 🟢 IN_SYNC | `buf.build/proto-diff-bsr/test-services` | 스키마가 일치해요! |
| **grpc-server-java** | 🟢 IN_SYNC | `buf.build/proto-diff-bsr/test-services` | 스키마가 일치해요! |

**상태 표시 의미:**
- 🟢 **초록색 (IN_SYNC)**: 배포된 서비스가 BSR 스키마와 완벽하게 일치 - 완벽해요!
- 🔴 **빨간색 (MISMATCH)**: 앗! 스키마 드리프트가 감지됨 - 동기화할 시간이에요
- 🟡 **노란색 (UNKNOWN)**: 검증할 수 없음 (서비스가 실행 중인지 확인해보세요)

**무대 뒤에서 일어나는 일:**

ProtoDiff는 계속해서 (30초마다):
1. gRPC reflection을 사용해서 실행 중인 Pod에서 라이브 스키마를 가져와요
2. Buf Schema Registry에 있는 스키마와 비교해요
3. 차이가 생기면 즉시 알려드려요

이렇게 하면 문서(BSR)가 실제 배포와 완벽하게 동기화된 상태를 유지할 수 있어요 - **"문서가 오래됐어요" 같은 말은 이제 안 해도 돼요**!

> **꿀팁:** 테스트 스키마는 이미 https://buf.build/proto-diff-bsr/test-services 에 퍼블릭 BSR 모듈로 게시되어 있어서, BSR 계정 설정 없이도 이 데모를 바로 체험할 수 있어요!

### 🧹 정리하기

`demo.sh`가 실행 중인 터미널에서 `Ctrl+C`를 누르세요 - 모든 포트 포워딩이 자동으로 정리돼요!

데모를 완전히 제거하려면:

```bash
kubectl delete namespace grpc-test
kubectl delete namespace protodiff-system
```

### 🚀 다음은 뭘 해볼까요?

ProtoDiff가 실제로 동작하는 걸 보셨으니, 이제 **여러분의 서비스**에 적용해볼 차례예요:

1. **ProtoDiff를 클러스터에 배포**하세요 ([메인 README 참고](../README.md))
2. **ConfigMap에서 서비스 매핑을 설정**해서 여러분의 BSR 모듈을 가리키게 하세요
3. **마법이 일어나는 걸 지켜보세요** - ProtoDiff가 자동으로 gRPC 서비스를 찾아서 모니터링해요

**더 실험해보고 싶으신가요?**
- proto 파일을 수정하고 재배포해서 스키마 드리프트 감지를 확인해보세요
- 같은 패턴으로 여러분만의 gRPC 서비스를 추가해보세요
- 상세한 로그 확인: `kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff -f`

### 💡 왜 이게 중요할까요?

스키마 드리프트는 마이크로서비스의 **조용한 살인자**예요:
- BSR 업데이트 없이 새 서비스 버전 배포 → 클라이언트가 깨져요
- proto 파일 업데이트했는데 BSR에 푸시 안 함 → 문서가 틀려요
- 시간이 지나면서 서비스들이 따로 놀기 시작 → 통합 악몽

ProtoDiff는 실행 중인 서비스가 진실의 원천(BSR)과 일치하는지 **지속적으로 검증**해서 이 문제를 해결해요. **스키마 드리프트를 위한 화재 감지기**라고 생각하시면 돼요 - 불이 나기 전에 문제를 잡아내는 거죠.

### 📚 더 알아보기

- **메인 문서**: [../README.md](../README.md) - 프로덕션 배포 가이드
- **ProtoDiff GitHub**: https://github.com/uzdada/protodiff - 도움이 되셨다면 스타 부탁드려요!
- **Buf Schema Registry**: https://buf.build - 스키마가 저장되는 곳
- **아키텍처 심화**: [../docs/ARCHITECTURE.md](../docs/ARCHITECTURE.md)

---

**유용하셨나요?** ⭐ 레포에 스타 주시고 팀원들과 공유해주세요!
