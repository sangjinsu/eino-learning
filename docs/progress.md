# Progress

이 문서는 현재 완료 상태를 빠르게 확인하기 위한 dashboard입니다. 자세한 개념 설명은 [notes.md](notes.md), chapter별 목표와 완료 기준은 [../guides/chapters.md](../guides/chapters.md)를 봅니다.

## Status Dashboard

| Chapter | 상태 | 대표 실행 | API key | 빠른 검증 |
| --- | --- | --- | --- | --- |
| 01. ChatModel | 완료 | `go run ./cmd/ch01-chatmodel "Eino는 어떤 문제를 해결하나요?"` | 불필요 | `go test ./internal/llm/chat ./internal/fake -count=1` |
| 02. Prompt Template | 완료 | `go run ./cmd/ch02-prompt-template 'Prompt Template은 어떤 메시지를 만드나요?'` | 불필요 | `go test ./internal/llm/chat ./internal/llm/prompting -count=1` |
| 03. OpenAI ChatModel | 완료 | `go run ./cmd/ch03-openai-chatmodel 'Eino ChatModel은 어떤 역할인가요?'` | 선택 | `RUN_EINO_INTEGRATION=0 go test ./internal/llm/openai -count=1` |
| 04. Tool Calling | 완료 | `go run ./cmd/ch04-tool-calling '15 * (2 + 6)'` | 필요 | `RUN_EINO_INTEGRATION=0 go test ./internal/llm/toolcalling ./internal/tools -count=1` |
| 05. Chain | 완료 | `go run ./cmd/ch05-chain 'Chain은 Prompt Template과 ChatModel을 어떻게 연결하나요?'` | 필요 | `go test ./internal/llm/chain -run 'TestService|TestNewService' -count=1` |
| 06. Graph | 완료 | `go run ./cmd/ch06-graph 'calculate: 9 * (3 + 4)'` | 필요 | `go test ./internal/llm/graph -run 'TestAssistantGraph|TestNewService' -count=1` |
| 07. Streaming | 완료 | `go run ./cmd/ch07-streaming 'Streaming은 Generate와 무엇이 다른가요?'` | 필요 | `go test ./internal/llm/streaming -run 'TestChatService.*|TestCollectMessageStream' -count=1` |
| 08. Callback | 완료 | `go run ./cmd/ch08-callback-observability 'ChatTemplate과 ChatModel 실행을 callback으로 어떻게 관찰하나요?'` | 필요 | `go test ./internal/llm/observability -run 'TestRunObservableChatChain' -count=1` |
| 09. RAG | 완료 | `go run ./cmd/ch09-rag 'RAG는 검색된 문서를 어떻게 답변 근거로 사용하나요?'` | 필요 | `go test ./cmd/ch09-rag ./internal/llm/rag -count=1` |
| 10. MCP | 완료 | `go run ./cmd/ch10-mcp-client` | 불필요 | `go test ./internal/mcpdemo ./cmd/ch10-mcp-server ./cmd/ch10-mcp-client -count=1` |

## 공통 검증

외부 API 없이 전체 테스트를 실행합니다.

```bash
RUN_EINO_INTEGRATION=0 OPENAI_API_KEY= go test ./...
```

정적 검증과 Markdown whitespace 확인입니다.

```bash
go vet ./...
git diff --check
```

실제 OpenAI provider 동작은 필요한 chapter에서만 opt-in으로 확인합니다.

```bash
RUN_EINO_INTEGRATION=1 go test ./internal/llm/openai -run TestOpenAIChatModelIntegration -count=1 -v
```

## Chapter별 확인 포인트

| Chapter | 출력에서 볼 것 |
| --- | --- |
| 01 | fake ChatModel 응답과 `model.BaseChatModel` 경계 |
| 02 | system, history, user message 순서 |
| 03 | integration flag가 꺼졌을 때 실제 API를 호출하지 않는 gate |
| 04 | `ToolCall`과 `ToolMessage`, 두 번의 model generate 흐름 |
| 05 | input variables, ChatTemplate output, ChatModel output trace |
| 06 | `selected route`로 calculator branch와 chat branch 분기 |
| 07 | stream chunks와 final answer 조립 결과 |
| 08 | Chain, ChatTemplate, ChatModel callback event timeline |
| 09 | retrieved sources, prompt context summary, final answer 순서 |
| 10 | MCP client가 server subprocess를 띄워 tool call과 resource read 수행 |

## 다음 작업

- Chapter 11에서 ReAct Agent를 다룹니다.
