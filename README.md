# Task Service

REST API для управления задачами в командах: JWT-аутентификация, роли в командах, аудит изменений задач, Redis-кеш, MySQL, rate limiting, structured logging, request id, graceful shutdown и Prometheus-метрики.

## Быстрый старт

```bash
make docker-up
make seed
```

API будет доступен на `http://localhost:8080`
Swagger UI доступен на `http://localhost:8080/swagger/`, OpenAPI-спецификация - на `http://localhost:8080/swagger/openapi.yaml`.

Команда `make seed` создает демо-пользователей, команды `Backend` и `Product`, участников с разными ролями и несколько задач с историей. Команду можно запускать повторно: существующие пользователи, команды и задачи не дублируются.

Тестовые пользователи:

- `owner@example.com` / `password123` - owner команды `Backend`, admin команды `Product`.
- `admin@example.com` / `password123` - admin команды `Backend`, owner команды `Product`.
- `member@example.com` / `password123` - member команды `Backend`.
- `outsider@example.com` / `password123` - пользователь без команд для проверки ограничений доступа.

## Основные эндпоинты

- `POST /api/v1/register` - регистрация.
- `POST /api/v1/login` - логин и получение JWT.
- `POST /api/v1/teams` - создать команду, создатель становится `owner`.
- `GET /api/v1/teams` - список команд пользователя.
- `DELETE /api/v1/teams/{id}` - удалить команду, доступно только `owner`.
- `POST /api/v1/teams/{id}/invite` - добавить пользователя в команду, доступно `owner` и `admin`.
- `POST /api/v1/tasks` - создать задачу в команде.
- `GET /api/v1/tasks?team_id=1&status=todo&assignee_id=5&page_size=20&cursor=...` - список задач с фильтрами и cursor-based пагинацией.
- `PUT /api/v1/tasks/{id}` - обновить задачу.
- `GET /api/v1/tasks/{id}/history` - история изменений задачи.
- `GET /api/v1/reports/team-summary` - JOIN 3+ таблиц и агрегация по командам, где пользователь `owner` или `admin`.
- `GET /api/v1/reports/top-creators` - top-3 создателей задач по управляемым командам через оконную функцию.
- `GET /api/v1/reports/invalid-assignees` - задачи в управляемых командах, где assignee не является участником команды.

## Для локального запуска без Docker:

```bash
make run
make run-worker
```

## Архитектура

Проект разделен по слоям Clean Architecture:

- `internal/domain` - доменные ошибки и базовые domain-типы.
- `internal/domain/models` - доменные модели, разнесенные по отдельным файлам.
- `internal/usecase` - бизнес-сценарии, проверки прав и orchestration; каждый сценарий вынесен в отдельный пакет с собственными client/port интерфейсами.
- `internal/infrastructure/mysql` - создание MySQL connection pool и настройка DB-подключения.
- `internal/infrastructure/redis` - создание Redis client и настройка Redis-подключения.
- `internal/repository/mysql` - реализация repository-слоя для MySQL, включая SQL-запросы, приватные row-модели и маппинг в domain.
- `internal/repository/redis` - реализация repository-слоя для Redis-кеша задач команды с TTL 5 минут и приватными cache DTO.
- `internal/usecase/outbox_dispatch` - обработка outbox-событий, retry transient-ошибок, dead-letter для permanent-ошибок и асинхронная доставка invite-email.
- `internal/usecase/outbox_cleanup` - периодическая очистка обработанных outbox-событий старше retention-периода.
- `internal/adapter/http` - delivery-адаптер: REST API router, Swagger, JWT/rate-limit/router middleware, Prometheus.
- `internal/adapter/http/handlers` - HTTP handlers, разнесенные по endpoint-папкам вида `task_list/handler.go`; общего handler на все ручки нет.
- `internal/adapter/http/request` и `internal/adapter/http/response` - HTTP DTO, mapper-функции, parsing/validation входа и serialization ответа.
- `internal/adapter/email` - инфраструктурный адаптер внешнего email-сервиса с circuit breaker.
- `cmd/api` - HTTP entrypoint.
- `cmd/worker` - background worker entrypoint для outbox dispatcher/cleanup.
- `internal/logger` - настройка `log/slog` logger-а; формат задается через `logging.format` или `LOG_FORMAT`, уровень - через `logging.level` или `LOG_LEVEL`.
