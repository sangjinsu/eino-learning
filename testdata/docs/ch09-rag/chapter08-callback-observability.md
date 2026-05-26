# Chapter 08 Callback Observability

Callback observability records model and prompt lifecycle events so an application can explain what happened during a run.
In a RAG pipeline, those callback events help show which prompt messages and retrieved context reached the model.
This makes retrieval behavior easier to debug when an answer is missing an expected source.

한국어 예시:
Chapter 8 callback은 RAG에서 검색된 context가 prompt message로 들어가고 ChatModel까지 전달되는 흐름을 관찰합니다.
callback은 RAG 답변이 어떤 source와 retrieved context에 근거했는지 확인하는 데 도움이 됩니다.
따라서 callback observability는 RAG 실행에서 검색, prompt 생성, model 응답 단계를 디버깅하는 연결 지점입니다.
