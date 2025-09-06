# l0-wb-techno-school-go

A demo microservice in Go that receives order data from Kafka, stores it in PostgreSQL, and caches it in memory for fast delivery via an HTTP API and a simple web page.

## What the service does
- Subscribes to a Kafka topic and processes JSON messages with the order model.
- Validates and stores orders in PostgreSQL using transactions.
- Caches recent orders in memory (sync.Map) to speed up repeated requests.
- Restores the cache from the database on startup.
- Returns an order by `order_uid` via a JSON HTTP API and a simple HTML page.

## Quick start and verification
1. Copy configuration templates:
```bash
cp .env.example .env
cp .config.yml.example config.yml
```

2. Go to the `deploy` directory and build/run Docker Compose:
```bash
cd deploy
docker compose build --no-cache
docker compose up -d
```

3. After startup complete, check running containers:
```bash
docker ps
```

4. HTTP checks:
```bash
curl -s http://host:port/health
# curl -s http://localhost:8080/health
curl -s http://host:port/static/index.html
# curl -s http://localhost:8080/static/index.html
```

5. Send a test message to Kafka (example — run inside Kafka container):
```bash
# host = localhost
# port = 9092
docker exec -it kafka sh -c 'echo "{\"order_uid\": \"b563feb7b2b84b6test\", \"track_number\": \"WBILMTESTTRACK\", \"entry\": \"WBIL\", \"delivery\": {\"name\": \"Test Testov\", \"phone\": \"+9720000000\", \"zip\": \"2639809\", \"city\": \"Kiryat Mozkin\", \"address\": \"Ploshad Mira 15\", \"region\": \"Kraiot\", \"email\": \"test@gmail.com\"}, \"payment\": {\"transaction\": \"b563feb7b2b84b6test\", \"request_id\": \"\", \"currency\": \"USD\", \"provider\": \"wbpay\", \"amount\": 1817, \"payment_dt\": 1637907727, \"bank\": \"alpha\", \"delivery_cost\": 1500, \"goods_total\": 317, \"custom_fee\": 0}, \"items\": [{\"chrt_id\": 9934930, \"track_number\": \"WBILMTESTTRACK\", \"price\": 453, \"rid\": \"ab4219087a764ae0btest\", \"name\": \"Mascaras\", \"sale\": 30, \"size\": \"0\", \"total_price\": 317, \"nm_id\": 2389212, \"brand\": \"Vivienne Sabo\", \"status\": 202}], \"locale\": \"en\", \"internal_signature\": \"\", \"customer_id\": \"test\", \"delivery_service\": \"meest\", \"shardkey\": \"9\", \"sm_id\": 99, \"date_created\": \"2021-11-26T06:22:19Z\", \"oof_shard\": \"1\"}" | kafka-console-producer --broker-list host:port --topic orders'
```

6. Check server logs:
```bash
docker compose logs server
```

7. Check cache / API:
```bash
curl -s http://host:port/order/b563feb7b2b84b6test | jq . 2>/dev/null || curl -s http://host:port/order/b563feb7b2b84b6test
# curl -s http://localhost:8080/order/b563feb7b2b84b6test | jq . 2>/dev/null || curl -s http://localhost:8080/order/b563feb7b2b84b6test
```

8. Check data in PostgreSQL (example):
```bash
docker exec -it postgres psql -U postgres -d postgres -c "SELECT * FROM orders WHERE order_uid = 'b563feb7b2b84b6test';"
docker exec -it postgres psql -U postgres -d postgres -c "SELECT * FROM delivery WHERE order_uid = 'b563feb7b2b84b6test';"
docker exec -it postgres psql -U postgres -d postgres -c "SELECT * FROM payment WHERE order_uid = 'b563feb7b2b84b6test';"
docker exec -it postgres psql -U postgres -d postgres -c "SELECT * FROM items WHERE order_uid = 'b563feb7b2b84b6test';"
```

9. Restart the server and verify cache restoration:
```bash
docker compose restart server
docker compose logs server
```

10. Stop:
```bash
docker compose stop
```

## Architecture and code layout (packages — tree)
- cmd/
  - service/             # application entry point
- configs/               # YAML configuration files (non-critical)
- containers/            # Docker images / Dockerfiles
  - docker/
- deploy/                # docker-compose, .env examples, deployment scripts
- docs/
  - db/                  # database documentation
  - http/                # HTTP endpoints (endpoints.http)
  - scheme/              # schema
  - video/               # demo links
- internal/
  - config/              # configuration loading and structure (Viper)
  - di/                  # dependency injection and component assembly
  - domain/              # domain models (structures)
  - ports/               # interfaces (application ports)
  - app/
    - order/             # pure business logic and validation
  - adapters/
    - cache/             # in-memory cache
    - db/
      - postgres/        # connection, repository, migrations
        - connect/
        - migration/
    - kafka/             # Kafka consumer adapter (segmentio/kafka-go)
    - server/            # HTTP server (Fiber v3) and handlers
  - logger/              # zap wrapper
  - shutdown/            # graceful shutdown helper
- migrations/            # SQL migrations
- static/                # static page (index.html)
- pkg/                   # utilities / reusable code

## Endpoints
Available endpoints: [docs/http/endpoints.http](docs/http/endpoints.http)
- GET /health — health check
- GET /static/index.html — frontend
- GET /order/{order_uid} — get order by UID (JSON)

## Used libraries (with versions)
- github.com/gofiber/fiber/v3 v3.0.0-rc.1
- github.com/golang-migrate/migrate/v4 v4.19.0
- github.com/jackc/pgx/v5 v5.7.5
- github.com/segmentio/kafka-go v0.4.49
- github.com/spf13/viper v1.20.1
- go.uber.org/zap v1.27.0
- golang.org/x/sync v0.16.0

## Go version
The project targets Go 1.25.0.
## Go version
The project targets Go 1.25.0.
