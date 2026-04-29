defmodule OptimalEngine.Embed.Ollama do
  @moduledoc """
  Thin HTTP wrapper around the Ollama local API.

  Stateless module — no GenServer. All calls are synchronous and return
  `{:ok, result}` or `{:error, reason}`.

  Uses Erlang `:httpc` (via `:inets`) so no additional dependencies are required.
  """

  require Logger

  @availability_ttl_ms 60_000
  @availability_cache_key :ollama_availability_cache

  # ---------------------------------------------------------------------------
  # Public API
  # ---------------------------------------------------------------------------

  @doc """
  Returns `true` if the Ollama daemon is reachable, `false` otherwise.

  Result is cached in the calling process dictionary for 60 seconds.
  """
  @spec available?() :: boolean()
  def available? do
    now = System.monotonic_time(:millisecond)

    case Process.get(@availability_cache_key) do
      {result, cached_at} when now - cached_at < @availability_ttl_ms ->
        result

      _ ->
        result = check_availability()
        Process.put(@availability_cache_key, {result, now})
        result
    end
  end

  @doc """
  Stronger than `available?/0` — actually tries an embed probe and
  verifies a non-empty vector comes back. Cached for 60 s per process
  so callers in a tight retrieval loop don't re-probe every call.

  `/api/tags` can be up while `/api/embed` returns empty for missing
  models or partial loads; `available?/0` only catches the former.
  This variant is what the search layer uses to decide whether to
  spend time on a vector hop.
  """
  @spec embed_healthy?() :: boolean()
  def embed_healthy? do
    now = System.monotonic_time(:millisecond)
    key = :oe_ollama_embed_health

    case Process.get(key) do
      {result, cached_at} when now - cached_at < @availability_ttl_ms ->
        result

      _ ->
        # Bound the probe to ~1 s via a dedicated Task so a hung Ollama
        # can't stall the caller for 5–30 s on first use. A healthy
        # local embedder round-trips in tens of milliseconds; anything
        # slower than 1 s is disqualified from the hot retrieval path.
        task =
          Task.async(fn ->
            case embed_text("probe",
                   model:
                     Application.get_env(:optimal_engine, :ollama, [])[:embed_model] ||
                       "nomic-embed-text"
                 ) do
              {:ok, v} when is_list(v) and v != [] -> true
              _ -> false
            end
          end)

        result =
          case Task.yield(task, 1_000) || Task.shutdown(task, :brutal_kill) do
            {:ok, r} when is_boolean(r) -> r
            _ -> false
          end

        Process.put(key, {result, now})
        result
    end
  end

  @doc """
  Generates a text embedding vector for the given input.

  Returns `{:ok, [float()]}` — a 768-dimension vector from nomic-embed-text.

  ## Options
  - `:model` — override the default embed model
  """
  @spec embed(String.t(), keyword()) :: {:ok, [float()]} | {:error, atom()}
  def embed(text, opts \\ []), do: embed_text(text, opts)

  @doc """
  Embed text using the configured text-embedding model.
  Default model: `nomic-embed-text` (configurable via `:optimal_engine, :ollama, :embed_model`).

  Returns `{:ok, [float()]}` — a single 768-dim vector — or `{:error, reason}`.
  """
  @spec embed_text(String.t(), keyword()) :: {:ok, [float()]} | {:error, atom()}
  def embed_text(text, opts \\ []) when is_binary(text) do
    cfg = config()
    model = Keyword.get(opts, :model, cfg[:embed_model])

    body = %{"model" => model, "input" => text}

    case post_json("/api/embed", body) do
      {:ok, %{"embeddings" => [first | _]}} when is_list(first) ->
        {:ok, first}

      {:ok, response} ->
        Logger.warning("Ollama embed_text: unexpected response shape: #{inspect(response)}")
        {:error, :unexpected_response}

      {:error, reason} ->
        {:error, reason}
    end
  end

  @doc """
  Embed an image using the configured vision-embedding model.
  Default model: `nomic-embed-vision` — aligned with `nomic-embed-text` in
  the same 768-dim space so text queries can retrieve image chunks.

  Accepts either a filesystem path or raw binary bytes. The image is sent
  base64-encoded in the `images` field of Ollama's `/api/embed` endpoint.

  Returns `{:ok, [float()]}` — a single 768-dim vector — or `{:error, reason}`.
  """
  @spec embed_image(String.t() | binary(), keyword()) :: {:ok, [float()]} | {:error, atom()}
  def embed_image(path_or_bytes, opts \\ [])

  def embed_image(path, opts) when is_binary(path) do
    cond do
      File.exists?(path) ->
        case File.read(path) do
          {:ok, bytes} -> embed_image_bytes(bytes, opts)
          {:error, reason} -> {:error, reason}
        end

      true ->
        embed_image_bytes(path, opts)
    end
  end

  defp embed_image_bytes(bytes, opts) do
    cfg = config()
    model = Keyword.get(opts, :model, cfg[:embed_vision_model] || "nomic-embed-vision")
    encoded = Base.encode64(bytes)

    # Ollama multi-modal embed: image goes in `images` array. `input` carries
    # an optional caption hint — empty string is fine for pure vision embed.
    body = %{"model" => model, "input" => "", "images" => [encoded]}

    case post_json("/api/embed", body) do
      {:ok, %{"embeddings" => [first | _]}} when is_list(first) ->
        {:ok, first}

      {:ok, response} ->
        Logger.warning("Ollama embed_image: unexpected response shape: #{inspect(response)}")
        {:error, :unexpected_response}

      {:error, reason} ->
        {:error, reason}
    end
  end

  @doc """
  Generates a text completion for the given prompt.

  Returns `{:ok, String.t()}` with the model's response.

  ## Options
  - `:model` — override the default generate model
  - `:system` — system prompt string to prepend to the request
  """
  @spec generate(String.t(), keyword()) :: {:ok, String.t()} | {:error, atom()}
  def generate(prompt, opts \\ []) do
    cfg = config()
    model = Keyword.get(opts, :model, cfg[:generate_model])

    body =
      %{"model" => model, "prompt" => prompt, "stream" => false}
      |> maybe_add_system(Keyword.get(opts, :system))

    case post_json("/api/generate", body) do
      {:ok, %{"response" => text}} ->
        {:ok, text}

      {:ok, response} ->
        Logger.warning("Ollama generate: unexpected response shape: #{inspect(response)}")
        {:error, :unexpected_response}

      {:error, reason} ->
        {:error, reason}
    end
  end

  # ---------------------------------------------------------------------------
  # Private helpers
  # ---------------------------------------------------------------------------

  defp check_availability do
    case get_json("/api/tags") do
      {:ok, _} -> true
      _ -> false
    end
  rescue
    _ -> false
  catch
    _, _ -> false
  end

  defp maybe_add_system(body, nil), do: body
  defp maybe_add_system(body, system), do: Map.put(body, "system", system)

  defp config do
    # 30s was too generous — it let a slow / degraded Ollama stall the
    # Search GenServer long enough that subsequent searches queued
    # behind a dead call. 5s fails fast and lets us degrade to FTS-only
    # retrieval cleanly.
    Application.get_env(:optimal_engine, :ollama,
      host: "http://localhost:11434",
      embed_model: "nomic-embed-text",
      generate_model: "qwen3:8b",
      timeout_ms: 5_000
    )
  end

  defp post_json(path, body) do
    :inets.start()

    cfg = config()
    url = (cfg[:host] <> path) |> to_charlist()
    timeout_ms = cfg[:timeout_ms] || 30_000

    headers = [{~c"content-type", ~c"application/json"}]
    content_type = ~c"application/json"
    encoded_body = Jason.encode!(body)

    http_opts = [timeout: timeout_ms, connect_timeout: 5_000]

    case :httpc.request(:post, {url, headers, content_type, encoded_body}, http_opts, []) do
      {:ok, {{_version, status, _reason}, _resp_headers, resp_body}} ->
        parse_response(status, resp_body)

      {:error, reason} ->
        Logger.warning("Ollama HTTP POST #{path} failed: #{inspect(reason)}")
        {:error, :ollama_unavailable}
    end
  end

  defp get_json(path) do
    :inets.start()

    cfg = config()
    url = (cfg[:host] <> path) |> to_charlist()
    timeout_ms = cfg[:timeout_ms] || 30_000

    http_opts = [timeout: timeout_ms, connect_timeout: 5_000]

    case :httpc.request(:get, {url, []}, http_opts, []) do
      {:ok, {{_version, status, _reason}, _resp_headers, resp_body}} ->
        parse_response(status, resp_body)

      {:error, reason} ->
        Logger.warning("Ollama HTTP GET #{path} failed: #{inspect(reason)}")
        {:error, :ollama_unavailable}
    end
  end

  defp parse_response(200, body) do
    case Jason.decode(to_string(body)) do
      {:ok, decoded} ->
        {:ok, decoded}

      {:error, reason} ->
        Logger.warning("Ollama: failed to decode JSON response: #{inspect(reason)}")
        {:error, :invalid_json}
    end
  end

  defp parse_response(status, body) do
    Logger.warning("Ollama: non-200 status #{status}, body: #{to_string(body)}")
    {:error, {:http_error, status}}
  end
end
