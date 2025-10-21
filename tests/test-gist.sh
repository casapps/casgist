#!/bin/bash

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOTBmN2Y1MGEtNTIxYy00Y2NkLTkwZWItMzUzMWQxYzExMjJmIiwidXNlcm5hbWUiOiJuZXd1c2VyIiwiZW1haWwiOiJuZXd1c2VyQGV4YW1wbGUuY29tIiwiaXNfYWRtaW4iOmZhbHNlLCJzZXNzaW9uX2lkIjoiNjgzOWVhZTEtMmY1Zi00ODcxLTkxOGMtYzFmYTViNDQ2YWEzIiwic3ViIjoiOTBmN2Y1MGEtNTIxYy00Y2NkLTkwZWItMzUzMWQxYzExMjJmIiwiZXhwIjoxNzU3MDgwMjMxLCJuYmYiOjE3NTcwNzkzMzEsImlhdCI6MTc1NzA3OTMzMSwianRpIjoiZDE5MjQzZGItMDJmNC00MzJiLTkyOTItYzAyYmVjMWNhZTY0In0.0s2zsbs63U-0xvFWhyizqfc7p8oi8IP79zA7uVQ97Qs"

PORT=64798

echo "Testing gist creation..."
curl -X POST http://127.0.0.1:$PORT/api/v1/gists \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "My First Gist",
    "description": "This is a test gist",
    "visibility": "public",
    "files": [
      {
        "filename": "hello.py",
        "content": "print(\"Hello, World!\")",
        "language": "python"
      }
    ]
  }'

echo -e "\n\nTesting gist list..."
curl -X GET http://127.0.0.1:$PORT/api/v1/gists \
  -H "Authorization: Bearer $TOKEN"