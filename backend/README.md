# AI-BUDDY — Backend

Backend приложения AI-BUDDY на Golang с использованием PostgreSQL и Redis.

## Требования

- Golang 1.27.1 или выше
- go
- API-key для использования модели nvidia/nemotron-3-ultra-550b-a55b:free (провайдер - https://app.kilo.ai/profile)
- API-key для использования эмбеддинг-модели nvidia/nemotron-3.5-embed-8b (провайдер - https://build.nvidia.com/)
- PosgreSQL 18 с расширением pgvector (https://github.com/andreiramani/pgvector_pgsql_windows - для Windows, как пример)
- официальный Docker-образ Redis, запущенный на порту 6379 (https://hub.docker.com/_/redis)
- прокси socks5 - Ip4-адрес и порт(можно посмотреть конфигурацию в Happ)

## Локальный запуск

В корне проекта выполнить:

```bash
cd backend
go mod tidy
go run agent/cmd/migrator/main.go -storage-path postgres://postgres:сюда_ввести_пароль_от _локального_субд@localhost:5432/chunks?sslmode=disable -migrations-path ./migrations -migrations-table migrations (необходима созданнная база данных chunks в pgAdmin4 и расширение pgvector)
go run agent/cmd/ingest/main.go
go run agent/cmd/agent/main.go
```

Слушается 8080 порт. CORS разрешает запросы с 3000 порта.

| Маршрут | Описание |
|---|---|
| `POST /auth/login` | Авторизация в системе |
| `POST /auth/logout` | Выход из системы | 
| `POST /auth/me` | recover | только из терминала |
| `POST /chat` | Отправка запроса на бэкенд |
| `GET /stream` | Стриминг ответов агента |
| `GET /health` | базовый healthcheck | только из терминала |

## Проверки и сборка

| Команда | Описание |
|---|---|
| `go run build` | Сборка |
| `go test ./` | Запуск тестов из необходимой папки |

## Docker

Сборка из корня репозитория (предварительно запустить Docker-образ Redis-a с на 6379 порту):

```bash
docker build -t ai-buddy-backend ./backend
```

Запуск с необходимыми переменными окружения:

```bash
docker run -d --name ai-buddy-backend -p 8080:8080 \
  -e API_KEY="ключ для модели nvidia/nemotron-3-ultra-550b-a55b:free (провайдер - https://app.kilo.ai/profile)" \
  -e BASE_URL="https://ai-gateway.kilo.ai/api/v1" \
  -e MODEL_ID="nvidia/nemotron-3-ultra-550b-a55b:free" \
  -e LLM_STREAM=true \
  -e AUTH_COOKIE_NAME=auth_session \
  -e AUTH_COOKIE_SECURE=false \
  -e AUTH_COOKIE_SAME_SITE=lax \
  -e CORS_ALLOWED_ORIGINS=http://localhost:3000 \
  -e REDIS_ADDR=host.docker.internal:6379 \
  -e HTTP_PROXY=socks5://Ip4-адрес:порт \
  -e HTTPS_PROXY=socks5://Ip4-адрес:порт \
  -e NO_PROXY=localhost,127.0.0.1,host.docker.internal \
  -e EMBED_URL="https://integrate.api.nvidia.com/v1/embeddings" \
  -e EMBED_KEY="ключ для модели nvidia/nemotron-3.5-embed-8b (провайдер - https://build.nvidia.com/)" \
  -e EMBED_MODEL="nvidia/nemotron-3.5-embed-8b" \
    -e DB_HOST=localhost \
    -e DB_PORT=5432 \
    -e DB_USER=postgres \
    -e DB_PASSWORD=пароль_ваш \
    -e DB_NAME=chunks \
    -e DB_SSLMODE=disable \
  -e DATABASE_URL=postgres://postgres:пароль_от_субд@host.docker.internal:5432/chunks?sslmode=disable \
  
  ai-buddy-backend
```

## Интеграция с frontend

- Backend по умолчанию слушает порт `8080`; frontend обслуживается на `3000`.
- Backend CORS разрешает origin frontend на порту 3000. 
- Авторизация использует HTTP-only cookie; API-запросы и SSE отправляют
  credentials.
- `/auth/me` восстанавливает пользователя по действующей cookie после
  перезагрузки приложения.
- `/chat` и `/stream` доступны только стажёру. События `employee_cards`
  отображаются в чате отдельными карточками, текстовые события остаются
  обычными сообщениями SSE.
- Подробные данные профиля стажёра в `getIntern()` пока являются демонстрационными:
  backend предоставляет только идентификатор, имя, логин и роль пользователя.
