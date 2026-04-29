defmodule OptimalEngine.Signal.Dispatcher do
  @moduledoc """
  Signal delivery engine — dispatches signals to matched handlers via pluggable adapters.

  The dispatcher integrates with `OptimalEngine.Signal.Router` to find handlers, then delivers
  signals through one of the built-in adapters:

  - `:pid` — sends `{:signal, signal}` to a pid
  - `:named` — sends to a named process
  - `:pubsub` — broadcasts via `OptimalEngine.Signal.PubSub` (built-in) or a custom module
  - `:logger` — logs the signal
  - `:noop` — discards (useful for testing)

  ## Delivery Modes

  - `dispatch/2` — synchronous, returns all results
  - `dispatch_async/2` — async, returns a Task
  - `dispatch_batch/3` — batch with configurable concurrency
  """

  alias OptimalEngine.Signal.Envelope, as: Signal

  @type adapter :: :pid | :named | :pubsub | :logger | :noop | module()
  @type dispatch_result :: {:ok, [term()]} | {:error, term()}

  @doc """
  Adapter behaviour for custom signal delivery.

  Implement this callback to create custom delivery mechanisms.
  """
  @callback deliver(Signal.t(), keyword()) :: :ok | {:error, term()}

  @optional_callbacks [deliver: 2]

  @doc """
  Dispatches a signal synchronously to all matched handlers.

  Handlers can be:
  - Anonymous functions (`fn signal -> ... end`)
  - MFA tuples (`{Module, :function, extra_args}`)

  ## Options

  - `:router` — router to use for matching (default: `OptimalEngine.Signal.Router`)
  - `:adapter` — delivery adapter (default: `:noop`)
  - `:adapter_opts` — options passed to the adapter

  ## Examples

      {:ok, results} = OptimalEngine.Signal.Dispatcher.dispatch(signal, router: my_router)
  """
  @spec dispatch(Signal.t(), keyword()) :: dispatch_result()
  def dispatch(%Signal{} = signal, opts \\ []) do
    router = Keyword.get(opts, :router)

    handlers =
      if router do
        OptimalEngine.Signal.Router.match(router, signal)
        |> Enum.map(fn {handler, _route} -> handler end)
      else
        Keyword.get(opts, :handlers, [])
      end

    results =
      Enum.map(handlers, fn handler ->
        try do
          {:ok, invoke_handler(handler, signal)}
        rescue
          e -> {:error, {handler, e}}
        catch
          kind, reason -> {:error, {handler, {kind, reason}}}
        end
      end)

    # Also deliver via adapter if specified
    adapter = Keyword.get(opts, :adapter)
    adapter_opts = Keyword.get(opts, :adapter_opts, [])

    if adapter && adapter != :noop do
      deliver_via_adapter(adapter, signal, adapter_opts)
    end

    errors = Enum.filter(results, &match?({:error, _}, &1))

    case errors do
      [] -> {:ok, Enum.map(results, fn {:ok, val} -> val end)}
      _ -> {:error, errors}
    end
  end

  @doc """
  Dispatches a signal asynchronously, returning a Task.

  The task resolves to the same result as `dispatch/2`.

  ## Examples

      task = OptimalEngine.Signal.Dispatcher.dispatch_async(signal, router: my_router)
      {:ok, results} = Task.await(task)
  """
  @spec dispatch_async(Signal.t(), keyword()) :: Task.t()
  def dispatch_async(%Signal{} = signal, opts \\ []) do
    Task.async(fn -> dispatch(signal, opts) end)
  end

  @doc """
  Dispatches a batch of signals with configurable concurrency.

  ## Options

  - `:max_concurrency` — max parallel dispatches (default: `System.schedulers_online()`)
  - `:timeout` — per-signal timeout in ms (default: 5000)
  - All other options are passed through to `dispatch/2`

  ## Examples

      results = OptimalEngine.Signal.Dispatcher.dispatch_batch(signals, router: my_router, max_concurrency: 4)
  """
  @spec dispatch_batch([Signal.t()], keyword()) :: [dispatch_result()]
  def dispatch_batch(signals, opts \\ []) when is_list(signals) do
    max_concurrency = Keyword.get(opts, :max_concurrency, System.schedulers_online())
    timeout = Keyword.get(opts, :timeout, 5_000)
    dispatch_opts = Keyword.drop(opts, [:max_concurrency, :timeout])

    signals
    |> Task.async_stream(
      fn signal -> dispatch(signal, dispatch_opts) end,
      max_concurrency: max_concurrency,
      timeout: timeout,
      on_timeout: :kill_task
    )
    |> Enum.map(fn
      {:ok, result} -> result
      {:exit, reason} -> {:error, {:timeout, reason}}
    end)
  end

  # ── Private: Handler Invocation ─────────────────────────────────

  defp invoke_handler(fun, signal) when is_function(fun, 1), do: fun.(signal)

  defp invoke_handler({mod, fun, args}, signal) when is_atom(mod) and is_atom(fun),
    do: apply(mod, fun, [signal | args])

  defp invoke_handler(handler, _signal),
    do: raise(ArgumentError, "invalid handler: #{inspect(handler)}")

  # ── Private: Adapter Delivery ───────────────────────────────────

  defp deliver_via_adapter(:pid, signal, opts) do
    pid = Keyword.fetch!(opts, :pid)
    send(pid, {:signal, signal})
    :ok
  end

  defp deliver_via_adapter(:named, signal, opts) do
    name = Keyword.fetch!(opts, :name)
    send(name, {:signal, signal})
    :ok
  end

  defp deliver_via_adapter(:pubsub, signal, opts) do
    # Use caller-supplied pubsub module; fall back to the built-in ETS broker.
    pubsub = Keyword.get(opts, :pubsub, OptimalEngine.Signal.PubSub)
    topic = Keyword.get(opts, :topic, "signals")
    OptimalEngine.Signal.PubSub.broadcast(pubsub, topic, {:signal, signal})
  end

  defp deliver_via_adapter(:logger, signal, _opts) do
    require Logger

    Logger.info(
      "[OptimalEngine.Signal.Core] #{signal.type} from #{signal.source} — #{inspect(signal.data)}"
    )

    :ok
  end

  defp deliver_via_adapter(:noop, _signal, _opts), do: :ok

  defp deliver_via_adapter(module, signal, opts) when is_atom(module) do
    module.deliver(signal, opts)
  end
end
