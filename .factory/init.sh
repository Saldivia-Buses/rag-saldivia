#!/bin/bash
set -e

# Install Python dependencies
cd /Users/enzo/rag-saldivia
uv sync --quiet

# Install frontend dependencies
cd /Users/enzo/rag-saldivia/services/sda-frontend
npm install --silent

echo "Environment ready."
