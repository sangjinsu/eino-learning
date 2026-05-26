# Chapter 07 Streaming

Streaming returns partial assistant messages through StreamReader instead of waiting for one complete Generate response.
A RAG pipeline can use streaming after retrieval and prompt construction so users see the grounded answer as it is generated.
The retrieved context is still prepared before ChatModel.Stream starts, and the final answer is assembled from stream chunks.

한국어 예시:
Chapter 7 streaming은 RAG에서 검색된 context를 prompt에 넣은 뒤 ChatModel.Stream으로 답변 chunk를 순서대로 받는 흐름과 연결됩니다.
streaming은 retrieval을 대신하지 않고, RAG가 만든 grounded prompt에 대한 응답을 더 빨리 보여주는 출력 방식입니다.
따라서 RAG와 streaming을 함께 쓰면 source 기반 답변을 생성하면서 chunk, final answer, retrieved sources를 함께 관찰할 수 있습니다.
