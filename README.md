# yadro-test-applied-dev

HTTP-сервис для разбора логов InfiniBand (архив ZIP с файлами `ibdiagnet2.db_csv` и `ibdiagnet2.sharp_an_info`) и выдачи метаданных, топологии и данных по узлам.

## Требования

- Go 1.25.3 или новее (см. `go.mod`)
- Docker и Docker Compose (для варианта с контейнерами)
- PostgreSQL 15 (для локального запуска без Docker)

## Переменные окружения

Скопируйте пример и заполните значения:

```sh
cp .env.example .env
```

Основные переменные: `PORT`, `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_SSLMODE`, `LOG_LEVEL`.

В `docker-compose.yaml` порт приложения на хосте задаётся как `${PORT}:8080`, а процесс внутри контейнера слушает тот же `PORT` из окружения. Чтобы маппинг совпадал с портом процесса, задайте в `.env`, например, `PORT=8080` (тогда сервис будет доступен на `http://localhost:8080`).

## Запуск через Docker Compose

Из корня репозитория:

```sh
docker compose up --build
```

Поднимаются PostgreSQL, одноразовый job миграций (`migrate`) и приложение (`app`). Остановка: `Ctrl+C` или `docker compose down`.

Образ приложения собирает бинарник командой `go build -o /service ./cmd/service` (см. `Dockerfile`, stage `builder`, target `app`).

## Тесты

```sh
go test ./...
```

> Покрытие юнит тестами всего проекта - 18.0%

## API и примеры `curl`

Базовый URL ниже — `http://localhost:8080`; подставьте свой хост и порт.

### Пинг

```sh
curl -sS "http://localhost:8080/api/v1/ping"
```

Ожидаемый ответ: JSON с полем `"message":"pong"`.

### Happy path: разбор архива

Архив должен быть доступен **на файловой системе процесса** сервиса (в Docker это файлы внутри контейнера). В образ копируется каталог `data/` в `/app/data/` (см. `Dockerfile`).

Соберите ZIP из примеров файлов в репозитории и скопируйте его в контейнер приложения:

```sh
(cd data/log && zip -q /tmp/happy.zip ibdiagnet2.db_csv ibdiagnet2.sharp_an_info)
docker cp /tmp/happy.zip "$(docker compose ps -q app):/app/data/happy.zip"
curl -sS -X POST "http://localhost:8080/api/v1/parse/?path=/app/data/happy.zip"
```

Успешный ответ: HTTP 201 и JSON с полем `log_id` (число). Дальше можно запросить, например:

```sh
LOG_ID=1
curl -sS "http://localhost:8080/api/v1/topology/${LOG_ID}"
curl -sS "http://localhost:8080/api/v1/log/${LOG_ID}"
```

Подставьте реальный `log_id` из ответа `parse`.

### Ошибка: невалидный ZIP (`POST /api/v1/parse/`)

```sh
printf 'not a zip file' > /tmp/garbage.zip
docker cp /tmp/garbage.zip "$(docker compose ps -q app):/app/data/garbage.zip"
curl -sS -w "\nHTTP %{http_code}\n" -X POST "http://localhost:8080/api/v1/parse/?path=/app/data/garbage.zip"
```

Ожидается HTTP 400 и тело вида `{"error":"broken zip"}`.

### Ошибка: отсутствует query-параметр `path`

```sh
curl -sS -w "\nHTTP %{http_code}\n" -X POST "http://localhost:8080/api/v1/parse/"
```

Ожидается HTTP 400 и `{"error":"path is required"}`.

### Ошибка: неверный идентификатор в пути (другой эндпоинт)

```sh
curl -sS -w "\nHTTP %{http_code}\n" "http://localhost:8080/api/v1/log/not-a-number"
```

Ожидается HTTP 400 и `{"error":"log_id is required"}`.

### Ошибка: запись не найдена

```sh
curl -sS -w "\nHTTP %{http_code}\n" "http://localhost:8080/api/v1/log/999999999"
```

Ожидается HTTP 404 и `{"error":"log not found"}` (если такого `log_id` нет в базе).

## Маршруты

| Метод | Путь | Назначение |
|--------|------|------------|
| GET | `/api/v1/ping` | Проверка сервиса |
| POST | `/api/v1/parse/?path=...` | Разбор ZIP по пути на диске |
| GET | `/api/v1/topology/{log_id}` | Топология по логу |
| GET | `/api/v1/log/{log_id}` | Метаданные лога |
| GET | `/api/v1/node/{node_id}` | Узел |
| GET | `/api/v1/port/{node_id}` | Порты узла |

## Демонстрация работы сервиса 

> Запуск приложения через `docker compose up --build` (с прогоном миграций)
![](assets/launch_and_check_endpoints.png)

> Получение статус кода 400 при отправке невалидного .zip 
![](assets/invalid_zip_400.png)

> Коллекция Postman
![](assets/postman.png)
