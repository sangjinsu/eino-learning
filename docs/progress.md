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

## 다음 작업

- Chapter 04에서 Tool Calling을 fake 기반으로 먼저 다룹니다.
