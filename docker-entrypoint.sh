#!/bin/bash
set -e

BACKEND_PORT="${BACKEND_PORT:-:8080}"

FRONTEND_PORT="${FRONTEND_PORT:-3000}"

echo "Starting Go backend on ${BACKEND_PORT}..."
(
  cd /app/backend
  exec resume-generator -server "${BACKEND_PORT}"
) &
BACKEND_PID=$!

# Give the backend a brief moment to bind its port before the frontend
# starts proxying requests to it.
sleep 1

echo "Starting Next.js frontend on port ${FRONTEND_PORT}..."
(
  cd /app/frontend
  exec npx next start -p "${FRONTEND_PORT}"
) &
FRONTEND_PID=$!

# If either process exits/crashes, stop the container so an orchestrator
# (or `docker run --restart`) can notice and react.
wait -n "${BACKEND_PID}" "${FRONTEND_PID}"
EXIT_CODE=$?

echo "One of the processes exited (code ${EXIT_CODE}); shutting down..."
kill "${BACKEND_PID}" "${FRONTEND_PID}" 2>/dev/null || true
wait
exit "${EXIT_CODE}"
