defmodule OptimalEngine.Ollama do
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
  Generates a text embedding vector for the given input.

  Returns `{:ok, [float()]}` — a 768-dimension vector from nomic-embed-text.

  ## Options
  - `:model` — override the default embed model
  """
  @spec embed(String.t(), keyword()) :: {:ok, [float()]} | {:error, atom()}
  def embed(text, opts \\ []) do
    cfg = config()
    model = Keyword.get(opts, :model, cfg[:embed_model])

    body = %{"model" => model, "input" => text}

    case post_json("/api/embed", body) do
      {:ok, %{"embeddings" => [first | _]}} ->
        {:ok, first}

      {:ok, response} ->
        Logger.warning("Ollama embed: unexpected response shape: #{inspect(response)}")
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
    Application.get_env(:optimal_engine, :ollama,
      host: "http://localhost:11434",
      embed_model: "nomic-embed-text",
      generate_model: "qwen3:8b",
      timeout_ms: 30_000
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
