# Learning Roadmap

## 목표

최종 목표는 Go + Eino 기반 DevOps Assistant를 만드는 것입니다. 각 장은 작고 테스트 가능한 예제로 진행합니다.

## 순서

1. ChatModel fake와 기본 service
2. Prompt Template과 Message 설계
3. OpenAI ChatModel integration
4. Tool Calling
5. Chain
6. Graph
7. Streaming
8. Callback과 Observability
9. RAG 기초 (Markdown/Text 기반 in-memory keyword RAG)
10. ReAct Agent
11. GraphTool
12. Human-in-the-loop
13. Mini DevOps Assistant

## 원칙

- 외부 API 없이 unit test가 통과해야 합니다.
- 실제 provider 연동은 opt-in integration test로 분리합니다.
- 각 장은 README 또는 docs에 실행 방법을 남깁니다.
