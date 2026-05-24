# eino-learning

Go와 CloudWeGo Eino를 단계별로 익히는 학습 저장소입니다. 처음에는 외부 LLM API 없이 fake 기반 예제로 구조와 테스트 방법을 먼저 익힙니다.

## Chapter 01. ChatModel

이번 장의 목표:

- Eino의 `ChatModel` 역할을 이해합니다.
- fake ChatModel을 만들어 외부 API 없이 테스트합니다.
- 이후 OpenAI 모델로 교체 가능한 service 경계를 만듭니다.

핵심 개념:

- `internal/fake`는 테스트용 fake model을 제공합니다.
- `internal/llm`은 `model.BaseChatModel`을 받아 질문/응답 흐름을 실행합니다.
- `cmd/ch01-chatmodel`은 fake model을 사용하는 최소 실행 예제입니다.

실행 명령:

```bash
go run ./cmd/ch01-chatmodel "What is Eino?"
```

테스트 명령:

```bash
go test ./...
```

다음 장에서 할 일:

- Chapter 02에서 system/user/assistant message와 prompt template을 구조화합니다.
- OpenAI 실제 연동은 Chapter 03에서 `RUN_EINO_INTEGRATION=1` 기반 integration test로 분리합니다.
