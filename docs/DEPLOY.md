# BusinessOS Deploy Runbook

Ordered, copy-pasteable procedure to ship all three artifacts and verify the live system.
Backend goes to Railway, frontend to Cloudflare Pages, desktop to a macOS DMG.

> Conventions: `$` lines are commands. Do not paste secret values into this file or into
> shell history; pipe them from a file or `openssl rand`. No command here runs automatically.

## 0. Infra reference (no secrets)

| Thing | Value |
|-------|-------|
| Railway workspace | `Roberto Luna's Projects` |
| Railway project | `BusinessOS` (environment `production`) |
| Railway service | `businessos-api` |
| Backend public URL | `https://businessos-api-production.up.railway.app` |
| Backend public host (browser) | `https://businessos.dev/api/*` via the Cloudflare Pages Function proxy |
| Railway database | Railway Postgres service `Postgres` in project `BusinessOS` |
| Frontend (web) | `https://app.businessos.dev` and `https://businessos.dev` (Cloudflare Pages, project `businessos-5`) |
| Session cookie | `Domain=.businessos.dev` (shared across the Pages app and the proxied `/api`) |
| Downloads | Cloudflare R2 bucket `businessos-downloads`, served at `https://downloads.businessos.dev` |
| Backend Dockerfile | `desktop/backend-go/Dockerfile` (binds `SERVER_PORT`, default 8080) |
| Backend CI | `.github/workflows/deploy-backend.yml` (`railway up` after tests pass) |
| Frontend proxy | `frontend/functions/api/[[path]].js` (Pages Function, proxies to Railway) |

### Required backend environment

Configured in the Railway service's **Variables** tab (project `BusinessOS`, service
`businessos-api`). Production `config.Validate()`
(`desktop/backend-go/internal/config/config_helpers.go`) **crashes the service at boot**
if any required value below is missing.

Non-secret variables:

| Var | Value | Why |
|-----|-------|-----|
| `ENVIRONMENT` | `production` | turns on prod validation + secure cookies |
| `SERVER_PORT` | `8080` | backend binds `SERVER_PORT`; must match the Dockerfile `EXPOSE` |
| `COOKIE_DOMAIN` | `.businessos.dev` | session cookie must be shared across the app and `/api` or **login does not stick** |
| `GOOGLE_REDIRECT_URI` | `https://businessos.dev/api/v1/auth/oauth/google/callback` | must match the Google console authorized redirect URI |
| `ALLOWED_ORIGINS` | `https://businessos.dev,https://app.businessos.dev,app://localhost,http://localhost:5173` | prod CORS allowlist (no wildcard, validator rejects `*`) |
| `AI_PROVIDER` | `anthropic` | |
| `ENABLE_LOCAL_MODELS` | `false` | |

Secrets (Railway Variables, masked):

| Secret | Constraint |
|--------|-----------|
| `DATABASE_URL` | provided by the Railway Postgres service; not localhost, no `CHANGE_ME` |
| `GOOGLE_CLIENT_ID` | |
| `GOOGLE_CLIENT_SECRET` | |
| `SECRET_KEY` | >= 32 chars (64 recommended) |
| `ANTHROPIC_API_KEY` | |
| `TOKEN_ENCRYPTION_KEY` | >= 32 chars, REQUIRED in prod |
| `REDIS_KEY_HMAC_SECRET` | >= 32 chars, REQUIRED in prod |
| `INTERNAL_API_SECRET` | >= 32 chars, REQUIRED in prod |
| `WEBHOOK_SIGNING_SECRET` | >= 16 chars, REQUIRED in prod |

Inspect or set them from an authenticated CLI (or the Railway dashboard):

```bash
$ railway variables --service businessos-api --environment production
$ railway variables --service businessos-api --environment production --set KEY=VALUE
```

### Frontend environment

Set in the Cloudflare Pages dashboard (project `businessos-5`):

| Var | Value | Used by |
|-----|-------|---------|
| `CLOUDFLARE_BUILD` | `true` | build command, selects the static adapter |
| `PUBLIC_ENVIRONMENT` | `production` | |
| `BUSINESSOS_BACKEND_URL` | `https://businessos-api-production.up.railway.app` | optional override for the Pages Function; defaults to the Railway URL baked into `[[path]].js` |

The browser always calls the same origin (`/api/*`); the Pages Function proxies to the
backend, so the session cookie stays first-party. `VITE_BACKEND_URL` / `VITE_API_URL`
should be left UNSET for the web build (the runtime resolver returns `/api/v1`).

---

## A. Apply pending migrations to the Railway database

Migrations live in `desktop/backend-go/internal/database/migrations/`. The backend can
auto-apply on boot, but apply explicitly first so a bad migration fails before the new
revision serves traffic.

```bash
# Open a psql shell against the Railway Postgres (proxied by the CLI).
$ railway connect Postgres --environment production
# Then, inside psql, confirm the applied set. Migrations live in
# desktop/backend-go/internal/database/migrations/ and are applied by the
# backend's built-in migrator / auto-apply runner.
```

To run a single file directly instead, read the connection string from the service and
apply it with psql:

```bash
$ export DATABASE_URL="$(railway variables --service businessos-api --environment production --kv | sed -n 's/^DATABASE_URL=//p')"
$ psql "$DATABASE_URL" -v ON_ERROR_STOP=1 -f desktop/backend-go/internal/database/migrations/<n>_<name>.sql
```

Verify the superadmin row exists after migrating:

```bash
$ psql "$DATABASE_URL" -c \
  "select email, role from users where email = 'roberto@businessos.dev';"
# Expect role = superadmin (or platform_admin per 104/105).
```

---

## B. Deploy backend to Railway

Preferred path is CI: push to `main` with changes under `desktop/backend-go/**` and
`.github/workflows/deploy-backend.yml` runs the Go tests then `railway up`.

To deploy manually, stage a copy of the backend outside the git working tree (Railway
prunes uploads with the repository-root `.gitignore`, which drops `cmd/server` from the
backend) and upload it:

```bash
$ STAGE="$(mktemp -d)/businessos-backend"
$ mkdir -p "$STAGE" && cp -R desktop/backend-go/. "$STAGE"/
$ railway up "$STAGE" --path-as-root --no-gitignore --service businessos-api --ci
```

Watch for the build to finish, then confirm the service is serving and reachable:

```bash
$ curl -fsS https://businessos-api-production.up.railway.app/health && echo OK
$ curl -fsS https://businessos-api-production.up.railway.app/health/detailed
# {"components":{"database":{"status":"connected"},...},"status":"healthy"}
```

If the new deployment never becomes healthy, check the build + runtime logs:

```bash
$ railway logs --service businessos-api
```

A config validation panic (missing variable) is the most common boot failure; fix the
value in the Variables tab and redeploy.

---

## C. Build + deploy frontend to Cloudflare Pages

Preferred path is the Cloudflare Pages Git integration (build command
`CLOUDFLARE_BUILD=true npm run build`, output dir `build`, set the env vars above, then
push to the production branch). To deploy manually with Wrangler:

```bash
$ cd frontend
$ CLOUDFLARE_BUILD=true npm run build
# Output dir is build/ (pages_build_output_dir in wrangler.toml).
$ npx wrangler pages deploy build --project-name=businessos-5 --branch=main --commit-dirty=true
```

The project is `businessos-5`, NOT `businessos`. `--branch=main` makes it the production
deployment (serves `businessos.dev` + `app.businessos.dev`).

The `frontend/functions/` directory is uploaded with the static assets and becomes the
Pages Function that proxies `/api/*`. Confirm both the SPA and the proxy:

```bash
$ curl -fsS -o /dev/null -w '%{http_code}\n' https://businessos.dev/            # 200
$ curl -fsS https://businessos.dev/api/v1/health && echo PROXY_OK              # via Function -> Railway
```

---

## D. Build the desktop DMG

Build the frontend (Electron static output) and the Go sidecar binary, then make the DMG.
Run on macOS for a macOS DMG.

```bash
$ cd desktop
$ npm install
$ npm run build:all        # build:frontend (ELECTRON_BUILD=true -> src/renderer)
                           # + build:backend (cross-compiled -> resources/bin/<platform>)
$ npm run make             # electron-forge make -> DMG under out/make/
```

What gets bundled:
- Renderer: the ELECTRON_BUILD SvelteKit static build copied to `desktop/src/renderer`,
  picked up by the Forge Vite plugin (`vite.renderer.config.ts`).
- Backend sidecar: `resources/bin/darwin-arm64/businessos-server` (and `darwin-x64`),
  shipped via `packagerConfig.extraResource` and spawned by `BackendManager`.
- Native modules (`better-sqlite3`, `electron-store`, `node-pty`) copied by the
  `packageAfterCopy` hook and unpacked by `AutoUnpackNativesPlugin`.

The DMG appears in `desktop/out/make/`.

Code signing + notarization run only if `APPLE_ID` (and `APPLE_IDENTITY`, `APPLE_PASSWORD`,
`APPLE_TEAM_ID`) are exported; otherwise the DMG is unsigned (fine for local testing,
Gatekeeper will warn end users).

Upload the built installers to the downloads bucket and update the landing-page links; see
`docs/deployment/BUILD-DESKTOP.md` for the upload step.

---

## E. Smoke test (do this after every deploy)

1. Backend health (Railway):
   ```bash
   $ curl -fsS https://businessos-api-production.up.railway.app/health && echo OK
   $ curl -fsS https://businessos-api-production.up.railway.app/health/detailed
   ```
2. Proxy health (web origin -> Function -> Railway):
   ```bash
   $ curl -fsS https://businessos.dev/api/v1/health && echo OK
   ```
3. Login (browser):
   - Open `https://app.businessos.dev`, sign in (Google or email).
   - After login, DevTools > Application > Cookies: confirm the session cookie has
     `Domain=.businessos.dev`, `Secure`, `HttpOnly`. If the cookie is missing a Domain,
     `COOKIE_DOMAIN` was not set on the backend (see step B / the env table).
   - Confirm OAuth round-trips without a redirect_uri mismatch (that means
     `GOOGLE_REDIRECT_URI` matches the Google console).
4. Superadmin visibility:
   - Log in as `roberto@businessos.dev`. The `/admin` route must be visible.
   - If not, re-check migrations 104/105 applied (step A).
5. Desktop:
   - Open the built app from `desktop/out/make/`. The Go sidecar should start
     (no "Failed to start Go backend" in the console); the app should reach the cloud
     backend and complete login.

---

## Notes carried over from the previous (Cloud Run) stack

- **`GOOGLE_REDIRECT_URI` must use the routed host, not the raw backend host.** It must be
  `https://businessos.dev/api/v1/auth/oauth/google/callback` and that exact URI must be in
  the Google OAuth client's authorized redirect URIs (console-only).
- **`ALLOWED_ORIGINS` must point at `businessos.dev`.** The live web app is
  `businessos.dev` / `app.businessos.dev`.
- **`COOKIE_DOMAIN` must be set.** Unset means the cookie has no Domain and does not
  survive the app <-> `/api` hop, so login silently fails.
- **Migrations are not automatic on the cloud DB.** Apply them explicitly (step A) before
  the new revision serves traffic.
