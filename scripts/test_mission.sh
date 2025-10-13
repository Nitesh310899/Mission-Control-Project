#!/bin/bash
set -e

API="http://localhost:8080"

echo "Submitting mission..."
MISSION_ID=$(curl -s -X POST -H "Content-Type: application/json" -d '{"payload":"Automated Test Mission"}' $API/missions | jq -r '.mission_id')

if [ -z "$MISSION_ID" ] || [ "$MISSION_ID" == "null" ]; then
  echo "Failed to get mission_id from response"
  exit 1
fi

echo "Mission ID: $MISSION_ID"

for i in {1..30}; do
  STATUS=$(curl -s $API/missions/$MISSION_ID | jq -r '.status')
  echo "[$i] Status: $STATUS"
  if [[ "$STATUS" == "COMPLETED" || "$STATUS" == "FAILED" ]]; then
    echo "Mission finished with status: $STATUS"
    exit 0
  fi
  sleep 1
done

echo "Mission did not complete within expected time"
exit 1
