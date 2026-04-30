import { HttpClient } from "./http.js";
import { MemoryClient } from "./memory.js";
import { WikiClient } from "./profile.js";
import { RecallClient } from "./recall.js";
import { RetrievalClient } from "./retrieval.js";
import { SubscriptionClient, SurfaceClient } from "./surface.js";
import type {
  ArchitectureEntry,
  AskOptions,
  AskResult,
  GrepOptions,
  GrepResult,
  OptimalEngineConfig,
  ProfileOptions,
  ProfileResult,
  SearchOptions,
  SearchResult,
  StatusResult,
} from "./types.js";
import { WorkspaceClient } from "./workspace.js";

export class OptimalEngine {
  /** Raw HTTP client — exposed for advanced use; prefer the typed sub-clients. */
  readonly http: HttpClient;

  /** Memory CRUD operations. */
  readonly memory: MemoryClient;

  /** Workspace management. */
  readonly workspaces: WorkspaceClient;

  /** Wiki articles and contradiction detection. */
  readonly wiki: WikiClient;

  /** Subscription management. */
  readonly subscriptions: SubscriptionClient;

  /** Server-sent event surface streams. */
  readonly surface: SurfaceClient;

  /** Recall queries (actions / who / when / where / owns). */
  readonly recall: RecallClient;

  private readonly retrieval: RetrievalClient;
  private readonly defaultWorkspace: string | undefined;

  constructor(config: OptimalEngineConfig = {}) {
    const baseUrl = config.baseUrl ?? "http://localhost:4200";
    this.http = new HttpClient(baseUrl, config.apiKey);
    this.defaultWorkspace = config.workspace;

    this.retrieval = new RetrievalClient(this.http, this.defaultWorkspace);
    this.memory = new MemoryClient(this.http, this.defaultWorkspace);
    this.workspaces = new WorkspaceClient(this.http);
    this.wiki = new WikiClient(this.http, this.defaultWorkspace);
    this.subscriptions = new SubscriptionClient(
      this.http,
      this.defaultWorkspace,
    );
    this.surface = new SurfaceClient(this.http, this.defaultWorkspace);
    this.recall = new RecallClient(this.http, this.defaultWorkspace);
  }

  // ---------------------------------------------------------------------------
  // Top-level retrieval shortcuts (mirror the task's client API spec)
  // ---------------------------------------------------------------------------

  /**
   * Ask the second brain a question. Curated wiki first, hybrid search second.
   * Returns an ACL-scoped, audience-shaped, bandwidth-matched envelope with
   * hot citations.
   */
  ask(query: string, opts: AskOptions = {}): Promise<AskResult> {
    return this.retrieval.ask(query, opts);
  }

  /**
   * Hybrid semantic + keyword search across memory.
   */
  search(query: string, opts: SearchOptions = {}): Promise<SearchResult> {
    return this.retrieval.search(query, opts);
  }

  /**
   * Structured grep across memory with intent, scale, and modality filters.
   */
  grep(query: string, opts: GrepOptions = {}): Promise<GrepResult> {
    return this.retrieval.grep(query, opts);
  }

  /**
   * Get a 4-tier workspace profile: static (ground truth), dynamic (rolling),
   * curated (wiki summary), activity (recent + top entities).
   */
  profile(opts: ProfileOptions = {}): Promise<ProfileResult> {
    return this.retrieval.profile(opts);
  }

  // ---------------------------------------------------------------------------
  // System
  // ---------------------------------------------------------------------------

  status(): Promise<StatusResult> {
    return this.http.get<StatusResult>("/api/status");
  }

  architectures(): Promise<ArchitectureEntry[]> {
    return this.http.get<ArchitectureEntry[]>("/api/architectures");
  }
}
