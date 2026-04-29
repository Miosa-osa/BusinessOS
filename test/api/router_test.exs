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
end
