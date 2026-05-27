# Workflow Guide

작업은 장별로 작게 진행합니다. 한 번에 여러 chapter를 구현하지 않고, 각 chapter가 `go test ./...`로 검증되도록 만듭니다.

## 기본 작업 순서

1. 현재 파일 구조를 확인합니다.
2. 관련 chapter 목표와 완료 기준을 확인합니다.
3. 변경 범위를 작게 정합니다.
4. 테스트를 먼저 작성하거나 테스트 가능한 구조를 먼저 설계합니다.
5. fake/mock 기반 구현으로 unit test를 통과시킵니다.
6. `gofmt -w .`, `go test ./...`, 가능하면 `go vet ./...`를 실행합니다.
7. `docs/progress.md`에 실행 방법을, `docs/notes.md`에 배운 점을 짧게 남깁니다.

## 테스트 원칙

- unit test는 외부 API를 호출하지 않습니다.
- 표준 `testing` 패키지를 기본으로 사용합니다.
- assertion library는 꼭 필요할 때만 도입합니다.
- LLM 응답은 비결정적이므로 exact string 비교를 integration test에 남발하지 않습니다.
- fake model을 먼저 만들고 실제 provider는 integration test로 분리합니다.

## 코드 스타일

- `context.Context`는 첫 번째 인자로 둡니다.
- 작은 interface를 선호합니다.
- 에러는 감싸서 반환합니다.
- panic은 테스트나 초기화 실패 외에는 사용하지 않습니다.
- 예제 코드는 학습자가 읽기 쉬운 것을 우선합니다.
