## Architecture

This project now supports a split frontend/backend setup:

- Go backend API (existing service)
- Next.js frontend UI (new app under `frontend/`)

The frontend calls the backend API for resume CRUD and PDF/LaTeX generation.

The backend is API-only and does not serve UI assets.

## Run Backend (Go API)

Start the API server:

```bash
cd backend
go run . -server :8080
```

For CLI generation from a JSON file, you can choose output type:

```bash
cd backend
go run . -input resume.json -output output -format latex
go run . -input resume.json -output output -format pdf
go run . -input resume.json -output output -format both
```

Available API endpoints include:

- `GET /api/resumes`
- `POST /api/resumes`
- `GET /api/resumes/{id}`
- `PUT /api/resumes/{id}`
- `DELETE /api/resumes/{id}`
- `POST /api/resume/pdf`
- `POST /api/resume/latex`

## Run Frontend (Next.js)

Install dependencies:

```bash
cd frontend
npm install
```

Set API base URL only if you need to bypass the frontend proxy.

By default, the frontend calls relative `/api/*` routes and Next.js proxies them to the backend.

```bash
export NEXT_PUBLIC_API_BASE_URL=http://localhost:8080
```

Run the frontend:

```bash
npm run dev
```

Open `http://localhost:3000` in your browser.

## Frontend Scripts

From `frontend/`:

```bash
npm run dev
npm run build
npm run start
npm run lint
npm run typecheck
```

## Desktop App (Tauri)

This repository includes a Tauri desktop shell in `frontend/src-tauri`.

- The frontend is exported as static assets for desktop builds.
- The Go backend is compiled as a bundled sidecar binary and auto-started by Tauri.
- End users install a single desktop app bundle and do not need Docker, Go, or Node installed.

From `frontend/`:

```bash
npm install
npm run tauri:dev
```

To create an installable desktop bundle:

```bash
npm run tauri:build
```

## Desktop Data Storage (SQLite)

When users install the desktop app, resumes are stored in a per-user SQLite database named `resumes.db` under the app data directory for your Tauri app identifier (`com.free.resume.generator`).

Default locations:

- macOS: `~/Library/Application Support/com.free.resume.generator/resumes.db`
- Windows: `%AppData%\\com.free.resume.generator\\resumes.db`
- Linux: `~/.local/share/com.free.resume.generator/resumes.db`

In backend-only mode (without the desktop shell), the default database path is `backend/data/resumes.db` unless overridden with `-db`.

## Docker (Single Image: Frontend + Backend)

The root `Dockerfile` builds a single image containing both the Go backend and the Next.js frontend. The backend listens on an internal port (`8080`) and is not published to the host; all API traffic is proxied through the Next.js server on port `3000`.

### Quick Start (Single Command)

Pull the prebuilt image from ghcr.io and run it with persistent storage in one command (`--pull always` ensures the latest image is fetched before running):

```bash
mkdir -p ./data && docker run -d --name resume-generator --pull always \
  -p 3000:3000 \
  -v "$(pwd)/data:/app/backend/data" \
  ghcr.io/alidevhere/free-resume-generator:latest
```

Then open `http://localhost:3000`. The SQLite database will be created at `./data/resumes.db` on the host and persists across container restarts. See the [Pull the Prebuilt Image](#pull-the-prebuilt-image-from-ghcrio) section below for stop/restart/remove instructions.

### Pull the Prebuilt Image from ghcr.io

A prebuilt multi-arch image (`linux/amd64`, `linux/arm64`) is published to the GitHub Container Registry, so you don't need to build it yourself:

```bash
docker pull ghcr.io/alidevhere/free-resume-generator:latest
```

Run it with a persistent volume so your resumes survive container restarts:

```bash
mkdir -p ./data
docker run -d --name resume-generator \
  -p 3000:3000 \
  -v "$(pwd)/data:/app/backend/data" \
  ghcr.io/alidevhere/free-resume-generator:latest
```

Then open `http://localhost:3000`. The SQLite database will be created at `./data/resumes.db` on the host.

To stop and restart later (data persists via the volume):

```bash
docker stop resume-generator
docker start resume-generator
```

To remove the container when you no longer need it:

```bash
docker rm -f resume-generator
```

### Build the Image Locally

Build the image from the project root:

```bash
docker build -t free-resume-generator .
```

Run it (ephemeral storage — the SQLite database is discarded when the container stops):

```bash
docker run --rm -p 3000:3000 free-resume-generator
```

Then open `http://localhost:3000`. API calls are available at `http://localhost:3000/api/*`.

The image pre-warms the Tectonic cache during build, so PDF generation works even if the running container has no internet access.

### Persisting Data

To keep resumes between runs, mount a host directory to `/app/backend/data`:

```bash
mkdir -p ./data
docker run --rm -p 3000:3000 -v "$(pwd)/data:/app/backend/data" free-resume-generator
```

The SQLite database will be created at `./data/resumes.db` on the host.

### Backend-Only Image

If you only need the API server (no frontend), build the backend image directly:

```bash
docker build -t resume-generator ./backend
docker run --rm -p 8080:8080 resume-generator
```

### Build Individual Images and Run with Docker Compose

The `docker-compose.yml` defines two services — `backend` (image `resume-generator-backend`) and `frontend` (image `resume-generator-frontend`) — that are wired together over the compose network (the frontend talks to the backend at `http://backend:8080`).

You can build either image independently and then bring the stack up with `docker compose`. Compose will reuse the image you built instead of rebuilding it.

**Build only the backend image:**

```bash
docker compose build backend
```

**Build only the frontend image:**

```bash
docker compose build frontend
```

**Build both images:**

```bash
docker compose build
```

**Run the stack (uses the images built above):**

```bash
docker compose up
```

Open `http://localhost:3000` in your browser. The frontend is published on port `3000` and proxies API calls to the backend service.

If you only want to run a single service (e.g. just the frontend, assuming the backend is already running or reachable elsewhere), use:

```bash
docker compose up frontend
```

> Note: the `frontend` service declares `depends_on: [backend]`, so `docker compose up frontend` will also start the backend. To run the frontend against an external backend instead, set `BACKEND_API_URL` in an override file or environment and start only the frontend service with `--no-deps`:
>
> ```bash
> docker compose up --no-deps frontend
> ```

You can also build the images directly with `docker build` (bypassing Compose) and then run them with `docker compose up` — just make sure the image names match those in `docker-compose.yml`:

```bash
docker build -t resume-generator-backend ./backend
docker build -t resume-generator-frontend ./frontend
docker compose up
```
