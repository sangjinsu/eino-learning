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

## 다음 작업

- Chapter 03에서 OpenAI ChatModel opt-in integration test를 추가합니다.
