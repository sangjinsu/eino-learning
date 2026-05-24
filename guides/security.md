# Security Guide

이 저장소는 GPT API를 사용할 수 있지만, 기본 학습 코드는 외부 API 없이 실행 가능해야 합니다.

## 환경 변수

```env
OPENAI_API_KEY=
OPENAI_MODEL=gpt-4.1-mini
OPENAI_BASE_URL=
RUN_EINO_INTEGRATION=0
```

## API Key 규칙

- API Key를 코드, 테스트, fixture에 하드코딩하지 않습니다.
- `.env`는 커밋하지 않고 `.env.example`만 커밋합니다.
- unit test는 GPT API를 호출하지 않습니다.
- GPT API 호출 테스트는 `RUN_EINO_INTEGRATION=1`일 때만 실행합니다.
- `OPENAI_API_KEY`가 없으면 integration test는 skip합니다.
- API 호출 실패 메시지나 로그에 secret이 노출되지 않게 합니다.

## Agent와 Tool 안전 규칙

- Tool은 allowlist 기반으로만 등록합니다.
- 파일 삭제, 배포, DB write, shell 실행은 초반 chapter에서 다루지 않습니다.
- 위험 작업은 human approval chapter 이후에만 다룹니다.
- 실제 Kubernetes, GitHub, ArgoCD, Slack 연동은 fake 구현 후 integration으로 확장합니다.
