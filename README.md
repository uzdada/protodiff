# ProtoDiff

[![CI](https://github.com/uzdada/protodiff/actions/workflows/ci.yaml/badge.svg)](https://github.com/uzdada/protodiff/actions/workflows/ci.yaml)
[![Docker Hub](https://img.shields.io/docker/v/wooojin2da/protodiff?label=docker&logo=docker)](https://hub.docker.com/r/uzdada/protodiff)
[![Docker Pulls](https://img.shields.io/docker/pulls/wooojin2da/protodiff)](https://hub.docker.com/r/wooojin2da/protodiff)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

**Kubernetes-native gRPC Schema Drift Detection Tool**

[English](#english) | [한국어](#korean)

---

## English

### Overview

**Ever deployed a gRPC service update and forgot to sync your schema registry?**

ProtoDiff is here to save you from that nightmare. It's a Kubernetes-native monitoring tool that automatically catches schema drift between your running gRPC services and the Buf Schema Registry (BSR) - **before it breaks production**.

Think of it as your **schema consistency guardian**: a lightweight agent that lives in your cluster, continuously validates your microservices, and alerts you the moment things go out of sync. No sidecars, no code changes, no hassle - just deploy and forget.

### Why You'll Love It

- 🚀 **Zero-Touch Deployment**: No sidecars, no service changes, no code modifications - just deploy and it works
- 📊 **Visual Dashboard**: See all your services at a glance with a clean, built-in web UI
- ⚙️ **Dead Simple Config**: Map services to BSR modules in one ConfigMap - that's it
- 🔍 **Auto-Discovery**: Point it at your cluster and it finds all your gRPC services automatically
- ⚡ **Real-Time Alerts**: Know within 30 seconds when schemas drift (configurable)
- 🎯 **Crystal Clear Status**: Traffic light indicators - Green (✓ synced), Red (✗ drift), Yellow (? unknown)
- 🔧 **Production Ready**: Multi-arch support (AMD64/ARM64), proven in real clusters

### What You'll Need

- **Kubernetes cluster** (v1.25 or newer)
- **kubectl** configured and working
- **gRPC services** with server reflection enabled (most frameworks support this)
- **BSR Token** (for private schemas - public modules work without it)
  - Get yours free at https://buf.build/settings/user
  - Takes 30 seconds to create
  - **Tip**: Testing with public modules? Skip the token entirely!

### Quick Start

> **Want to try it first?** Check out the [**one-command demo**](examples/SAMPLE_QUICKSTART.md) - sets up everything in 60 seconds!

#### 1. Get Your BSR Token (Optional for Public Modules)

Grab a free API token from https://buf.build/settings/user - takes 30 seconds. Testing with public BSR modules? Skip this step!

#### 2. Download & Configure

Grab the installation manifest:

```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
```

Open it up and configure your services (around line 69-71):

```bash
vi install.yaml
```

Tell ProtoDiff which services to monitor:
```yaml
data:
  user-service: "buf.build/acme/user"        # Your service → Your BSR module
  order-service: "buf.build/acme/order"      # Add as many as you need
  payment-service: "buf.build/acme/payment"
```

#### 3. Set Your BSR Token

**Option A: Edit install.yaml (for testing/quickstart)**

Find the Secret section in install.yaml and add your token:

```yaml
---
# Secret for BSR authentication token
apiVersion: v1
kind: Secret
metadata:
  name: bsr-token
  namespace: protodiff-system
stringData:
  token: "YOUR_BSR_TOKEN_HERE"  # Replace with your actual token
```

**Security Warning**: Only use this method for local testing. Never commit real tokens to Git!

**Option B: Create Secret Manually (recommended for production)**

Keep the Secret section in install.yaml with empty token value, then create it separately:

```bash
kubectl apply -f install.yaml

kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_BSR_TOKEN_HERE \
  -n protodiff-system \
  --dry-run=client -o yaml | kubectl apply -f -
```

Verify deployment:

```bash
kubectl get pods -n protodiff-system
```

**Security Note**: For production, use secret management tools (Sealed Secrets, External Secrets Operator, Vault) instead of storing tokens in Git or plain kubectl commands.

#### Alternative: Automated Installation

For quick testing, use the interactive installation script:

```bash
curl -sL https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.sh | bash
```

**Note**: Services specified in the ConfigMap will be automatically discovered by their `app` label. No additional labels are required.

#### Access the Dashboard

```bash
kubectl port-forward -n protodiff-system svc/protodiff 18080:80
```

Open your browser to http://localhost:18080

### Architecture

ProtoDiff follows the Hexagonal Architecture pattern with a clean separation of concerns:

```
protodiff/
├── cmd/protodiff/              # Application entrypoint
├── internal/
│   ├── core/
│   │   ├── domain/             # Business models
│   │   └── store/              # Thread-safe in-memory storage
│   ├── adapters/
│   │   ├── k8s/                # Kubernetes client (auto port detection)
│   │   ├── grpc/               # gRPC reflection client
│   │   ├── bsr/                # BSR clients (buf CLI & HTTP)
│   │   └── web/                # HTTP server & dashboard
│   └── scanner/                # Schema validation orchestrator
├── web/templates/              # HTML dashboard templates
└── deploy/k8s/                 # Kubernetes manifests
```

**BSR Integration Methods:**
- **BufClient** (default): Uses `buf export` CLI for reliable schema fetching
- **HTTPClient** (experimental): Direct BSR API access (not production-ready)

### How It Works

1. **Discovery**: Scans the cluster for services specified in ConfigMap (or falls back to label-based discovery if ConfigMap is empty)
2. **Resolution**: Resolves BSR module names from ConfigMap entries
3. **Validation**:
   - Fetches "live schema" from pod via gRPC Reflection
   - Fetches "truth schema" from Buf Schema Registry
   - Compares and detects drift
4. **Storage**: Stores results in thread-safe in-memory store
5. **Dashboard**: Renders real-time status via web UI

### Configuration

#### Environment Variables

| Variable               | Description                           | Default               |
|------------------------|---------------------------------------|-----------------------|
| `CONFIGMAP_NAMESPACE`  | Namespace of the mapping ConfigMap    | `protodiff-system`    |
| `CONFIGMAP_NAME`       | Name of the mapping ConfigMap         | `protodiff-mapping`   |
| `DEFAULT_BSR_TEMPLATE` | Fallback BSR module template          | `""`                  |
| `WEB_ADDR`             | Web server listen address             | `:18080`              |
| `SCAN_INTERVAL`        | Time between scans                    | `30s`                 |

#### BSR Template

If a service is not found in the ConfigMap, ProtoDiff can use a template:

```bash
DEFAULT_BSR_TEMPLATE="buf.build/acme/{service}"
```

For a service named `user-service`, this resolves to `buf.build/acme/user-service`.

### Development

#### Prerequisites

- Go 1.21+
- Docker
- kubectl
- Access to a Kubernetes cluster

#### Building from Source

```bash
git clone https://github.com/uzdada/protodiff.git
cd protodiff

make deps          # Install dependencies
make build         # Build binary
make test          # Run tests
make docker-build  # Build Docker image
```

#### Local Development

```bash
make run           # Run locally (requires kubeconfig)
make fmt           # Format code
make lint          # Run linter
```

#### Docker Hub CI/CD Setup

The project automatically builds and pushes Docker images to Docker Hub when changes are pushed to the main branch.

**For Repository Maintainers:**

Set up the following GitHub Secrets:

1. Go to repository Settings → Secrets and variables → Actions
2. Add the following secrets:
   - `DOCKERHUB_USERNAME`: Your Docker Hub username
   - `DOCKERHUB_TOKEN`: Docker Hub access token (create at https://hub.docker.com/settings/security)

**Docker Image Tags:**
- `wooojin2da/protodiff:latest` - Latest build from main branch
- `wooojin2da/protodiff:main-<sha>` - Specific commit SHA

### Technical Details

#### Port Auto-Detection

ProtoDiff automatically detects gRPC ports from pod container specifications:
- Scans pod's `containerPort` definitions
- Looks for ports named "grpc" or using TCP protocol
- Falls back to default port 9090 if not specified

**Example:**
```yaml
ports:
  - name: grpc
    containerPort: 9091  # Automatically detected
```

#### Multi-Architecture Support

All components support both AMD64 and ARM64 architectures:
- Docker images built with `docker buildx --platform linux/amd64,linux/arm64`
- Tested on: x86_64 Linux, Apple Silicon (M1/M2), AWS Graviton

### Troubleshooting

#### No Services Discovered

**Issue**: Dashboard shows "No gRPC services discovered yet"

**Solutions**:
- Verify services are listed in the ConfigMap: `kubectl get configmap protodiff-mapping -n protodiff-system -o yaml`
- Ensure your service pods have the `app` label matching the service name in ConfigMap
- Check ProtoDiff logs: `kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff`
- Ensure pods are in `Running` state

#### Schema Fetch Failed

**Issue**: Status shows "UNKNOWN" with error message

**Solutions**:
- Verify gRPC reflection is enabled on your service
- Check pod IP is accessible from ProtoDiff pod
- Ensure gRPC port is correct (auto-detected from containerPort or defaults to 9090)
- Check logs for "Connection refused" errors

#### BSR Export Failed

**Issue**: "buf export failed: read-only file system"

**Solution**: This is already fixed in the latest deployment manifest. The deployment includes:
- Writable `/tmp` volume mount (`emptyDir`)
- `HOME=/tmp` environment variable for buf CLI cache

If using an old manifest, update with:
```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
kubectl apply -f install.yaml
```

#### No BSR Mapping Found

**Issue**: "No BSR module mapping found" message

**Solutions**:
- Add service mapping to `protodiff-mapping` ConfigMap
- Set `DEFAULT_BSR_TEMPLATE` environment variable
- Restart ProtoDiff pod after ConfigMap changes: `kubectl rollout restart deployment/protodiff -n protodiff-system`

### Contributing

We'd love your help making ProtoDiff even better! Found a bug? Have a brilliant idea? Check out [CONTRIBUTING.md](CONTRIBUTING.md) to get started.

**If this project helps you** ⭐ please star the repo - it really motivates us!

### License

This project is licensed under Apache License 2.0. Feel free to use it! See [LICENSE](LICENSE) for details.

### Contact

- Issues/Questions/Suggestions: https://github.com/uzdada/protodiff/issues

---

**Let ProtoDiff guard your microservices!** 🛡️

---

## Korean

### 개요

**gRPC 서비스를 업데이트하고 스키마 레지스트리 동기화를 깜빡하신 적 있나요?**

ProtoDiff가 그 악몽에서 구해드릴게요. 실행 중인 gRPC 서비스와 Buf Schema Registry(BSR) 사이의 스키마 드리프트를 자동으로 잡아내는 Kubernetes 네이티브 모니터링 도구입니다 - **프로덕션이 망가지기 전에** 말이죠.

**스키마 일관성 수호자**라고 생각하시면 돼요: 클러스터에 상주하면서 마이크로서비스를 계속 검증하고, 뭔가 동기화가 안 되는 순간 즉시 알려드려요. 사이드카도 필요 없고, 코드 수정도 필요 없고, 골치 아픈 것도 없어요 - 그냥 배포하고 잊으시면 됩니다.

### 왜 좋아하실 거예요

- 🚀 **제로 터치 배포**: 사이드카도, 서비스 변경도, 코드 수정도 필요 없어요 - 그냥 배포하면 작동해요
- 📊 **한눈에 보는 대시보드**: 깔끔한 웹 UI로 모든 서비스를 한눈에 확인
- ⚙️ **초간단 설정**: ConfigMap 하나로 서비스를 BSR 모듈에 매핑 - 끝!
- 🔍 **자동 발견**: 클러스터를 가리키면 모든 gRPC 서비스를 자동으로 찾아요
- ⚡ **실시간 알림**: 30초 안에 스키마 드리프트를 알 수 있어요 (설정 가능)
- 🎯 **명확한 상태**: 신호등 표시 - 초록색 (✓ 동기화), 빨강 (✗ 드리프트), 노랑 (? 알 수 없음)
- 🔧 **프로덕션 준비 완료**: 멀티 아키텍처 지원 (AMD64/ARM64), 실제 클러스터에서 검증됨

### 필요한 것들

- **Kubernetes 클러스터** (v1.25 이상)
- **kubectl** 설정 완료 및 정상 작동
- **서버 reflection이 켜진 gRPC 서비스** (대부분의 프레임워크가 지원해요)
- **BSR 토큰** (프라이빗 스키마용 - 퍼블릭 모듈은 토큰 없이도 돼요)
  - https://buf.build/settings/user 에서 무료로 받으세요
  - 30초면 만들 수 있어요
  - **꿀팁**: 퍼블릭 모듈로 테스트하신다고요? 토큰 건너뛰셔도 됩니다!

### 빠른 시작

> **먼저 체험해보고 싶으신가요?** [**원 커맨드 데모**](examples/SAMPLE_QUICKSTART.md)를 확인하세요 - 60초면 모든 설정이 끝나요!

#### 1. BSR 토큰 받기 (퍼블릭 모듈은 선택사항)

https://buf.build/settings/user 에서 무료 API 토큰을 받으세요 - 30초 걸려요. 퍼블릭 BSR 모듈로 테스트하신다고요? 이 단계는 건너뛰세요!

#### 2. 다운로드 & 설정

설치 매니페스트를 받아오세요:

```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
```

파일을 열어서 서비스를 설정하세요 (69-71번째 줄 근처):

```bash
vi install.yaml
```

ProtoDiff에게 어떤 서비스를 모니터링할지 알려주세요:
```yaml
data:
  user-service: "buf.build/acme/user"        # 여러분의 서비스 → BSR 모듈
  order-service: "buf.build/acme/order"      # 필요한 만큼 추가하세요
  payment-service: "buf.build/acme/payment"
```

#### 3. BSR 토큰 설정

**옵션 A: install.yaml 편집 (테스트/빠른 시작용)**

install.yaml의 Secret 섹션을 찾아 토큰을 추가하세요:

```yaml
---
# BSR 인증 토큰용 Secret
apiVersion: v1
kind: Secret
metadata:
  name: bsr-token
  namespace: protodiff-system
stringData:
  token: "YOUR_BSR_TOKEN_HERE"  # 실제 토큰으로 교체
```

**보안 경고**: 이 방법은 로컬 테스트용으로만 사용하세요. 실제 토큰을 Git에 커밋하지 마세요!

**옵션 B: Secret 수동 생성 (프로덕션 권장)**

install.yaml의 Secret 섹션은 빈 토큰 값으로 유지하고, 별도로 생성하세요:

```bash
kubectl apply -f install.yaml

kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_BSR_TOKEN_HERE \
  -n protodiff-system \
  --dry-run=client -o yaml | kubectl apply -f -
```

배포 확인:

```bash
kubectl get pods -n protodiff-system
```

**보안 주의사항**: 프로덕션 환경에서는 Git이나 평문 kubectl 명령어에 토큰을 저장하는 대신, 시크릿 관리 도구(Sealed Secrets, External Secrets Operator, Vault)를 사용하세요.

#### 대안: 자동 설치

빠른 테스트를 위해 대화형 설치 스크립트 사용:

```bash
curl -sL https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.sh | bash
```

**참고**: ConfigMap에 지정된 서비스는 `app` 레이블을 통해 자동으로 발견됩니다. 추가 레이블이 필요하지 않습니다.

#### 대시보드 접속

```bash
kubectl port-forward -n protodiff-system svc/protodiff 18080:80
```

브라우저에서 http://localhost:18080 열기

### 아키텍처

ProtoDiff는 헥사고날 아키텍처 패턴을 따르며 관심사를 명확히 분리합니다:

```
protodiff/
├── cmd/protodiff/              # 애플리케이션 엔트리포인트
├── internal/
│   ├── core/
│   │   ├── domain/             # 비즈니스 모델
│   │   └── store/              # Thread-safe 인메모리 저장소
│   ├── adapters/
│   │   ├── k8s/                # Kubernetes 클라이언트 (자동 포트 감지)
│   │   ├── grpc/               # gRPC 리플렉션 클라이언트
│   │   ├── bsr/                # BSR 클라이언트 (buf CLI & HTTP)
│   │   └── web/                # HTTP 서버 & 대시보드
│   └── scanner/                # 스키마 검증 오케스트레이터
├── web/templates/              # HTML 대시보드 템플릿
└── deploy/k8s/                 # Kubernetes 매니페스트
```

**BSR 통합 방식:**
- **BufClient** (기본값): 안정적인 스키마 가져오기를 위해 `buf export` CLI 사용
- **HTTPClient** (실험적): 직접 BSR API 접근 (프로덕션 미지원)

### 동작 방식

1. **발견**: ConfigMap에 지정된 서비스 스캔 (ConfigMap이 비어있으면 레이블 기반 발견으로 폴백)
2. **해석**: ConfigMap 항목에서 BSR 모듈 이름 해석
3. **검증**:
   - gRPC Reflection을 통해 Pod에서 "라이브 스키마" 가져오기
   - Buf Schema Registry에서 "진실 스키마" 가져오기
   - 비교 및 드리프트 감지
4. **저장**: Thread-safe 인메모리 저장소에 결과 저장
5. **대시보드**: 웹 UI를 통한 실시간 상태 렌더링

### 설정

#### 환경 변수

| 변수                   | 설명                           | 기본값                |
|------------------------|--------------------------------|-----------------------|
| `CONFIGMAP_NAMESPACE`  | 매핑 ConfigMap의 네임스페이스   | `protodiff-system`    |
| `CONFIGMAP_NAME`       | 매핑 ConfigMap의 이름           | `protodiff-mapping`   |
| `DEFAULT_BSR_TEMPLATE` | 폴백 BSR 모듈 템플릿            | `""`                  |
| `WEB_ADDR`             | 웹 서버 수신 주소              | `:18080`              |
| `SCAN_INTERVAL`        | 스캔 간격                      | `30s`                 |

#### BSR 템플릿

ConfigMap에서 서비스를 찾을 수 없는 경우 ProtoDiff는 템플릿을 사용할 수 있습니다:

```bash
DEFAULT_BSR_TEMPLATE="buf.build/acme/{service}"
```

`user-service`라는 서비스의 경우 `buf.build/acme/user-service`로 해석됩니다.

### 개발

#### 사전 요구사항

- Go 1.21+
- Docker
- kubectl
- Kubernetes 클러스터 접근 권한

#### 소스에서 빌드

```bash
git clone https://github.com/uzdada/protodiff.git
cd protodiff

make deps          # 의존성 설치
make build         # 바이너리 빌드
make test          # 테스트 실행
make docker-build  # Docker 이미지 빌드
```

#### 로컬 개발

```bash
make run           # 로컬 실행 (kubeconfig 필요)
make fmt           # 코드 포맷
make lint          # 린터 실행
```

### 기술 세부사항

#### 포트 자동 감지

ProtoDiff는 Pod 컨테이너 스펙에서 자동으로 gRPC 포트를 감지합니다:
- Pod의 `containerPort` 정의를 스캔
- "grpc" 이름 또는 TCP 프로토콜 사용 포트 검색
- 지정되지 않은 경우 기본 포트 9090으로 폴백

**예시:**
```yaml
ports:
  - name: grpc
    containerPort: 9091  # 자동으로 감지됨
```

#### 다중 아키텍처 지원

모든 구성요소가 AMD64와 ARM64 아키텍처를 모두 지원합니다:
- Docker 이미지는 `docker buildx --platform linux/amd64,linux/arm64`로 빌드
- 지원 환경: x86_64 Linux, Apple Silicon (M1/M2), AWS Graviton

### 문제 해결

#### 서비스가 안 보여요

**증상**: 대시보드에 "No gRPC services discovered yet"라고 떠요

**해결책**:
- ConfigMap에 서비스가 제대로 들어갔는지 확인: `kubectl get configmap protodiff-mapping -n protodiff-system -o yaml`
- 서비스 Pod에 ConfigMap의 이름과 같은 `app` 레이블이 있는지 확인하세요
- ProtoDiff 로그를 확인해보세요: `kubectl logs -n protodiff-system -l app.kubernetes.io/name=protodiff`
- Pod가 `Running` 상태인지 체크!

#### 스키마를 못 가져와요

**증상**: 상태가 "UNKNOWN"이고 에러 메시지가 나와요

**해결책**:
- 서비스에 gRPC reflection이 켜져 있는지 확인하세요
- ProtoDiff Pod에서 서비스 Pod IP에 접근할 수 있는지 확인
- gRPC 포트가 맞는지 확인 (containerPort에서 자동 감지하거나 기본값 9090)
- "Connection refused" 에러가 있는지 로그 확인

#### BSR Export가 실패해요

**증상**: "buf export failed: read-only file system" 에러

**해결책**: 최신 배포 매니페스트에서 이미 고쳐졌어요. 다음이 포함되어 있습니다:
- 쓰기 가능한 `/tmp` 볼륨 마운트 (`emptyDir`)
- buf CLI 캐시용 `HOME=/tmp` 환경 변수

예전 매니페스트 쓰고 계시면 업데이트하세요:
```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
kubectl apply -f install.yaml
```

#### BSR 매핑을 못 찾겠어요

**증상**: "No BSR module mapping found" 메시지

**해결책**:
- `protodiff-mapping` ConfigMap에 서비스 매핑을 추가하세요
- `DEFAULT_BSR_TEMPLATE` 환경 변수를 설정하세요
- ConfigMap 바꾼 후엔 ProtoDiff Pod 재시작: `kubectl rollout restart deployment/protodiff -n protodiff-system`

### 기여하기

여러분의 기여를 환영해요! 버그를 찾으셨거나 멋진 아이디어가 있으신가요? [CONTRIBUTING.md](CONTRIBUTING.md)를 확인해주세요.

**도움이 되셨다면** ⭐ 스타 하나 눌러주시면 정말 감사하겠습니다!

### 라이선스

이 프로젝트는 Apache License 2.0으로 배포됩니다. 자유롭게 사용하세요! 자세한 내용은 [LICENSE](LICENSE)를 참고하세요.

### 연락처

- 이슈/질문/제안: https://github.com/uzdada/protodiff/issues

---

**ProtoDiff가 여러분의 마이크로서비스를 안전하게 지켜드릴게요!** 🛡️
