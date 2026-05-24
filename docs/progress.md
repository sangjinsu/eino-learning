# Progress

## Chapter 01. ChatModel

상태: 완료

완료 기준:

- fake ChatModel 구현
- ChatService 구현
- CLI 예제 추가
- `go test ./...` 통과

## Chapter 02. Prompt Template과 Message 설계

상태: 완료

완료 기준:

- 기본 ChatTemplate 구현
- system prompt, optional history, user question message 순서 테스트
- ChatService의 `AskWithHistory` 구현
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
- tool result를 history에 붙이고 model final answer를 다시 생성하는 `AskWithTools` 구현
- Eino `compose.ToolsNode` 기반 tool 실행 테스트 추가
- OpenAI ChatModel 기반 CLI 예제 추가
- OpenAI tool calling integration test 추가
- 기본 `go test ./...`는 외부 API 없이 통과

## 다음 작업

- Chapter 05에서 Chain 구성을 다룹니다.
