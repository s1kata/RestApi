# REST API задач (Go + PostgreSQL)

Сервис управления списком задач: HTTP API на Go, хранение в PostgreSQL. **Продакшен-развёртывание — на локальном Kubernetes-кластере** (Helm). Для разработки на машине разработчика используется Docker Compose (только БД) и прямой запуск бинарника API.

Репозиторий на GitHub: [github.com/s1kata/RestApi](https://github.com/s1kata/RestApi). Сборка и деплой в кластер идут через **GitLab CI** на локальном GitLab (`192.168.56.111`).

## Архитектура

```text
                    ┌─────────────────────────────────────────┐
                    │     Локальный Kubernetes-кластер       │
                    │                                         │
  Клиент ──────────►│  Service rest-api (NodePort :8080)      │
                    │           │                             │
                    │           ▼                             │
                    │  Deployment api (ghcr.io/s1kata/rest-api)│
                    │           │  Database_URL → db:5432   │
                    │           ▼                             │
                    │  Service db (ClusterIP :5432)           │
                    │           │                             │
                    │           ▼                             │
                    │  Deployment db (postgres:15-alpine)     │
                    │           + Secret db-secret            │
                    └─────────────────────────────────────────┘

  Разработка (хост):  docker compose → PostgreSQL :5432
                      go run ./api → API :8080
                      опционально proxy → :8090 → :8080
```

Поток запроса в приложении:

```text
HTTP → handlers → TaskStore (sqlx) → PostgreSQL
```

## Компоненты репозитория

| Компонент | Назначение |
| --- | --- |
| `api/` | Точка входа HTTP-сервера, маршруты, middleware логирования |
| `internal/models/` | Модели `Task`, DTO создания/обновления, `TaskStats` |
| `internal/db/` | Подключение к PostgreSQL, SQL-операции |
| `internal/handlers/` | JSON-ответы и коды HTTP |
| `sql/init.sql/init.sql` | Схема таблицы `tasks` и seed-данные |
| `docker-compose.yml` | Локальный PostgreSQL 15 (не поднимает API) |
| `Dockerfile` | Multi-stage образ API для registry |
| `restapi/` | Helm chart: Deployments/Services/Secret для api и db |
| `.gitlab-ci.yml` | test → build/push образа → `helm upgrade` в кластер |
| `proxy/` | Учебный reverse-proxy на `:8090` к API на `:8080` |

## HTTP API

Базовый путь: `/tasks`. Ответы — JSON, ошибки — `{ "error": "..." }`.

| Метод | Путь | Описание |
| --- | --- | --- |
| `GET` | `/tasks` | Список задач. Query: `completed=true` \| `false` (опционально) |
| `POST` | `/tasks` | Создание задачи (обязателен непустой `title`) |
| `GET` | `/tasks/{id}` | Одна задача |
| `PUT` | `/tasks/{id}` | Частичное обновление (поля через указатели в JSON) |
| `DELETE` | `/tasks/{id}` | Удаление |
| `GET` | `/tasks/stats` | Статистика: total / completed / pending |

Пример задачи:

```json
{
  "id": 1,
  "title": "Изучить Go",
  "description": "…",
  "completed": true,
  "created_at": "2026-09-09T12:00:00Z",
  "updated_at": "2026-09-09T12:00:00Z"
}
```

## Переменные окружения (API)

| Переменная | По умолчанию | Описание |
| --- | --- | --- |
| `Database_URL` | `postgres://taskuser:taskpass@localhost:5432/tasksdb?sslmode=disable` | DSN PostgreSQL |
| `SERVER_PORT` | `8080` | Порт HTTP-сервера |

В Kubernetes DSN задаётся в `restapi/templates/api-deployment.yml`: хост сервиса БД — `db`, учётные данные совпадают с `values.yaml` и Secret `db-secret`.

## Локальная разработка

### 1. PostgreSQL (Docker Compose)

```powershell
cd RestApi
docker compose up -d
```

- Образ: `postgres:15-alpine`, контейнер `tasks-postgres`
- Порт: `5432`
- БД / пользователь / пароль: `tasksdb` / `taskuser` / `taskpass`
- Volume `postgres_data` сохраняет данные между перезапусками
- `sql/init.sql/init.sql` монтируется в `/docker-entrypoint-initdb.d/` и выполняется **только при первом создании** volume

Остановка:

```powershell
docker compose down
docker compose down -v   # с удалением данных (осторожно)
```

### 2. Запуск API на хосте

```powershell
go run ./api
```

Сервер слушает `:8080` (или `SERVER_PORT`).

### 3. Прокси (опционально)

```powershell
go run ./proxy/proxy
```

Прокси слушает `:8090` и пересылает запросы на `http://localhost:8080`. Подробности — в [proxy/README.md](proxy/README.md).

### 4. Сборка и проверки

```powershell
go build ./...
go test ./...
go vet ./...
```

## Docker-образ API

`Dockerfile`: сборка статического бинарника `./api` в `golang:1.24-alpine`, runtime — `alpine:3.22`, пользователь `app`, порт `8080`.

```powershell
docker build -t ghcr.io/s1kata/rest-api:local .
```

## Kubernetes и Helm (локальный кластер)

Chart: каталог `restapi/`, release name по умолчанию в CI — `restapi`, namespace — `default`.

### Ресурсы

| Kubernetes | Имя | Содержимое |
| --- | --- | --- |
| Deployment | `api` | Контейнер `ghcr.io/s1kata/rest-api:<tag>`, env `Database_URL` |
| Service | `rest-api` | **NodePort**, порт `8080` — доступ к API снаружи кластера |
| Deployment | `db` | `postgres:15-alpine`, env из Secret |
| Service | `db` | **ClusterIP**, порт `5432` — только внутри кластера |
| Secret | `db-secret` | `POSTGRES_DB`, `POSTGRES_USER`, `POSTGRES_PASSWORD` |

Параметры по умолчанию — в [restapi/values.yaml](restapi/values.yaml): один replica, образ API из GHCR, тег `latest` (в CI подставляется `$CI_COMMIT_SHORT_SHA`).

### Ручной деплой (если есть `kubectl` и `helm`)

```powershell
cd restapi
helm upgrade --install restapi . --namespace default --set restApi.image.tag=latest
```

Проверка:

```powershell
kubectl get pods,svc
kubectl get svc rest-api
```

NodePort сервиса `rest-api` — точка входа в API на узлах кластера.

## GitLab CI/CD

Файл [.gitlab-ci.yml](.gitlab-ci.yml), runner-теги: `test`, `build`.

| Stage | Действие |
| --- | --- |
| `test` | `go test ./...` в `golang:1.24-alpine` |
| `build` | `docker build`, push в `ghcr.io/s1kata/rest-api` (`latest` + SHA коммита). Нужен секрет `GHCR_TOKEN` |
| `deploy` | `helm upgrade --install restapi ./restapi --set restApi.image.tag=$CI_COMMIT_SHORT_SHA` |

Remote GitLab в git: `gitlab` → `git@192.168.56.111:gitlab/git-lab-ci.git`.

## Структура каталогов

```text
RestApi/
├── api/main.go
├── internal/
│   ├── models/task.go
│   ├── db/database.go, tasks.go
│   └── handlers/handlers.go
├── sql/init.sql/init.sql
├── docker-compose.yml
├── Dockerfile
├── restapi/                 # Helm chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
├── proxy/                   # учебный прокси
├── .gitlab-ci.yml
├── go.mod
└── README.md
```

## Зависимости Go

- [github.com/jmoiron/sqlx](https://github.com/jmoiron/sqlx) — работа с SQL
- [github.com/lib/pq](https://github.com/lib/pq) — драйвер PostgreSQL

Модуль: `restapi-tasks`, Go 1.24.
