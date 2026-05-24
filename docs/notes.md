# Notes

## Chapter 01

- fake ChatModel은 외부 API 없이 Eino의 `ChatModel` 형태를 이해하기 위한 학습용 구현입니다.
- `ChatService`는 `model.BaseChatModel`에만 의존하므로 나중에 OpenAI ChatModel로 교체할 수 있습니다.
- `OPENAI_API_KEY`는 Chapter 03 integration test에서만 사용합니다.

## Chapter 02

- `prompt.ChatTemplate`은 변수 map을 받아 `[]*schema.Message`를 생성합니다.
- 이번 장의 기본 template은 system prompt, optional history, user question 순서로 메시지를 만듭니다.
- `schema.MessagesPlaceholder("history", true)`를 사용하면 history가 없을 때도 같은 template을 재사용할 수 있습니다.
- `ChatService`는 template이 만든 메시지 목록을 `ChatModel.Generate`에 전달합니다.
