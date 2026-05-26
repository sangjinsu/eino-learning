# Chapter Guide

학습은 작은 chapter 단위로 진행합니다. 각 chapter는 외부 API 없이 통과하는 unit test를 우선합니다.

## Chapter 01. Eino 개요와 ChatModel

목표:

- Eino가 LLM 애플리케이션에서 제공하는 Component 개념을 이해합니다.
- `ChatModel`의 역할을 이해합니다.
- 실제 GPT API 없이 fake ChatModel로 테스트 가능한 구조를 만듭니다.
- 이후 OpenAI 모델로 교체 가능한 interface 기반 흐름을 잡습니다.

완료 기준:

- `internal/fake`에 fake ChatModel이 있습니다.
- `internal/llm/chat`에 fake model을 주입할 수 있는 service가 있습니다.
- `cmd/ch01-chatmodel` 예제가 외부 API 없이 실행됩니다.
- `go test ./...`가 통과합니다.

## Chapter 02. Prompt Template과 Message 설계

목표:

- Eino의 `ChatTemplate`이 변수를 message 목록으로 바꾸는 흐름을 이해합니다.
- system prompt, optional history, user question 순서의 message 설계를 테스트합니다.
- fake ChatModel로 template 결과가 model 입력에 전달되는지 검증합니다.

완료 기준:

- `internal/llm/prompting`에 기본 ChatTemplate이 있습니다.
- `internal/llm/chat.Service`가 `AskWithHistory`로 history를 받을 수 있습니다.
- `cmd/ch02-prompt-template` 예제가 외부 API 없이 실행됩니다.
- `go test ./...`가 통과합니다.

## Chapter 03. OpenAI ChatModel 연동

목표:

- Eino extension의 OpenAI ChatModel을 생성하는 방법을 이해합니다.
- fake ChatModel과 실제 OpenAI ChatModel이 같은 `model.BaseChatModel` 경계로 교체되는 구조를 확인합니다.
- 실제 API 호출을 opt-in integration test로 분리합니다.

완료 기준:

- `internal/llm/openai`에 OpenAI ChatModel factory가 있습니다.
- `.env` 또는 환경 변수 기반 `OPENAI_API_KEY`, `OPENAI_MODEL`, `OPENAI_BASE_URL` config가 있습니다.
- `RUN_EINO_INTEGRATION=1`일 때만 실제 API 호출 test가 실행됩니다.
- `cmd/ch03-openai-chatmodel` 예제가 기본 실행에서는 API를 호출하지 않습니다.
- 기본 `go test ./...`가 외부 API 없이 통과합니다.

## Chapter 04. Tool Calling

목표:

- Eino의 tool metadata와 tool 실행 interface를 이해합니다.
- `ChatModel.WithTools`로 model에게 tool schema를 전달하는 흐름을 확인합니다.
- model이 생성한 `ToolCall`과 tool 결과인 `ToolMessage`의 연결 방식을 확인합니다.
- 위험한 시스템 접근 없이 안전한 calculator tool로 실제 tool 실행을 테스트합니다.

완료 기준:

- `internal/tools`에 실제 계산을 수행하는 safe `calculator` tool이 있습니다.
- Eino `compose.ToolsNode`로 tool call을 실행하는 helper가 있습니다.
- `internal/llm/toolcalling`에 model -> tool -> model final answer loop를 실행하는 service가 있습니다.
- `cmd/ch04-tool-calling` 예제가 `OPENAI_API_KEY`를 읽어 실제 model-backed tool calling을 실행합니다.
- Integration test는 `RUN_EINO_INTEGRATION=1`에서만 실제 OpenAI API를 호출합니다.
- `go test ./...`가 통과합니다.

## Chapter 05. Chain 구성

목표:

- Eino의 `compose.NewChain`이 component를 선형 pipeline으로 묶는 방식을 이해합니다.
- `ChatTemplate -> ChatModel` 흐름을 compiled `Runnable`로 실행합니다.
- 기존 manual service 호출과 Chain 기반 호출의 차이를 테스트로 비교합니다.

완료 기준:

- `internal/llm/chain`에 Chain service가 있습니다.
- Chain은 `map[string]any -> ChatTemplate -> ChatModel -> *schema.Message` 흐름입니다.
- `cmd/ch05-chain` 예제가 `OPENAI_API_KEY`를 읽어 실제 OpenAI ChatModel 기반 Chain을 실행합니다.
- Integration test는 `RUN_EINO_INTEGRATION=1`에서만 실제 OpenAI API를 호출합니다.
- `go test ./...`가 통과합니다.

## Chapter 06. Graph 구성

목표:

- Eino의 `compose.NewGraph`로 named node와 explicit edge를 구성하는 방법을 이해합니다.
- `AddBranch`로 입력에 따라 다른 node path를 선택하는 흐름을 확인합니다.
- calculator branch와 chat model branch를 나눠 Graph가 Chain보다 어울리는 상황을 학습합니다.

완료 기준:

- `internal/llm/graph`에 Graph service가 있습니다.
- Graph는 `route -> calculator` 또는 `route -> ChatTemplate -> ChatModel`로 분기합니다.
- calculator branch는 model을 호출하지 않고 실제 계산을 수행합니다.
- `cmd/ch06-graph` 예제가 `OPENAI_API_KEY`를 읽어 실제 OpenAI ChatModel 기반 Graph를 실행합니다.
- Integration test는 `RUN_EINO_INTEGRATION=1`에서만 실제 OpenAI API를 호출합니다.
- `go test ./...`가 통과합니다.

## Chapter 07. Streaming

목표:

- Eino `ChatModel.Stream`과 `schema.StreamReader` 사용법을 이해합니다.
- `Recv()` loop로 assistant chunk를 읽고 `io.EOF`에서 종료하는 패턴을 확인합니다.
- 같은 `ChatTemplate` prompt를 streaming model 입력으로 전달합니다.

완료 기준:

- `internal/llm/streaming`에 streaming helper가 있습니다.
- `streaming.Service.StreamWithHistory`는 `ChatTemplate -> ChatModel.Stream` 흐름을 실행합니다.
- `streaming.Service.AskWithHistory`는 chunk를 모아 `streaming.Result`를 반환합니다.
- `cmd/ch07-streaming` 예제가 `OPENAI_API_KEY`를 읽어 실제 OpenAI ChatModel stream을 출력합니다.
- Integration test는 `RUN_EINO_INTEGRATION=1`에서만 실제 OpenAI API를 호출합니다.
- `go test ./...`가 통과합니다.

## Chapter 08. Callback과 Observability

목표:

- Eino callback handler가 component lifecycle을 관찰하는 방식을 이해합니다.
- `compose.WithCallbacks`로 실행 단위 observability를 붙입니다.
- `ChatTemplate -> ChatModel` Chain에서 start/end/error event를 확인합니다.

완료 기준:

- `internal/llm/observability`에 `CallbackRecorder`와 observable Chain helper가 있습니다.
- `callbacks.NewHandlerBuilder`로 start/end/error event를 수집합니다.
- 테스트는 `Chain start -> ChatTemplate start/end -> ChatModel start/end -> Chain end` 순서를 검증합니다.
- `cmd/ch08-callback-observability` 예제가 `OPENAI_API_KEY`를 읽어 실제 OpenAI ChatModel 실행 event를 출력합니다.
- CLI 기본 질문과 history는 한국어 예시로 구성합니다.
- Integration test는 `RUN_EINO_INTEGRATION=1`에서만 실제 OpenAI API를 호출합니다.
- `go test ./...`가 통과합니다.

## Chapter 09. RAG 기초

목표:

- Eino `Retriever`가 자연어 질문으로 관련 `schema.Document`를 검색하는 방식을 이해합니다.
- Markdown/Text 예시 문서를 in-memory keyword retriever로 검색합니다.
- 검색 context를 prompt에 삽입해 ChatModel이 source-grounded answer를 만들게 합니다.
- CLI에서 retrieved sources, prompt context summary, final answer를 순서대로 확인합니다.

완료 기준:

- `internal/llm/rag`에 RAG service와 in-memory keyword retriever가 있습니다.
- `cmd/ch09-rag` 예제가 `OPENAI_API_KEY`를 읽어 실제 OpenAI ChatModel 기반 RAG를 실행합니다.
- CLI 기본 질문은 한국어 예시로 구성합니다.
- 예시 문서는 `testdata/docs/ch09-rag`의 Markdown/Text 파일만 사용합니다.
- README와 notes에 `question -> Retriever -> context -> ChatTemplate -> ChatModel -> answer + sources` 흐름 그래프가 있습니다.
- PDF parser, embedding provider, vector store는 v1 범위에서 제외합니다.
- Integration test는 `RUN_EINO_INTEGRATION=1`에서만 실제 OpenAI API를 호출합니다.
- `go test ./...`가 통과합니다.

## Chapter 09 이후 로드맵

- Chapter 10: ReAct Agent
- Chapter 11: GraphTool
- Chapter 12: Human-in-the-loop
- Chapter 13: Mini Project, DevOps Assistant
