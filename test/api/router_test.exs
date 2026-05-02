defmodule OptimalEngine.API.RouterTest do
  use ExUnit.Case, async: false

  import Plug.Test
  import Plug.Conn

  alias OptimalEngine.API.Router
  alias OptimalEngine.Wiki.{Page, Store}

  @opts Router.init([])

  defp request(method, path, body \\ nil) do
    conn =
      case body do
        nil ->
          conn(method, path)

        b ->
          conn(method, path, Jason.encode!(b)) |> put_req_header("content-type", "application/json")
      end

    Router.call(conn, @opts)
  end

  describe "GET /api/status" do
    test "returns a JSON status payload" do
      conn = request(:get, "/api/status")
      assert conn.status == 200
      assert {:ok, body} = Jason.decode(conn.resp_body)
      assert Map.has_key?(body, "status")
      assert Map.has_key?(body, "ok?")
      assert Map.has_key?(body, "checks")
    end
  end

  describe "GET /api/metrics" do
    test "returns counters + histograms + uptime" do
      conn = request(:get, "/api/metrics")
      assert conn.status == 200
      assert {:ok, body} = Jason.decode(conn.resp_body)
      assert Map.has_key?(body, "counters")
      assert Map.has_key?(body, "histograms")
      assert Map.has_key?(body, "uptime_ms")
    end
  end

  describe "POST /api/rag" do
    test "requires a query in the body" do
      conn = request(:post, "/api/rag", %{})
      assert conn.status == 400
    end

    test "returns a RAG envelope for a valid query" do
      conn =
        request(:post, "/api/rag", %{
          "query" => "nonexistent-rag-api-probe-#{System.unique_integer([:positive])}",
          "format" => "markdown",
          "audience" => "default"
        })

      assert conn.status == 200
      {:ok, body} = Jason.decode(conn.resp_body)
      assert Map.has_key?(body, "source")
      assert Map.has_key?(body, "envelope")
      assert Map.has_key?(body, "trace")
    end
  end

  describe "GET /api/grep" do
    test "returns 400 when q is missing" do
      conn = request(:get, "/api/grep")
      assert conn.status == 400
      {:ok, body} = Jason.decode(conn.resp_body)
      assert body["error"] =~ "q is required"
    end

    test "returns JSON with query, workspace_id, and results array" do
      conn = request(:get, "/api/grep?q=test&workspace=default&limit=5")
      assert conn.status == 200
      {:ok, body} = Jason.decode(conn.resp_body)
      assert Map.has_key?(body, "query")
      assert Map.has_key?(body, "workspace_id")
      assert Map.has_key?(body, "results")
      assert is_list(body["results"])
    end

    test "each result has the full signal trace keys" do
      conn = request(:get, "/api/grep?q=test&limit=3")
      assert conn.status == 200
      {:ok, body} = Jason.decode(conn.resp_body)

      Enum.each(body["results"], fn r ->
        assert Map.has_key?(r, "slug")
        assert Map.has_key?(r, "scale")
        assert Map.has_key?(r, "intent")
        assert Map.has_key?(r, "sn_ratio")
        assert Map.has_key?(r, "modality")
        assert Map.has_key?(r, "snippet")
        assert Map.has_key?(r, "score")
      end)
    end

    test "literal=true still returns well-formed response" do
      conn = request(:get, "/api/grep?q=test&literal=true&limit=3")
      assert conn.status == 200
      {:ok, body} = Jason.decode(conn.resp_body)
      assert is_list(body["results"])
    end

    test "intent and scale params are accepted without error" do
      conn = request(:get, "/api/grep?q=pricing&intent=record_fact&scale=section&limit=5")
      assert conn.status == 200
    end

    test "unknown intent is gracefully tolerated" do
      conn = request(:get, "/api/grep?q=test&intent=totally_bogus&limit=3")
      # Should not crash — engine ignores invalid intent
      assert conn.status == 200
    end
  end

  describe "GET /api/wiki" do
    test "returns a list of wiki pages" do
      conn = request(:get, "/api/wiki?tenant=default")
      assert conn.status == 200
      {:ok, body} = Jason.decode(conn.resp_body)
      assert is_list(body["pages"])
    end
  end

  describe "GET /api/wiki/:slug" do
    test "returns 404 for an unknown slug" do
      conn = request(:get, "/api/wiki/does-not-exist-#{System.unique_integer([:positive])}")
      assert conn.status == 404
    end

    test "returns rendered body for an existing page" do
      suffix = System.unique_integer([:positive])
      slug = "api-wiki-render-#{suffix}"

      :ok =
        Store.put(%Page{
          tenant_id: "default",
          slug: slug,
          audience: "default",
          version: 1,
          frontmatter: %{"slug" => slug},
          body: "## Summary\n\nHello."
        })

      conn = request(:get, "/api/wiki/#{slug}?format=plain")
      assert conn.status == 200
      {:ok, body} = Jason.decode(conn.resp_body)
      assert body["slug"] == slug
      assert body["body"] =~ "Hello"
    end
  end

  # ---------------------------------------------------------------------------
  # API versioning — /v1/ prefix rewrite + response headers + status field
  # ---------------------------------------------------------------------------

  describe "GET /v1/status (version rewrite)" do
    test "returns 200 — same as /api/status" do
      conn = request(:get, "/v1/status")
      assert conn.status == 200
    end

    test "response body has same keys as /api/status" do
      v1_conn = request(:get, "/v1/status")
      api_conn = request(:get, "/api/status")
      {:ok, v1_body} = Jason.decode(v1_conn.resp_body)
      {:ok, api_body} = Jason.decode(api_conn.resp_body)
      assert Enum.sort(Map.keys(v1_body)) == Enum.sort(Map.keys(api_body))
    end

    test "status response includes api_version field set to v1" do
      conn = request(:get, "/v1/status")
      {:ok, body} = Jason.decode(conn.resp_body)
      assert body["api_version"] == "v1"
    end
  end

  describe "GET /api/status (backward compat)" do
    test "still returns 200" do
      conn = request(:get, "/api/status")
      assert conn.status == 200
    end

    test "includes api_version field" do
      conn = request(:get, "/api/status")
      {:ok, body} = Jason.decode(conn.resp_body)
      assert body["api_version"] == "v1"
    end
  end

  describe "POST /v1/rag (version rewrite)" do
    test "returns 400 when query is missing — same as /api/rag" do
      conn = request(:post, "/v1/rag", %{})
      assert conn.status == 400
    end

    test "returns a RAG envelope for a valid query" do
      conn =
        request(:post, "/v1/rag", %{
          "query" => "nonexistent-v1-rag-probe-#{System.unique_integer([:positive])}",
          "format" => "markdown",
          "audience" => "default"
        })

      assert conn.status == 200
      {:ok, body} = Jason.decode(conn.resp_body)
      assert Map.has_key?(body, "source")
      assert Map.has_key?(body, "envelope")
    end
  end

  describe "X-API-Version response header" do
    test "is present on GET /api/status" do
      conn = request(:get, "/api/status")
      assert get_resp_header(conn, "x-api-version") == ["v1"]
    end

    test "is present on GET /v1/status" do
      conn = request(:get, "/v1/status")
      assert get_resp_header(conn, "x-api-version") == ["v1"]
    end
  end
end
