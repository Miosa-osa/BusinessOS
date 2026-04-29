defmodule OptimalEngine.Pipeline.Embedder do
  @moduledoc """
  Stage 5 of the ingestion pipeline — produces a vector per chunk.

  Dispatches on `chunk.modality`:

    * `:text` / `:code` / `:data` / `:mixed` → text embedder (nomic-embed-text)
    * `:image`                                → vision embedder (nomic-embed-vision)
    * `:audio`                                → whisper-transcript → text embedder
    * `:video`                                → (Phase 10+; falls back to chunk text if present)

  All three modalities land in the same 768-dim aligned space. That's the
  single architectural win of this stage: `SearchEngine` can answer a text
  query with an image or audio chunk because their embeddings live in the
  same coordinate system — no separate indexes, no model-switching logic
  at query time.

  ## Graceful degradation

  If Ollama or whisper.cpp is unreachable, `embed/2` returns
  `{:error, :unavailable}`. `embed_tree/2` collects successes and drops
  failures — the ChunkTree itself is unchanged; chunks without embeddings
  simply aren't retrievable via vector search until the providers come back.

  ## Asset-aware embedding

  For chunks that carry an `asset_ref` (images, audio), this module reads
  the original binary from disk via `OptimalEngine.Assets` (Phase 6+ will
  formalize asset storage; today we read the path on `chunk.asset_ref`).
  When no asset is available, text-based fallback: embed `chunk.text` with
  the text model.
  """

  alias OptimalEngine.Embed.{Ollama, Whisper}
  alias OptimalEngine.Pipeline.Decomposer.{Chunk, ChunkTree}
  alias OptimalEngine.Pipeline.Embedder.Embedding

  require Logger

  @text_model "nomic-embed-text"
  @vision_model "nomic-embed-vision"
  @dim 768

  @doc "Embed a single chunk. Returns `{:ok, %Embedding{}}` or `{:error, reason}`."
  @spec embed(Chunk.t(), keyword()) :: {:ok, Embedding.t()} | {:error, term()}
  def embed(%Chunk{} = chunk, opts \\ []) do
    case dispatch(chunk, opts) do
      {:ok, vector, model, effective_modality} ->
        {:ok,
         Embedding.new(
           chunk_id: chunk.id,
           tenant_id: chunk.tenant_id,
           model: model,
           modality: effective_modality,
           dim: length(vector),
           vector: vector
         )}

      {:error, _} = err ->
        err
    end
  end

  @doc """
  Embed every chunk in the tree. Returns `{:ok, [%Embedding{}], %{errors: [{chunk_id, reason}]}}`
  so callers can see which chunks were skipped and why.
  """
  @spec embed_tree(ChunkTree.t(), keyword()) ::
          {:ok, [Embedding.t()], %{errors: [{String.t(), term()}]}}
  def embed_tree(%ChunkTree{chunks: chunks}, opts \\ []) do
    {embeddings, errors} =
      Enum.reduce(chunks, {[], []}, fn chunk, {acc_emb, acc_err} ->
        case embed(chunk, opts) do
          {:ok, emb} -> {[emb | acc_emb], acc_err}
          {:error, reason} -> {acc_emb, [{chunk.id, reason} | acc_err]}
        end
      end)

    {:ok, Enum.reverse(embeddings), %{errors: Enum.reverse(errors)}}
  end

  # ─── dispatch ────────────────────────────────────────────────────────────

  # Text-ish modalities: plain text embedding
  defp dispatch(%Chunk{modality: m, text: text}, opts)
       when m in [:text, :code, :data, :mixed] and is_binary(text) and text != "" do
    with {:ok, vector} <- Ollama.embed_text(text, opts) do
      {:ok, vector, model_name(opts, :text), m}
    end
  end

  # Image: prefer asset bytes (embed the picture); fall back to OCR text if
  # the chunk has text but no readable asset path.
  defp dispatch(%Chunk{modality: :image, asset_ref: ref} = chunk, opts) when is_binary(ref) do
    # In Phase 5 the asset path convention is inherited from the Parser's
    # asset.path. The chunk carries `asset_ref` = hash; the lookup lives in
    # opts[:asset_paths][ref] for now, a future AssetStore will own this.
    asset_path = get_in(opts, [:asset_paths, ref])

    cond do
      is_binary(asset_path) and File.exists?(asset_path) ->
        case Ollama.embed_image(asset_path, opts) do
          {:ok, vector} -> {:ok, vector, model_name(opts, :image), :image}
          {:error, _} = err -> err
        end

      chunk.text != "" ->
        # Vision unavailable or asset missing — fall back to OCR text embedding.
        fallback_text(chunk, opts, :image)

      true ->
        {:error, :no_embeddable_content}
    end
  end

  # Image with no asset ref — OCR text only.
  defp dispatch(%Chunk{modality: :image} = chunk, opts) do
    fallback_text(chunk, opts, :image)
  end

  # Audio: if a transcript is already on the chunk (Phase 2 whisper path),
  # embed that. Otherwise attempt whisper transcription of the asset.
  defp dispatch(%Chunk{modality: :audio, text: transcript} = chunk, opts)
       when is_binary(transcript) and transcript != "" do
    # Existing transcript → text embed.
    with {:ok, vector} <- Ollama.embed_text(transcript, opts) do
      {:ok, vector, "whisper+" <> model_name(opts, :text), :audio}
    end
  end

  defp dispatch(%Chunk{modality: :audio, asset_ref: ref}, opts) when is_binary(ref) do
    # No transcript yet — ask whisper.
    asset_path = get_in(opts, [:asset_paths, ref])

    cond do
      is_binary(asset_path) and File.exists?(asset_path) ->
        case Whisper.transcribe(asset_path, opts) do
          {:ok, %{text: text}} when text != "" ->
            with {:ok, vector} <- Ollama.embed_text(text, opts) do
              {:ok, vector, "whisper+" <> model_name(opts, :text), :audio}
            end

          _ ->
            {:error, :transcript_unavailable}
        end

      true ->
        {:error, :asset_unavailable}
    end
  end

  defp dispatch(%Chunk{modality: :audio}, _opts), do: {:error, :no_embeddable_content}

  # Video: defer to text fallback for now. Phase 10+ extracts frames.
  defp dispatch(%Chunk{modality: :video} = chunk, opts) do
    fallback_text(chunk, opts, :video)
  end

  # Empty text on a text-modality chunk
  defp dispatch(_chunk, _opts), do: {:error, :empty_content}

  defp fallback_text(%Chunk{text: text}, opts, effective_modality)
       when is_binary(text) and text != "" do
    with {:ok, vector} <- Ollama.embed_text(text, opts) do
      {:ok, vector, model_name(opts, :text), effective_modality}
    end
  end

  defp fallback_text(_chunk, _opts, _modality), do: {:error, :no_embeddable_content}

  defp model_name(opts, :text), do: Keyword.get(opts, :text_model, @text_model)
  defp model_name(opts, :image), do: Keyword.get(opts, :vision_model, @vision_model)

  @doc "The canonical vector dimension for Phase 5 providers (768)."
  @spec dim() :: non_neg_integer()
  def dim, do: @dim
end
