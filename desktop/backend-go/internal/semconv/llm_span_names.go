package semconv

const (
	// llm_adapter_apply is the span name for "llm.adapter.apply".
	//
	// LLM adapter application — applying a parameter-efficient fine-tuning adapter to customize a base model.
	// Kind: internal
	// Stability: development
	LlmAdapterApplySpan = "llm.adapter.apply"
	// llm_batch_run is the span name for "llm.batch.run".
	//
	// LLM batch inference job — processing multiple requests in a single batch for efficiency.
	// Kind: internal
	// Stability: development
	LlmBatchRunSpan = "llm.batch.run"
	// llm_cache_lookup is the span name for "llm.cache.lookup".
	//
	// LLM response cache lookup — checks if a cached response exists for the given prompt hash.
	// Kind: internal
	// Stability: development
	LlmCacheLookupSpan = "llm.cache.lookup"
	// llm_chain_of_thought is the span name for "llm.chain_of_thought".
	//
	// Executing chain-of-thought reasoning — multi-step LLM inference with intermediate reasoning.
	// Kind: internal
	// Stability: development
	LlmChainOfThoughtSpan = "llm.chain_of_thought"
	// llm_context_compress is the span name for "llm.context.compress".
	//
	// Context compression — reducing token count of context using configured strategy.
	// Kind: internal
	// Stability: development
	LlmContextCompressSpan = "llm.context.compress"
	// llm_context_compress_process is the span name for "llm.context.compress.process".
	//
	// Processing a single context compression operation — validates compression ratio and token savings.
	// Kind: internal
	// Stability: development
	LlmContextCompressProcessSpan = "llm.context.compress.process"
	// llm_context_manage is the span name for "llm.context.manage".
	//
	// Context window management — handles overflow by applying the configured strategy.
	// Kind: internal
	// Stability: development
	LlmContextManageSpan = "llm.context.manage"
	// llm_cost_record is the span name for "llm.cost.record".
	//
	// Recording cost for a completed LLM inference — captures input/output token costs.
	// Kind: internal
	// Stability: development
	LlmCostRecordSpan = "llm.cost.record"
	// llm_distillation_train is the span name for "llm.distillation.train".
	//
	// Knowledge distillation training — transferring knowledge from teacher to student model.
	// Kind: internal
	// Stability: development
	LlmDistillationTrainSpan = "llm.distillation.train"
	// llm_embedding_generate is the span name for "llm.embedding.generate".
	//
	// LLM embedding generation — converting text input into dense vector representations for semantic search or retrieval.
	// Kind: internal
	// Stability: development
	LlmEmbeddingGenerateSpan = "llm.embedding.generate"
	// llm_evaluation is the span name for "llm.evaluation".
	//
	// Evaluating an LLM response quality using a scoring rubric.
	// Kind: internal
	// Stability: development
	LlmEvaluationSpan = "llm.evaluation"
	// llm_few_shot_retrieve is the span name for "llm.few_shot.retrieve".
	//
	// Few-shot example retrieval — selecting and ranking examples for in-context learning.
	// Kind: internal
	// Stability: development
	LlmFewShotRetrieveSpan = "llm.few_shot.retrieve"
	// llm_finetune_run is the span name for "llm.finetune.run".
	//
	// LLM fine-tuning job execution — training a language model on domain-specific data.
	// Kind: internal
	// Stability: development
	LlmFinetuneRunSpan = "llm.finetune.run"
	// llm_function_call_route is the span name for "llm.function_call.route".
	//
	// LLM function call routing — directing a function call from LLM output to the appropriate handler.
	// Kind: internal
	// Stability: development
	LlmFunctionCallRouteSpan = "llm.function_call.route"
	// llm_guardrail_check is the span name for "llm.guardrail.check".
	//
	// Evaluating LLM safety guardrails on a request or response.
	// Kind: internal
	// Stability: development
	LlmGuardrailCheckSpan = "llm.guardrail.check"
	// llm_inference is the span name for "llm.inference".
	//
	// A single LLM inference call — prompt sent, completion received.
	// Kind: client
	// Stability: development
	LlmInferenceSpan = "llm.inference"
	// llm_lora_train is the span name for "llm.lora.train".
	//
	// LoRA fine-tuning run — applies Low-Rank Adaptation to update a pre-trained model efficiently.
	// Kind: internal
	// Stability: development
	LlmLoraTrainSpan = "llm.lora.train"
	// llm_multimodal_process is the span name for "llm.multimodal.process".
	//
	// Multi-modal LLM processing — handling inputs that combine text with images, audio, video, or documents.
	// Kind: internal
	// Stability: development
	LlmMultimodalProcessSpan = "llm.multimodal.process"
	// llm_prompt_render is the span name for "llm.prompt.render".
	//
	// Rendering a prompt template — substituting variables to produce the final LLM request payload.
	// Kind: internal
	// Stability: development
	LlmPromptRenderSpan = "llm.prompt.render"
	// llm_rag_retrieve is the span name for "llm.rag.retrieve".
	//
	// Retrieval-augmented generation retrieval step — fetching relevant documents from a vector store.
	// Kind: internal
	// Stability: development
	LlmRagRetrieveSpan = "llm.rag.retrieve"
	// llm_response_validate is the span name for "llm.response.validate".
	//
	// LLM response validation — checking a model output against a JSON schema or contract for type safety and completeness.
	// Kind: internal
	// Stability: development
	LlmResponseValidateSpan = "llm.response.validate"
	// llm_sampling_configure is the span name for "llm.sampling.configure".
	//
	// Configuration of LLM sampling parameters for a generation request.
	// Kind: internal
	// Stability: development
	LlmSamplingConfigureSpan = "llm.sampling.configure"
	// llm_streaming_complete is the span name for "llm.streaming.complete".
	//
	// Completion of a streaming LLM response — tracks TTFT, throughput, and chunk delivery.
	// Kind: client
	// Stability: development
	LlmStreamingCompleteSpan = "llm.streaming.complete"
	// llm_streaming_start is the span name for "llm.streaming_start".
	//
	// Start of a streaming LLM response — first token received.
	// Kind: client
	// Stability: development
	LlmStreamingStartSpan = "llm.streaming_start"
	// llm_structured_output_generate is the span name for "llm.structured_output.generate".
	//
	// Structured output generation — LLM produces output conforming to a defined schema.
	// Kind: internal
	// Stability: development
	LlmStructuredOutputGenerateSpan = "llm.structured_output.generate"
	// llm_token_budget is the span name for "llm.token.budget".
	//
	// Token budget enforcement for an LLM session — tracks prompt/completion token usage.
	// Kind: internal
	// Stability: development
	LlmTokenBudgetSpan = "llm.token.budget"
	// llm_tool_orchestrate is the span name for "llm.tool.orchestrate".
	//
	// LLM tool orchestration — coordinates multiple tool calls according to a defined strategy.
	// Kind: internal
	// Stability: development
	LlmToolOrchestrateSpan = "llm.tool.orchestrate"
)