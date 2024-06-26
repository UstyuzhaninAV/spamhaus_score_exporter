# Spamhaus Exporter

Spamhaus Score Exporter - это Prometheus Exporter, который запрашивает данные из API Spamhaus для указанных доменов и экспортирует результаты в формате метрик Prometheus.

## Описание

Spamhaus Score Exporter собирает следующие метрики для каждого домена:

- `spamhaus_score`: Общий счет Spamhaus.
- `spamhaus_score_dimension`: Счет по различным категориям (human, identity, infra, malware, smtp).

## Примеры использования

### Запуск экспортера

#### Использование go run

```bash
go run exporter.go --web.listen-address=9115
```

#### Сборка и запуск исполняемого файла

```bash
go build -o spamhaus_score_exporter
./spamhaus_score_exporter --web.listen-address=9115
```

### Проверка работы экспортера

Для проверки работы экспортера используйте `curl`, чтобы отправить HTTP-запрос с таргетами:

```bash
curl "http://localhost:9115/probe?target=adtrafico.com&target=adtraffico.com&target=https://affiliates.adtrafico.com/"
```

### Проверка метрик

Для проверки метрик, отправьте GET-запрос на `/metrics`:

```bash
curl http://localhost:9115/metrics
```

### Настройка Prometheus

Добавьте следующий job в ваш `prometheus.yml`:

```yaml
scrape_configs:
  - job_name: 'spamhaus_score_exporter'
    metrics_path: /probe
    params:
      module: [http_2xx]
    static_configs:
      - targets:
        - google.com
				- spamhaius.org
    relabel_configs:
      - source_labels: [__address__]
        target_label: __param_target
      - source_labels: [__param_target]
        target_label: instance
      - target_label: __address__
        replacement: localhost:9115  # адрес вашего экспортера
```

### Системный юнит systemd

Создайте файл с именем `spamhaus_score_exporter.service` в каталоге `/etc/systemd/system/`:

```ini
[Unit]
Description=Spamhaus Exporter
After=network.target

[Service]
Type=simple
ExecStart=/usr/local/bin/spamhaus_score_exporter --web.listen-address=9115
Restart=on-failure

[Install]
WantedBy=multi-user.target
```

Перезапустите systemd, чтобы он прочитал новый unit файл:

```bash
sudo systemctl daemon-reload
```

Запустите сервис и убедитесь, что он работает:

```bash
sudo systemctl start spamhaus_score_exporter
sudo systemctl status spamhaus_score_exporter
```

Настройте автоматический запуск сервиса при загрузке системы:

```bash
sudo systemctl enable spamhaus_score_exporter
```

### Docker

#### Dockerfile

Создайте файл `Dockerfile` в корневом каталоге вашего проекта:

```Dockerfile
# Stage 1: Build the Go application
FROM golang:1.21 AS builder

# Set the working directory
WORKDIR /app

# Copy the Go module files
COPY go.mod go.sum ./

# Download and cache the Go modules
RUN go mod download

# Copy the rest of the application code
COPY . .

# Build the Go application
RUN go build -o spamhaus_score_exporter .

# Stage 2: Run the application
FROM debian:bullseye-slim

# Set the working directory
WORKDIR /app

# Copy the binary from the builder stage
COPY --from=builder /app/spamhaus_score_exporter /usr/local/bin/spamhaus_score_exporter

# Expose the port the app runs on
EXPOSE 9115

# Command to run the application
ENTRYPOINT ["/usr/local/bin/spamhaus_score_exporter"]
CMD ["--web.listen-address=0.0.0.0:9115"]
```

#### Docker Compose

Создайте файл `docker-compose.yml` в корневом каталоге вашего проекта:

```yaml
version: '3.8'

services:
  spamhaus_score_exporter:
    build:
      context: .
      dockerfile: Dockerfile
    ports:
      - "9115:9115"
    environment:
      - WEB_LISTEN_ADDRESS=0.0.0.0:9115
```

#### Сборка и запуск контейнеров

Постройте и запустите контейнеры с помощью Docker Compose:

```bash
docker-compose up --build
```

Остановите контейнеры:

```bash
docker-compose down
```

## Лицензия

Этот проект лицензирован под лицензией MIT. См. файл LICENSE для подробностей.
