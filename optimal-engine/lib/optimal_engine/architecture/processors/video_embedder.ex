defmodule OptimalEngine.Architecture.Processors.VideoEmbedder do
  @moduledoc """
  Video embedding — stub for a scene-level visual encoder (VideoMAE /
  InternVideo / VideoCLIP family). Phase 15.

  Until the underlying model is wired, `process/3` returns
  `{:error, :not_implemented}` cleanly so `Architecture.Apply.run/3`
  logs + continues.
  """

  @behaviour OptimalEngine.Architecture.Processor

  @impl true
  def id, do: :video_embedder

  @impl true
  def modality, do: :video

  @impl true
  def emits, do: [:embedding, :caption]

  @impl true
  def init(_config), do: {:ok, %{}}

  @impl true
  def process(_field, _value, _state), do: {:error, :not_implemented}
end
