# Chapter 08 Callback Observability

Callback observability records model and prompt lifecycle events so an application can explain what happened during a run.
In a RAG pipeline, those callback events help show which prompt messages and retrieved context reached the model.
This makes retrieval behavior easier to debug when an answer is missing an expected source.
