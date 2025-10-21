#!/bin/bash

TOKEN="eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjoiOTBmN2Y1MGEtNTIxYy00Y2NkLTkwZWItMzUzMWQxYzExMjJmIiwidXNlcm5hbWUiOiJuZXd1c2VyIiwiZW1haWwiOiJuZXd1c2VyQGV4YW1wbGUuY29tIiwiaXNfYWRtaW4iOmZhbHNlLCJzZXNzaW9uX2lkIjoiNjgzOWVhZTEtMmY1Zi00ODcxLTkxOGMtYzFmYTViNDQ2YWEzIiwic3ViIjoiOTBmN2Y1MGEtNTIxYy00Y2NkLTkwZWItMzUzMWQxYzExMjJmIiwiZXhwIjoxNzU3MDgwMjMxLCJuYmYiOjE3NTcwNzkzMzEsImlhdCI6MTc1NzA3OTMzMSwianRpIjoiZDE5MjQzZGItMDJmNC00MzJiLTkyOTItYzAyYmVjMWNhZTY0In0.0s2zsbs63U-0xvFWhyizqfc7p8oi8IP79zA7uVQ97Qs"
GIST_ID="4849ffcc-9756-4ce7-bbc1-1b7ccdece2be"
PORT=64798

echo "1. Testing get specific gist..."
curl -s http://127.0.0.1:$PORT/api/v1/gists/$GIST_ID \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n2. Testing search endpoint..."  
curl -s "http://127.0.0.1:$PORT/api/v1/search?q=hello" \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n3. Testing user profile..."
curl -s http://127.0.0.1:$PORT/api/v1/user \
  -H "Authorization: Bearer $TOKEN"

echo -e "\n\n4. Testing public gist view (no auth)..."
curl -s http://127.0.0.1:$PORT/g/$GIST_ID

echo -e "\n\n5. Testing web home page..."
curl -s http://127.0.0.1:$PORT/ | head -10