defmodule OptimalEngine.Pipeline.Parser.Video do
  @moduledoc """
  Video parser. If `ffmpeg` is available, extracts:

    * a single key-frame JPEG (via `ffmpeg -i <path> -frames:v 1 -q:v 3 …`)
      — handed to the Image backend for OCR + kept as an Asset
    * the audio track — handed to the Audio backend for transcription

  If `ffmpeg` isn't on `PATH`, returns a metadata-only ParsedDoc with the
  video preserved as an Asset. Phase 5 will add frame-sequence embedding
  for richer video retrieval.
  """

  @behaviour OptimalEngine.Pipeline.Parser.Backend

  alias OptimalEngine.Pipeline.Parser.{Asset, Audio, Image, ParsedDoc}

  @impl true
  def parse(path, opts) when is_binary(path) do
    video_asset =
      case Asset.from_path(path, modality: :video) do
        {:ok, a} -> a
        _ -> nil
      end

    case System.find_executable("ffmpeg") do
      nil ->
        {:ok,
         ParsedDoc.new(
           path: path,
           text: "",
           modality: :video,
           metadata: %{format: Path.extname(path) |> String.trim_leading(".") |> String.downcase()},
           assets: if(video_asset, do: [video_asset], else: []),
           warnings: ["ffmpeg not on PATH — install for frame + audio extraction"]
         )}

      _ ->
        extract_with_ffmpeg(path, video_asset, opts)
    end
  end

  @impl true
  def parse_text(_text, _opts), do: {:error, :binary_format_requires_path}

  defp extract_with_ffmpeg(path, video_asset, opts) do
    tmp_dir = System.tmp_dir!() |> Path.join("optimal_video_#{System.unique_integer([:positive])}")
    File.mkdir_p!(tmp_dir)

    try do
      frame_path = Path.join(tmp_dir, "frame.jpg")
      audio_path = Path.join(tmp_dir, "audio.wav")

      _ =
        System.cmd(
          "ffmpeg",
          ["-y", "-i", path, "-frames:v", "1", "-q:v", "3", frame_path],
          stderr_to_stdout: true
        )

      _ =
        System.cmd(
          "ffmpeg",
          ["-y", "-i", path, "-vn", "-acodec", "pcm_s16le", "-ar", "16000", "-ac", "1", audio_path],
          stderr_to_stdout: true
        )

      {frame_text, frame_assets, frame_warnings} = maybe_parse(Image, frame_path, opts)
      {audio_text, audio_assets, audio_warnings} = maybe_parse(Audio, audio_path, opts)

      combined_text =
        [frame_text, audio_text]
        |> Enum.reject(&(&1 == "" or is_nil(&1)))
        |> Enum.join("\n\n")

      assets =
        [video_asset | frame_assets ++ audio_assets]
        |> Enum.reject(&is_nil/1)

      {:ok,
       ParsedDoc.new(
         path: path,
         text: combined_text,
         structure: [],
         modality: :video,
         metadata: %{
           format: Path.extname(path) |> String.trim_leading(".") |> String.downcase(),
           has_frame: File.exists?(frame_path),
           has_audio: File.exists?(audio_path)
         },
         assets: assets,
         warnings: frame_warnings ++ audio_warnings
       )}
    after
      File.rm_rf!(tmp_dir)
    end
  end

  defp maybe_parse(_backend, path, _opts) when not is_binary(path), do: {"", [], []}

  defp maybe_parse(backend, path, opts) do
    if File.exists?(path) do
      case backend.parse(path, opts) do
        {:ok, doc} -> {doc.text, doc.assets, doc.warnings}
        _ -> {"", [], []}
      end
    else
      {"", [], []}
    end
  end
end
