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

ProtoDiff is a Kubernetes-native monitoring tool that runs inside your cluster to automatically detect schema drift between live gRPC services and the Buf Schema Registry (BSR). Deployed as a lightweight agent in your Kubernetes environment, it continuously validates schema consistency across your microservices without requiring any changes to your existing services.

### Key Features

- **Non-Invasive Design**: Zero changes required to existing microservices (no sidecars, no YAML modifications)
- **Visual Dashboard**: Built-in HTML dashboard accessible via `kubectl port-forward`
- **Centralized Configuration**: Service-to-BSR mappings managed through ConfigMap
- **Automatic Discovery**: Discovers gRPC pods using Kubernetes labels
- **gRPC Reflection**: Uses server reflection to fetch live schemas
- **Real-time Monitoring**: Continuous validation with configurable scan intervals
- **Clear Status Indicators**: Traffic light UI (Green=Sync, Red=Mismatch, Yellow=Unknown)

### Prerequisites

- Kubernetes cluster (v1.25+)
- kubectl configured to access your cluster
- gRPC services with server reflection enabled
- Pods labeled with `grpc-service=true`
- **BSR Token** (required for schema validation)
  - Sign up at https://buf.build
  - Get your token from https://buf.build/settings/user
  - Click "Create Token" and save it securely

### Quick Start

#### 1. Get Your BSR Token

Visit https://buf.build/settings/user and create an API token. Keep it secure - you'll need it in step 3.

#### 2. Download and Configure Service Mappings

Download the installation manifest:

```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
```

Edit the ConfigMap section to map your services to BSR modules:

```bash
vi install.yaml  # or use your preferred editor
```

**Configure Service Mappings** (around line 69-71):
```yaml
data:
  user-service: "buf.build/acme/user"        # Replace with your services
  order-service: "buf.build/acme/order"      # Add more as needed
  payment-service: "buf.build/acme/payment"
```

#### 3. Deploy

Apply the manifest:

```bash
kubectl apply -f install.yaml
```

**Create BSR Token Secret** (do NOT commit this to Git):

```bash
kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_BSR_TOKEN_HERE \
  -n protodiff-system
```

Verify deployment:

```bash
kubectl get pods -n protodiff-system
```

**Security Note**: Never commit the Secret creation command or token to Git. Use environment variables, secret management tools (Sealed Secrets, External Secrets Operator, Vault), or store it in a `.env` file that's git-ignored.

#### Alternative: Automated Installation

For quick testing, use the interactive installation script:

```bash
curl -sL https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.sh | bash
```

#### Label Your gRPC Pods

Add the `grpc-service=true` label to your gRPC service pods:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  template:
    metadata:
      labels:
        app: user-service
        grpc-service: "true"  # Required for ProtoDiff discovery
```

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
│   │   ├── k8s/                # Kubernetes client
│   │   ├── grpc/               # gRPC reflection client
│   │   ├── bsr/                # BSR API client
│   │   └── web/                # HTTP server & dashboard
│   └── scanner/                # Schema validation orchestrator
├── web/templates/              # HTML dashboard templates
└── deploy/k8s/                 # Kubernetes manifests
```

### How It Works

1. **Discovery**: Scans the cluster for pods labeled `grpc-service=true`
2. **Resolution**: Resolves BSR module names using ConfigMap or template fallback
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

### Troubleshooting

#### No Services Discovered

**Issue**: Dashboard shows "No gRPC services discovered yet"

**Solutions**:
- Verify pods have `grpc-service=true` label
- Check ProtoDiff logs: `make logs`
- Ensure pods are in `Running` state

#### Schema Fetch Failed

**Issue**: Status shows "UNKNOWN" with error message

**Solutions**:
- Verify gRPC reflection is enabled on your service
- Check pod IP is accessible from ProtoDiff pod
- Ensure gRPC port is correct (default: 9090)

#### No BSR Mapping Found

**Issue**: "No BSR module mapping found" message

**Solutions**:
- Add service mapping to `protodiff-mapping` ConfigMap
- Set `DEFAULT_BSR_TEMPLATE` environment variable
- Restart ProtoDiff pod after ConfigMap changes

### Contributing

Contributions are welcome! Please see [CONTRIBUTING.md](CONTRIBUTING.md) for details.

### License

This project is licensed under the Apache License 2.0. See [LICENSE](LICENSE) for details.

### Contact

- GitHub Issues: https://github.com/uzdada/protodiff/issues

---

## Korean

### 개요

ProtoDiff는 Kubernetes 클러스터 내부에서 실행되는 네이티브 모니터링 도구로, 실행 중인 gRPC 서비스와 Buf Schema Registry(BSR) 간의 스키마 드리프트를 자동으로 감지합니다. Kubernetes 환경에 경량 에이전트로 배포되어 기존 서비스의 변경 없이 마이크로서비스 전반의 스키마 일관성을 지속적으로 검증합니다.

### 주요 기능

- **비침투적 설계**: 기존 마이크로서비스 변경 불필요 (사이드카 없음, YAML 수정 없음)
- **시각적 대시보드**: kubectl port-forward로 접근 가능한 내장 HTML 대시보드
- **중앙 집중식 설정**: ConfigMap을 통한 서비스-BSR 매핑 관리
- **자동 발견**: Kubernetes 레이블을 사용한 gRPC Pod 발견
- **gRPC Reflection**: 서버 리플렉션을 사용한 라이브 스키마 가져오기
- **실시간 모니터링**: 설정 가능한 스캔 간격의 지속적인 검증
- **명확한 상태 표시**: 신호등 UI (녹색=동기화, 빨강=불일치, 노랑=알 수 없음)

### 사전 요구사항

- Kubernetes 클러스터 (v1.25+)
- kubectl 설정 완료
- 서버 리플렉션이 활성화된 gRPC 서비스
- `grpc-service=true` 레이블이 있는 Pod
- **BSR 토큰** (스키마 검증에 필요)
  - https://buf.build 에서 가입
  - https://buf.build/settings/user 에서 토큰 생성
  - "Create Token" 클릭 후 안전하게 보관

### 빠른 시작

#### 1. BSR 토큰 발급

https://buf.build/settings/user 에 방문하여 API 토큰을 생성합니다. 안전하게 보관하세요 - 3단계에서 필요합니다.

#### 2. 다운로드 및 서비스 매핑 설정

설치 매니페스트 다운로드:

```bash
curl -O https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.yaml
```

ConfigMap 섹션을 편집하여 서비스를 BSR 모듈에 매핑:

```bash
vi install.yaml  # 또는 원하는 에디터 사용
```

**서비스 매핑 설정** (69-71번째 줄 근처):
```yaml
data:
  user-service: "buf.build/acme/user"        # 실제 서비스로 교체
  order-service: "buf.build/acme/order"      # 필요한 만큼 추가
  payment-service: "buf.build/acme/payment"
```

#### 3. 배포

매니페스트 적용:

```bash
kubectl apply -f install.yaml
```

**BSR 토큰 Secret 생성** (Git에 커밋하지 마세요):

```bash
kubectl create secret generic bsr-token \
  --from-literal=token=YOUR_BSR_TOKEN_HERE \
  -n protodiff-system
```

배포 확인:

```bash
kubectl get pods -n protodiff-system
```

**보안 주의사항**: Secret 생성 명령어나 토큰을 Git에 커밋하지 마세요. 환경 변수, 시크릿 관리 도구(Sealed Secrets, External Secrets Operator, Vault)를 사용하거나 `.env` 파일에 저장 후 git-ignore 처리하세요.

#### 대안: 자동 설치

빠른 테스트를 위해 대화형 설치 스크립트 사용:

```bash
curl -sL https://raw.githubusercontent.com/uzdada/protodiff/main/deploy/k8s/install.sh | bash
```

#### gRPC Pod에 레이블 추가

gRPC 서비스 Pod에 `grpc-service=true` 레이블 추가:

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: user-service
spec:
  template:
    metadata:
      labels:
        app: user-service
        grpc-service: "true"  # ProtoDiff 발견에 필요
```

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
│   │   ├── k8s/                # Kubernetes 클라이언트
│   │   ├── grpc/               # gRPC 리플렉션 클라이언트
│   │   ├── bsr/                # BSR API 클라이언트
│   │   └── web/                # HTTP 서버 & 대시보드
│   └── scanner/                # 스키마 검증 오케스트레이터
├── web/templates/              # HTML 대시보드 템플릿
└── deploy/k8s/                 # Kubernetes 매니페스트
```

### 동작 방식

1. **발견**: `grpc-service=true` 레이블이 있는 Pod 스캔
2. **해석**: ConfigMap 또는 템플릿 폴백을 사용한 BSR 모듈 이름 해석
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

### 문제 해결

#### 서비스가 발견되지 않음

**문제**: 대시보드에 "No gRPC services discovered yet" 표시

**해결 방법**:
- Pod에 `grpc-service=true` 레이블이 있는지 확인
- ProtoDiff 로그 확인: `make logs`
- Pod가 `Running` 상태인지 확인

#### 스키마 가져오기 실패

**문제**: 상태가 "UNKNOWN"으로 표시되고 오류 메시지 발생

**해결 방법**:
- 서비스에서 gRPC 리플렉션이 활성화되어 있는지 확인
- ProtoDiff Pod에서 Pod IP에 접근 가능한지 확인
- gRPC 포트가 올바른지 확인 (기본값: 9090)

#### BSR 매핑을 찾을 수 없음

**문제**: "No BSR module mapping found" 메시지

**해결 방법**:
- `protodiff-mapping` ConfigMap에 서비스 매핑 추가
- `DEFAULT_BSR_TEMPLATE` 환경 변수 설정
- ConfigMap 변경 후 ProtoDiff Pod 재시작

### 기여

기여를 환영합니다! 자세한 내용은 [CONTRIBUTING.md](CONTRIBUTING.md)를 참조하세요.

### 라이선스

이 프로젝트는 Apache License 2.0에 따라 라이선스가 부여됩니다. 자세한 내용은 [LICENSE](LICENSE)를 참조하세요.

### 연락처

- GitHub Issues: https://github.com/uzdada/protodiff/issues
