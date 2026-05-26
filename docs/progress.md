# Progress

## Chapter 01. ChatModel

상태: 완료

완료 기준:

- fake ChatModel 구현
- `internal/llm/chat.Service` 구현
- CLI 예제 추가
- `go test ./...` 통과

## Chapter 02. Prompt Template과 Message 설계

상태: 완료

완료 기준:

- 기본 ChatTemplate 구현
- system prompt, optional history, user question message 순서 테스트
- `chat.Service`의 `AskWithHistory` 구현
- CLI 예제 추가
- `go test ./...` 통과

## Chapter 03. OpenAI ChatModel 연동

상태: 완료

완료 기준:

- OpenAI ChatModel 설정 로더 구현
- repo root `.env` 자동 로드
- Eino OpenAI ChatModel factory 구현
- `RUN_EINO_INTEGRATION=1` opt-in integration test 추가
- CLI 예제 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## Chapter 04. Tool Calling

상태: 완료

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

완료 기준:

- `callbacks.NewHandlerBuilder` 기반 `CallbackRecorder` 구현
- `compose.WithCallbacks`로 Chain 실행에 callback handler 연결
- `ChatTemplate`, `ChatModel` start/end/error event 수집
- callback이 관찰자 역할로 event timeline을 기록하는지 테스트
- model error가 발생해도 callback event를 확인할 수 있는 테스트 추가
- OpenAI ChatModel 기반 `cmd/ch08-callback-observability` 예제 추가
- OpenAI Callback integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## 다음 작업

- Chapter 09에서 RAG 기초를 다룹니다.
