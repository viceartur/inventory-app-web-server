Inventory System Application: Backend.

Prerequisites: Docker, Go, PostgreSQL.

Local development:
```bash
go run .
```

Project Deployment:
```bash
docker compose -f compose.prod.yaml -p inventory-app-prod up -d --build
docker compose -f compose.prod.yaml -p inventory-app-prod down --rmi all
```

Test Server Deployment:
```bash
docker compose -f compose.test.yaml -p inventory-app-test up --build -d
docker compose -f compose.test.yaml -p inventory-app-test down --rmi all
```
