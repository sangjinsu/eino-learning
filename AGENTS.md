# AGENTS.md

이 저장소는 Go + CloudWeGo Eino 학습용 저장소입니다. AGENTS.md는 얇은 진입점으로 두고, 세부 규칙은 아래 guide 문서로 분리합니다.

- 목적: 작은 예제와 테스트로 Eino 개념을 단계적으로 익히는 것
- 현재 상태: Chapter 01-11 완료, 다음 주제는 Chapter 12 GraphTool
- 검증 기본값: `RUN_EINO_INTEGRATION=0 OPENAI_API_KEY= go test ./...`, `go vet ./...`, `git diff --check`
- 외부 API 정책: 기본 검증은 API 없이 통과해야 하며, OpenAI 호출은 `RUN_EINO_INTEGRATION=1`으로 opt-in 합니다

상세 지침:

@guides/project.md
@guides/workflow.md
@guides/security.md
@guides/chapters.md
