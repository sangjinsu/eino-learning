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

## Chapter 02. Prompt Template과 Message 설계

이번 장의 목표:

- Eino의 `prompt.ChatTemplate`이 변수를 `[]*schema.Message`로 바꾸는 흐름을 이해합니다.
- system prompt, optional chat history, user question 순서의 message 설계를 테스트합니다.
- `ChatService`가 template으로 만든 message를 `ChatModel.Generate`에 전달하게 만듭니다.

핵심 개념:

- `prompt.FromMessages(schema.FString, ...)`는 message template 목록을 만듭니다.
- `schema.MessagesPlaceholder("history", true)`는 history가 없을 때 빈 message 목록으로 처리합니다.
- fake model의 `LastInput`으로 실제 모델에 들어간 role/content 순서를 검증합니다.

실행 명령:

```bash
go run ./cmd/ch02-prompt-template 'How does ChatTemplate work?'
```

테스트 명령:

```bash
go test ./...
```

다음 장에서 할 일:

- Chapter 03에서 OpenAI ChatModel을 `RUN_EINO_INTEGRATION=1` 기반 opt-in integration test로 연동합니다.

## Chapter 03. OpenAI ChatModel 연동

이번 장의 목표:

- Eino extension의 OpenAI ChatModel을 `ChatService`에 주입합니다.
- `.env` 또는 환경 변수의 `OPENAI_API_KEY`, `OPENAI_MODEL`, `OPENAI_BASE_URL`로 provider 설정을 분리합니다.
- 실제 API 호출은 `RUN_EINO_INTEGRATION=1`일 때만 실행되게 만듭니다.

핵심 개념:

- `internal/llm`은 fake model과 OpenAI model 모두를 `model.BaseChatModel`로 다룹니다.
- 기본 모델명은 `.env.example`과 같은 `gpt-4.1-mini`입니다.
- 설정 우선순위는 shell 환경 변수, repo root `.env`, 코드 기본값 순서입니다.
- unit test는 API를 호출하지 않고, integration test만 opt-in으로 실제 OpenAI API를 호출합니다.

`.env` 예시:

```env
OPENAI_API_KEY=your-api-key
OPENAI_MODEL=gpt-4.1-mini
OPENAI_BASE_URL=
RUN_EINO_INTEGRATION=1
```

기본 실행 명령:

```bash
go run ./cmd/ch03-openai-chatmodel 'What does Eino ChatModel do?'
```

integration test:

```bash
go test ./internal/llm -run TestOpenAIChatModelIntegration -count=1 -v
```

외부 API 없이 전체 테스트:

```bash
RUN_EINO_INTEGRATION=0 go test ./...
```
