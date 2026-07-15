## Run locally

```bash
go run . --input resume.json --template templates/enhanced-faang-resume.tex.tmpl --output output
```

## Run with Docker

Build the image:

```bash
docker build -t resume-generator .
```

Start the app:

```bash
docker run --rm -p 8080:8080 resume-generator
```

Or use Docker Compose:

```bash
docker compose up --build
```

Open http://localhost:8080 in your browser.
