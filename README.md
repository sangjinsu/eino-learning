# eino-learning

Go와 CloudWeGo Eino를 단계별로 익히는 학습 저장소입니다. 목표는 작은 예제와 테스트를 쌓아 최종적으로 Eino 기반 DevOps Assistant를 만드는 것입니다.

## 빠른 시작

외부 API 없이 전체 unit test를 먼저 확인합니다.

```bash
RUN_EINO_INTEGRATION=0 OPENAI_API_KEY= go test ./...
```

가장 작은 예제부터 실행합니다.

```bash
go run ./cmd/ch01-chatmodel "Eino는 어떤 문제를 해결하나요?"
```

최근 추가된 MCP 예시는 외부 API 없이 바로 실행할 수 있습니다.

```bash
go run ./cmd/ch10-mcp-client
```

## 문서 읽는 순서

| 문서 | 용도 |
| --- | --- |
| [docs/learning-roadmap.md](docs/learning-roadmap.md) | 전체 학습 순서와 원칙 |
| [guides/chapters.md](guides/chapters.md) | chapter별 목표와 완료 기준 |
| [docs/progress.md](docs/progress.md) | 완료 상태, 대표 실행 명령, 검증 명령 |
| [docs/notes.md](docs/notes.md) | 개념 설명, 흐름 그래프, 학습 포인트 |
| [guides/workflow.md](guides/workflow.md) | 작업 순서와 테스트 원칙 |
| [guides/security.md](guides/security.md) | API key와 외부 호출 안전 규칙 |

## 학습 흐름

```mermaid
flowchart LR
    chat["ChatModel"] --> prompt["Prompt Template"]
    prompt --> openai["OpenAI integration"]
    openai --> tool["Tool Calling"]
    tool --> chain["Chain"]
    chain --> graph["Graph"]
    graph --> streaming["Streaming"]
    streaming --> callback["Callback"]
    callback --> rag["RAG"]
    rag --> mcp["MCP"]
    mcp --> agent["ReAct Agent"]
    agent --> graphtool["GraphTool"]
```

## Chapter Index

| Chapter | 핵심 주제 | CLI | API key |
| --- | --- | --- | --- |
| 01 | fake ChatModel과 service 경계 | `cmd/ch01-chatmodel` | 불필요 |
| 02 | ChatTemplate과 message 순서 | `cmd/ch02-prompt-template` | 불필요 |
| 03 | OpenAI ChatModel 설정과 opt-in integration | `cmd/ch03-openai-chatmodel` | 선택 |
| 04 | Tool Calling과 calculator tool | `cmd/ch04-tool-calling` | 필요 |
| 05 | `ChatTemplate -> ChatModel` Chain | `cmd/ch05-chain` | 필요 |
| 06 | Graph branch와 deterministic calculator path | `cmd/ch06-graph` | 필요 |
| 07 | ChatModel streaming과 chunk 수집 | `cmd/ch07-streaming` | 필요 |
| 08 | Callback과 observability timeline | `cmd/ch08-callback-observability` | 필요 |
| 09 | Markdown/Text 기반 RAG | `cmd/ch09-rag` | 필요 |
| 10 | local stdio MCP server와 demo client | `cmd/ch10-mcp-client` | 불필요 |
| 11 | ReAct Agent와 calculator tool loop | `cmd/ch11-react-agent` | 필요 |

상세한 목표와 완료 기준은 [guides/chapters.md](guides/chapters.md)를, 대표 실행 명령과 검증 명령은 [docs/progress.md](docs/progress.md)를, 각 chapter의 개념 설명과 그래프는 [docs/notes.md](docs/notes.md)를 봅니다.

## 실행 정책

- 기본 테스트는 외부 API 없이 통과해야 합니다.
- 실제 OpenAI 호출은 `RUN_EINO_INTEGRATION=1`을 명시했을 때만 실행합니다.
- `OPENAI_API_KEY`는 shell 환경 변수 또는 repository root의 `.env`에서 읽습니다.
- API key, prompt 결과, secret은 코드와 fixture에 저장하지 않습니다.

`.env` 예시:

```env
OPENAI_API_KEY=your-api-key
OPENAI_MODEL=gpt-4.1-mini
OPENAI_BASE_URL=
RUN_EINO_INTEGRATION=1
```

## 자주 쓰는 검증 명령

외부 API 없는 전체 검증:

```bash
RUN_EINO_INTEGRATION=0 OPENAI_API_KEY= go test ./...
go vet ./...
git diff --check
```

특정 chapter만 빠르게 확인:

```bash
go test ./internal/llm/chain -run 'TestService|TestNewService' -count=1
go test ./cmd/ch09-rag ./internal/llm/rag -count=1
go test ./internal/mcpdemo ./cmd/ch10-mcp-server ./cmd/ch10-mcp-client -count=1
go test ./cmd/ch11-react-agent ./internal/llm/agent -count=1
```

실제 OpenAI integration test 예시:

```bash
RUN_EINO_INTEGRATION=1 go test ./internal/llm/openai -run TestOpenAIChatModelIntegration -count=1 -v
```

## 현재 진행 상태

- Chapter 01-11은 완료되어 있습니다.
- 다음 주제는 Chapter 12 GraphTool입니다.
- 최신 완료 상태와 검증 포인트는 [docs/progress.md](docs/progress.md)에 유지합니다.
