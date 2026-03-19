#!/bin/bash
set -e

MAX_RETRIES=5
SLEEP=10

echo "[health_check] Checking gateway (port 9000)..."
for i in $(seq 1 $MAX_RETRIES); do
  if curl -sf --max-time 5 http://localhost:9000/health > /dev/null 2>&1; then
    echo "[health_check] ✓ Gateway OK"
    break
  fi
  if [ $i -eq $MAX_RETRIES ]; then
    echo "[health_check] ✗ Gateway health check failed after ${MAX_RETRIES} attempts"
    exit 1
  fi
  echo "[health_check] Attempt $i/$MAX_RETRIES — waiting ${SLEEP}s..."
  sleep $SLEEP
done

echo "[health_check] Checking RAG server (port 8081)..."
if ! curl -sf --max-time 5 http://localhost:8081/health > /dev/null 2>&1; then
  echo "[health_check] ✗ RAG server not responding"
  exit 1
fi
echo "[health_check] ✓ RAG server OK"

echo "[health_check] Checking frontend (port 3000)..."
if ! curl -sf --max-time 5 http://localhost:3000/ > /dev/null 2>&1; then
  echo "[health_check] ✗ Frontend not responding"
  exit 1
fi
echo "[health_check] ✓ Frontend OK"

echo "[health_check] All services healthy ✓"
