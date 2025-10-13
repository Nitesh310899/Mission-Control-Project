#!/bin/bash

COMMANDER_URL="http://localhost:8080"

# Submit 5 missions rapidly and record their IDs
echo "Submitting 5 missions..."
mission_ids=()
for i in {1..5}; do
  response=$(curl -s -X POST "$COMMANDER_URL/missions" -H "Content-Type: application/json" -d "{\"payload\":\"mission-$i\"}")
  mission_id=$(echo "$response" | jq -r '.mission_id')
  if [[ "$mission_id" == "null" || -z "$mission_id" ]]; then
    echo "Failed to submit mission $i"
    exit 1
  fi

  echo "Mission $i submitted with ID: $mission_id"
  mission_ids+=("$mission_id")
done

echo
echo "Polling mission statuses until all complete or failed..."

all_done=false
while [[ "$all_done" == "false" ]]; do
  all_done=true
  for id in "${mission_ids[@]}"; do
    status=$(curl -s "$COMMANDER_URL/missions/$id" | jq -r '.status')
    echo "Mission $id status: $status"
    if [[ "$status" != "COMPLETED" && "$status" != "FAILED" ]]; then
      all_done=false
    fi
  done
  echo "-----------------------------------"
  if [[ "$all_done" == "false" ]]; then
    sleep 5
  fi
done

echo "All missions completed or failed."

echo
echo "Waiting 90 seconds to verify token rotation logs in Soldier container..."
echo "(Check logs in another terminal by running: docker-compose logs -f soldier)"

sleep 90

echo
echo "Test completed. Please verify soldier logs show token rotation messages."
