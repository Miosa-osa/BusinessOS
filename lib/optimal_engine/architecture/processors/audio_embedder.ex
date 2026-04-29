defmodule OptimalEngine.Architecture.Processors.AudioEmbedder do
  @moduledoc """
  Audio embedding — stub. Phase 5's embedder wired whisper.cpp for
  transcription; a true audio embedder (CLAP / AudioCLIP) lands in
  Phase 15 when we add the wav2vec / audio-modality model.

  For now this processor transcribes the audio (if reachable) and
  emits the transcript as a text-embedding-ready output so downstream
  retrieval still has something to match against.
  """

  @behaviour OptimalEngine.Architecture.Processor

  @impl true
  def id, do: :audio_embedder

  @impl true
  def modality, do: :audio

  @impl true
  def emits, do: [:embedding, :transcription]

  @impl true
  def init(_config), do: {:ok, %{}}

  @impl true
  def process(_field, _value, _state) do
    # Phase 15 wiring — return :not_implemented so `Apply` logs + continues.
    {:error, :not_implemented}
  end
end
