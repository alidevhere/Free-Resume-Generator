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
