#!/bin/bash
echo '{"email":"admin@localhost","password":"changeme"}' > /tmp/login_data.json

echo "=== Test Login ==="
curl -s -w "\nHTTP:%{http_code}\n" \
  -X POST http://localhost:3000/api/auth/login \
  -H "Content-Type: application/json" \
  -d @/tmp/login_data.json

echo ""
echo "=== Test Health ==="
curl -s -o /dev/null -w "HTTP:%{http_code}\n" http://localhost:3000/api/health

echo ""
echo "=== Test Login page ==="
curl -s -o /dev/null -w "HTTP:%{http_code}\n" http://localhost:3000/login
