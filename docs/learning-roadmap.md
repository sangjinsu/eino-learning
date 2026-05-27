# Learning Roadmap

## 목표

최종 목표는 Go + Eino 기반 DevOps Assistant를 만드는 것입니다. 각 장은 작고 테스트 가능한 예제로 진행합니다.

## 순서

| Chapter | 주제 | 상태 |
| --- | --- | --- |
| 01 | ChatModel fake와 기본 service | 완료 |
| 02 | Prompt Template과 Message 설계 | 완료 |
| 03 | OpenAI ChatModel integration | 완료 |
| 04 | Tool Calling | 완료 |
| 05 | Chain | 완료 |
| 06 | Graph | 완료 |
| 07 | Streaming | 완료 |
| 08 | Callback과 Observability | 완료 |
| 09 | RAG 기초 | 완료 |
| 10 | MCP 기초 | 완료 |
| 11 | ReAct Agent | 예정 |
| 12 | GraphTool | 예정 |
| 13 | Human-in-the-loop | 예정 |
| 14 | Mini DevOps Assistant | 예정 |

## 원칙

- 외부 API 없이 unit test가 통과해야 합니다.
- 실제 provider 연동은 opt-in integration test로 분리합니다.
- 각 장의 대표 실행 방법은 `docs/progress.md`에 남깁니다.
