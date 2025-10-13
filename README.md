# Mission Control System

## Overview

A secure command and control system with:

- **Commander Service:** Issues missions and manages tokens.
- **Soldier Service:** Executes missions asynchronously, refreshing tokens.

All communication uses RabbitMQ queues for secure one-way messaging.

---

## Quick Start

### Prerequisites

- Docker & Docker Compose installed
- Ports 5672 & 15672 open

### Run Services

git clone <your-repo-url>
cd <project-folder>
docker-compose up --build -d



### Verify Services

- Check containers: `docker-compose ps`
- Access RabbitMQ UI: [http://localhost:15672](http://localhost:15672) (guest/guest)

---

## Usage

- Submit a mission via Commander API:

curl -X POST http://localhost:8080/missions -H "Content-Type: application/json" -d '{"payload":"mission data"}'


- Soldier auto-fetches tokens and processes missions concurrently.

---

## Testing

Run mission flow and token rotation test:

./scripts/test_missions.sh


View soldier logs for token rotations:

docker-compose logs -f soldier


---

## Stop Services

docker-compose down


---

## Notes

- Soldier obtains and refreshes short-lived tokens from Commander.
- Communication is asynchronous and secure.
- Monitor logs for mission status and token renewals.

---
