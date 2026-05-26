# Progress

## Chapter 01. ChatModel

상태: 완료

대표 실행 예시:

```bash
go run ./cmd/ch01-chatmodel "Eino는 어떤 문제를 해결하나요?"
```

검증 포인트:

- 외부 API 없이 fake ChatModel 응답이 출력됩니다.
- `chat.Service`가 `model.BaseChatModel` 경계로 호출됩니다.

완료 기준:

- fake ChatModel 구현
- `internal/llm/chat.Service` 구현
- CLI 예제 추가
- `go test ./...` 통과

## Chapter 02. Prompt Template과 Message 설계

상태: 완료

대표 실행 예시:

```bash
go run ./cmd/ch02-prompt-template 'Prompt Template은 어떤 메시지를 만드나요?'
```

검증 포인트:

- system, history, user message 순서가 `messages sent to model`에 출력됩니다.
- 현재 질문이 마지막 user message로 들어갑니다.

완료 기준:

- 기본 ChatTemplate 구현
- system prompt, optional history, user question message 순서 테스트
- `chat.Service`의 `AskWithHistory` 구현
- CLI 예제 추가
- `go test ./...` 통과

## Chapter 03. OpenAI ChatModel 연동

상태: 완료

대표 실행 예시:

```bash
go run ./cmd/ch03-openai-chatmodel 'Eino ChatModel은 어떤 역할인가요?'
```

검증 포인트:

- integration flag가 꺼져 있으면 실제 API 호출 없이 설정 안내가 출력됩니다.
- flag와 API key를 설정하면 같은 `chat.Service`가 OpenAI ChatModel을 사용합니다.

완료 기준:

- OpenAI ChatModel 설정 로더 구현
- repo root `.env` 자동 로드
- Eino OpenAI ChatModel factory 구현
- `RUN_EINO_INTEGRATION=1` opt-in integration test 추가
- CLI 예제 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## Chapter 04. Tool Calling

상태: 완료

대표 실행 예시 (`OPENAI_API_KEY` 필요):

```bash
go run ./cmd/ch04-tool-calling '15 * (2 + 6)'
```

API key 없이 확인할 명령:

```bash
RUN_EINO_INTEGRATION=0 go test ./internal/llm/toolcalling ./internal/tools
```

검증 포인트:

- model이 만든 `ToolCall`과 calculator가 반환한 `ToolMessage`를 분리해서 확인합니다.
- final answer는 tool result를 history에 붙인 뒤 다시 생성됩니다.

완료 기준:

- 실제 계산을 수행하는 safe `calculator` invokable tool 구현
- `ChatModel.WithTools`로 model에 `schema.ToolInfo` 전달
- model이 생성한 `schema.ToolCall`을 `schema.ToolMessage`로 실행하는 helper 구현
- tool result를 history에 붙이고 model final answer를 다시 생성하는 `toolcalling.Service.Ask` 구현
- Eino `compose.ToolsNode` 기반 tool 실행 테스트 추가
- OpenAI ChatModel 기반 CLI 예제 추가
- OpenAI tool calling integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## Chapter 05. Chain 구성

상태: 완료

대표 실행 예시 (`OPENAI_API_KEY` 필요):

```bash
go run ./cmd/ch05-chain 'Chain은 Prompt Template과 ChatModel을 어떻게 연결하나요?'
```

API key 없이 확인할 명령:

```bash
go test ./internal/llm/chain -run 'TestService|TestNewService' -count=1
```

검증 포인트:

- input variables, ChatTemplate output, ChatModel output trace가 순서대로 출력됩니다.
- 기존 prompt/model 흐름이 compiled `Runnable`로 실행됩니다.

완료 기준:

- `compose.NewChain`으로 `ChatTemplate -> ChatModel` 선형 pipeline 구현
- compiled `Runnable`을 사용하는 `chain.Service` 구현
- history가 Chain 입력 변수로 전달되는지 테스트
- blank question이 Chain 실행 전에 거부되는지 테스트
- OpenAI ChatModel 기반 `cmd/ch05-chain` 예제 추가
- OpenAI Chain integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## Chapter 06. Graph 구성

상태: 완료

대표 실행 예시 (`OPENAI_API_KEY` 필요):

```bash
go run ./cmd/ch06-graph 'calculate: 9 * (3 + 4)'
go run ./cmd/ch06-graph 'Graph는 Chain과 언제 다르게 쓰나요?'
```

API key 없이 확인할 명령:

```bash
go test ./internal/llm/graph -run 'TestAssistantGraph|TestNewService' -count=1
```

검증 포인트:

- `selected route`로 calculator branch와 chat branch를 구분합니다.
- calculator branch는 CLI config 검증 이후 route가 선택되면 model 호출 없이 종료됩니다.

완료 기준:

- `compose.NewGraph`로 route, calculator, prompt, model node 구성
- `AddBranch`로 calculator branch와 chat model branch 분기 구현
- calculator branch는 model 호출 없이 실제 계산 수행
- chat branch는 `ChatTemplate -> ChatModel` 흐름 실행
- OpenAI ChatModel 기반 `cmd/ch06-graph` 예제 추가
- OpenAI Graph integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## Chapter 07. Streaming

상태: 완료

대표 실행 예시 (`OPENAI_API_KEY` 필요):

```bash
go run ./cmd/ch07-streaming 'Streaming은 Generate와 무엇이 다른가요?'
```

API key 없이 확인할 명령:

```bash
go test ./internal/llm/streaming -run 'TestChatService.*|TestCollectMessageStream' -count=1
```

검증 포인트:

- `stream chunks`가 먼저 출력되고, 마지막에 이어 붙인 `final answer`가 출력됩니다.
- `received chunks`로 Recv loop가 content를 받은 횟수를 확인합니다.

완료 기준:

- `streaming.Service.StreamWithHistory`로 `ChatTemplate -> ChatModel.Stream` 흐름 구현
- `streaming.Service.AskWithHistory`로 stream chunk를 모아 최종 answer 반환
- Chapter 7용 `fake.StreamingChatModel` 추가
- blank question이 stream 호출 전에 거부되는지 테스트
- OpenAI ChatModel 기반 `cmd/ch07-streaming` 예제 추가
- OpenAI Streaming integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## Chapter 08. Callback과 Observability

상태: 완료

대표 실행 예시 (`OPENAI_API_KEY` 필요):

```bash
go run ./cmd/ch08-callback-observability 'ChatTemplate과 ChatModel 실행을 callback으로 어떻게 관찰하나요?'
```

API key 없이 확인할 명령:

```bash
go test ./internal/llm/observability -run 'TestRunObservableChatChain' -count=1
```

검증 포인트:

- `callback events`에서 Chain, ChatTemplate, ChatModel start/end event를 확인합니다.
- callback은 답변 생성 흐름을 바꾸지 않고 관찰 정보만 남깁니다.

완료 기준:

- `callbacks.NewHandlerBuilder` 기반 `CallbackRecorder` 구현
- `compose.WithCallbacks`로 Chain 실행에 callback handler 연결
- `ChatTemplate`, `ChatModel` start/end/error event 수집
- callback이 관찰자 역할로 event timeline을 기록하는지 테스트
- model error가 발생해도 callback event를 확인할 수 있는 테스트 추가
- OpenAI ChatModel 기반 `cmd/ch08-callback-observability` 예제 추가
- OpenAI Callback integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## Chapter 09. RAG 기초

상태: 완료

대표 실행 예시 (`OPENAI_API_KEY` 필요):

```bash
go run ./cmd/ch09-rag 'RAG는 검색된 문서를 어떻게 답변 근거로 사용하나요?'
```

API key 없이 확인할 명령:

```bash
go test ./cmd/ch09-rag ./internal/llm/rag -count=1
```

검증 포인트:

- `retrieved sources`로 검색 근거를 먼저 확인합니다.
- `prompt context summary`로 검색 문서가 prompt context에 들어갔는지 확인합니다.

완료 기준:

- Markdown/Text 예시 문서를 `schema.Document`로 읽는 Chapter 9 CLI 추가
- in-memory keyword retriever로 관련 문서를 검색
- 검색 context를 prompt에 넣고 ChatModel 답변 생성
- CLI 출력 순서가 retrieved sources -> prompt context summary -> final answer
- PDF parser, embedding provider, vector store는 v1 범위에서 제외
- OpenAI RAG integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## 다음 작업

- Chapter 10에서 ReAct Agent를 다룹니다.
