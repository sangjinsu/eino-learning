# Project Guide

`eino-learning`은 Go와 CloudWeGo Eino를 단계별로 익히는 학습 저장소입니다. Codex는 코드를 대신 대량 생성하는 도구가 아니라, 학습자가 Eino 개념을 작은 실습과 테스트로 이해하도록 돕는 학습 파트너로 행동합니다.

## 기본 정보

- Module path: `github.com/sangjinsu/eino-learning`
- Go: 1.22 이상
- Eino: `github.com/cloudwego/eino`
- OpenAI model 연동은 Chapter 03 이후에 분리합니다.
- 기본 GPT 모델명: `gpt-4.1-mini`

## 우선순위

1. 학습 가능성
2. 테스트 가능성
3. 단순한 구조
4. Go idiom
5. Eino 개념 이해
6. 확장성
7. 운영 안전성

처음에는 프로덕션 수준의 추상화보다 작고 명확한 예제를 우선합니다.

## 권장 구조

```text
cmd/
  ch01-chatmodel/
internal/
  fake/
  llm/
  tools/
  rag/
  agent/
  observability/
docs/
  learning-roadmap.md
  progress.md
  notes.md
testdata/
  docs/
  golden/
```
