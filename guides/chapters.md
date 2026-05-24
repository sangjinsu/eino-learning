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
- `internal/llm`에 fake model을 주입할 수 있는 service가 있습니다.
- `cmd/ch01-chatmodel` 예제가 외부 API 없이 실행됩니다.
- `go test ./...`가 통과합니다.

## Chapter 02 이후 로드맵

- Chapter 02: Prompt Template과 Message 설계
- Chapter 03: OpenAI ChatModel 연동, opt-in integration test
- Chapter 04: Tool Calling
- Chapter 05: Chain 구성
- Chapter 06: Graph 구성
- Chapter 07: Streaming
- Chapter 08: Callback과 Observability
- Chapter 09: RAG 기초
- Chapter 10: ReAct Agent
- Chapter 11: GraphTool
- Chapter 12: Human-in-the-loop
- Chapter 13: Mini Project, DevOps Assistant
