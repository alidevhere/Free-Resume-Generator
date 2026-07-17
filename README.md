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

## Docker (Single Command: Frontend + Backend)

From the project root, run:

```bash
docker compose up --build
```

Then open:

- Frontend: `http://localhost:3000`

In Docker Compose, backend is not published to host ports. APIs are accessible through frontend routes only (`http://localhost:3000/api/*`).

The backend image pre-caches Tectonic bundles during build, so PDF generation works even if the running container has no internet access.

## Docker (Backend Only)

Build the image:

```bash
docker build -t resume-generator ./backend
```

Start the backend:

```bash
docker run --rm -p 8080:8080 resume-generator
```

Run backend only with Docker:

```bash
docker run --rm -p 8080:8080 resume-generator
```
