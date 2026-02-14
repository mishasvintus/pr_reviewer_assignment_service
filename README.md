# PR Reviewer Assignment Service

Микросервис для автоматического назначения ревьюеров на Pull Request'ы: управление командами, пользователями и правилами выбора ревьюеров. REST API, состояние в PostgreSQL.

---

## Стек

- **Go 1.24**
- **HTTP:** [Gin](https://github.com/gin-gonic/gin) 
- **DB:** [lib/pq](https://github.com/lib/pq)
- **Config:** [godotenv](https://github.com/joho/godotenv) 
- **PostgreSQL** 
- **Docker / Docker Compose** 

---

## Возможности

- **Команды и пользователи** — создание команд с участниками, флаг активности пользователя (`is_active`). Пользователь с `is_active = false` не назначается ревьюером.
- **PR с привязкой к команде** — при создании PR сохраняется команда автора (`team_name`). Все последующие действия (переназначение, добор ревьюеров) идут **из этой команды**, а не из текущей команды автора/ревьюера.
- **Назначение ревьюеров** — при создании PR автоматически назначается до 2 активных ревьюеров из команды автора (автор исключается). Выбор случайный (crypto/rand).
- **Переназначение** — замена одного ревьюера на другого из **команды PR**; автор и текущие ревьюеры исключаются.
- **Добор ревьюеров** — если у PR меньше 2 ревьюеров, сервис может доназначить кандидатов из команды PR (используется при деактивации команды).
- **Деактивация команды** — массовое отключение пользователей команды; у открытых PR ревьюеры из этой команды снимаются и при необходимости заменяются на участников команды PR.
- **MERGED** — после перевода PR в статус MERGED изменения ревьюеров запрещены. Операция merge идемпотентна.
- **Статистика** — общая сводка и разбивка по ревьюерам и авторам.

---

## Запуск

### Всё в Docker (сервис + БД)

```bash
make docker-up
# или
docker-compose up -d
```

Сервис: **http://localhost:8080**

### Локально (БД в Docker)

```bash
cp .env.example .env
docker-compose up -d postgres
make run
```

---

## Переменные окружения

| Переменная    | Описание           |
|---------------|--------------------|
| `SERVER_HOST` | Хост HTTP-сервера  |
| `SERVER_PORT` | Порт (по умолчанию 8080) |
| `DB_HOST`     | Хост PostgreSQL    |
| `DB_PORT`     | Порт PostgreSQL    |
| `DB_USER`     | Пользователь БД    |
| `DB_PASSWORD` | Пароль БД          |
| `DB_NAME`     | Имя базы           |
| `DB_SSLMODE`  | Режим SSL (например `disable`) |

Пример: см. `.env.example`.

---

## API

| Метод | Путь | Описание |
|-------|------|----------|
| POST | `/team/add` | Создать команду с участниками |
| GET  | `/team/get?team_name=...` | Получить команду |
| POST | `/team/deactivate` | Деактивировать команду |
| POST | `/users/setIsActive` | Установить активность пользователя |
| GET  | `/users/getReview?user_id=...` | Список PR, где пользователь ревьюер |
| POST | `/pullRequest/create` | Создать PR и назначить ревьюеров |
| POST | `/pullRequest/merge` | Перевести PR в MERGED |
| POST | `/pullRequest/reassign` | Переназначить ревьюера |
| GET  | `/stats` | Статистика |

Полная спецификация: **openapi.yml**.

---

## Тестирование

```bash
make test-unit          # Unit-тесты (handlers, domain)
make test-integration   # Интеграционные тесты (нужен PostgreSQL)
make test-all           # Все тесты
make test-coverage      # Покрытие + HTML-отчёт (coverage.html)
make generate-mocks     # Регенерация моков (mockery)
```

Нагрузочные тесты (сервис должен быть запущен на `http://localhost:8080`):

```bash
make loadtest-burst     # Burst-сценарий
make loadtest-rampup    # Ramp-up
make loadtest-all       # Все сценарии
```

Целевые SLI: 99.9% успешных ответов, среднее время ответа ≤300 ms.

---

## Сборка и код

```bash
make build    # Сборка бинарника в bin/api
make fmt      # Форматирование
make lint     # golangci-lint
make clean    # Удаление bin/, coverage-файлов
```

---

## Структура проекта

```
cmd/api/           — точка входа, конфиг, роутер
internal/
  config/          — загрузка конфигурации из env
  domain/          — доменные модели (User, Team, PullRequest, PRStatus)
  handler/         — HTTP-обработчики, запросы/ответы
  repository/      — работа с БД (pr, user, team, stats)
  router/          — маршруты Gin
  service/         — бизнес-логика (команды, пользователи, PR, статистика, выбор ревьюеров)
migrations/        — SQL-миграции (up/down)
docs/              — DECISIONS.md, schema.dbml
tests/             — unit, integration, stress
openapi.yml        — спецификация API
```

---

## Документация

- **docs/schema.dbml** — схема БД (удобно открыть в [dbdiagram.io](https://dbdiagram.io)).
