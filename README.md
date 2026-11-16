# AI-BUDDY — Помощник для стажеров

## Описание
AI-BUDDY — это Full-stack приложение - AI-агент, помогающее стажерам Авито с ежедневными задачами, отслеживанием прогресса и общением в чате. Бэкенд написан на Golang с использованием фреймворка Fiber, а фронтенд — на React + TypeScript с Vite. Приложение интегрируется с моделями NVIDIA для генерации текстовых ответов и эмбеддингов.

## Технологии

| Слой | Технологии |
|------|------------|
| **Бэкенд** | Golang 1.27.1, Fiber, go-openai, PostgreSQL с расширением pgvector, Redis |
| **Фронтенд** | React 19, TypeScript, Vite |
| **Контейнеризация** | Docker & Docker Compose |
| **AI-модели** | NVIDIA Nemotron (LLM), NVIDIA embedding модель (интеграция через ai-gateway.kilo.ai и build.nvidia.com) |

## Требования

- **Golang** 1.27.1 или выше
- **Node.js** 22 или выше (с npm)
- **Docker** и **Docker Compose**
- **API-ключи** от NVIDIA:
  - Модель `nvidia/nemotron-3-ultra-550b-a55b:free` (провайдер: [kilo.ai](https://app.kilo.ai/profile))
  - Эмбеддинг-модель `nvidia/nemotron-3.5-embed-8b` (провайдер: [build.nvidia.com](https://build.nvidia.com/))
- **PostgreSQL** 18 с расширением pgvector (для Windows см. инструкцию [pgvector_pgsql_windows](https://github.com/andreiramani/pgvector_pgsql_windows))
- **Socks5 proxy** Ip4-адрес и порт (при действующей конфигурации vpn можно посмотреть в приложении(например - Happ -> Настройки -> Разрешить подключение из LAN -> текущий IP и SOCKS5 прокси))
- **Redis** официальный Docker-образ, запущенный на порту 6379 (https://hub.docker.com/_/redis)

## Локальный запуск

### Бэкенд

1. Перейдите в папку бэкенда:
   ```bash
   cd backend
   ```
2. Установите зависимости:
   ```bash
   go mod tidy
   ```
3. Выполните миграции (предварительно создайте базу данных `chunks` в PostgreSQL):
   ```bash
   go run agent/cmd/migrator/main.go -storage-path postgres://postgres:password@localhost:5432/chunks?sslmode=disable -migrations-path ./migrations -migrations-table migrations
   ```
4. Запустите ingestion-скрипт:
   ```bash
   go run agent/cmd/ingest/main.go
   ```
5. Запустите самого агента с авторизацией:
   ```bash
   go run agent/cmd/agent/main.go
   ```
   Бэкенд будет слушать порт **8080**. CORS настроен на разрешение запросов с порта **3000**.

### Фронтенд

1. Перейдите в папку фронтенда:
   ```bash
   cd frontend
   ```
2. Установите зависимости:
   ```bash
   npm ci
   ```
3. Запустите dev-сервер:
   ```bash
   npm run dev
   ```
   Vite будет слушать `0.0.0.0:3000`. Откройте <http://localhost:3000>.

> **Примечание**: Фронтенд проксирует запросы, начинающиеся с `/api`, на бэкенд на `http://localhost:8080`. Адрес бэкенда можно изменить через переменную `BACKEND_PROXY_TARGET`.

### Docker (запуск образов)
**Backend**

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
**Frontend**

Docker-образ использует Node/Vite Preview; Nginx не используется.

Сборка из корня репозитория:

```bash
docker build -t ai-buddy-frontend ./frontend
```

Запуск с backend, доступным на хост-машине:

```bash
docker run -d --name ai-buddy-frontend -p 3000:3000 \
  -e BACKEND_PROXY_TARGET=http://host.docker.internal:8080 \
  ai-buddy-frontend
```

Если frontend и backend находятся в одной Docker-сети, задайте
`BACKEND_PROXY_TARGET` как адрес backend-сервиса, например
`http://backend:8080`.


### Docker (однокомандный запуск)

Из корня проекта:

```bash
docker compose up --pull always -d
```

Это поднимет все сервисы: бэкенд, фронтенд, PostgreSQL и Redis.

## Переменные окружения

В корне проекта есть файл `.env.example`. Скопируйте его в `.env` и заполните обязательные переменные:

- `API_KEY` — ключ для доступа к модели NVIDIA (kilo.ai)
- `BASE_URL` — базовый URL ai-gateway (по умолчанию `https://ai-gateway.kilo.ai/api/v1`)
- `MODEL_ID` — идентификатор модели (например, `nvidia/nemotron-3-ultra-550b-a55b:free`)
- `EMBED_KEY` — ключ для эмбеддинг-модели (build.nvidia.com)
- `EMBED_MODEL` — имя эмбеддинг-модели (по умолчанию `nvidia/nemotron-3.5-embed-8b`)
- `DB_USER`, `DB_PASSWORD`, `DB_NAME` — данные для подключения к PostgreSQL
- `DATABASE_URL` — полный URL连接 к БД
- `REDIS_ADDR` — адрес Redis (`localhost:6379` для локального запуска или `host.docker.internal:6379` для Docker)
- `HTTP_PROXY`, `HTTPS_PROXY` — опциональные прокси (socks5)
- `NO_PROXY` — исключения из прокси
- `AUTH_COOKIE_NAME`, `AUTH_COOKIE_SECURE`, `AUTH_COOKIE_SAME_SITE` — настройки авторизации
- `CORS_ALLOWED_ORIGINS` — разрешенные origins (по умолчанию `http://localhost:3000`)
- `AUTH_SESSION_TTL` — время жизни сессии (по умолчанию `24h`)
- `BACKEND_PROXY_TARGET` — целевой адрес бэкенда

## API эндпоинты

| Метод | Эндпоинт | Описание |
|-------|----------|----------|
| `POST` | `/auth/login` | Авторизация в системе |
| `POST` | `/auth/logout` | Выход из системы |
| `POST` | `/auth/me` | Восстановление пользователя по действующей cookie |
| `POST` | `/chat` | Отправка запроса на бэкенд (только для стажеров) |
| `GET` | `/stream` | Стриминг ответов агента (SSE) |
| `GET` | `/health` | Базовый healthcheck (доступен из терминала) |

## Структура проекта

```
Hack&Change/
├── backend/           # Бэкенд на Go
│   ├── agent/         # Логика агентов
│   ├── migrations/    # DB migrations
│   ├── playbook-master # документация Авито 
│   ├── env.example # не забудьте исаользовать свои ключи и прокси-сервер!!!
│   ├── go.mod # зависимости
│   ├── go.sum # hash-функции зависимостей
│   ├── Dockerfile #docker-образ 
│   ├── intern_questions/md # вопросы стажёра - отражены в user interface
│   ├── .env           # Локальные переменные окружения
│   └── README.md      # Подробная документация бэкенда
├── frontend/          # Фронтенд на React + Vite
│   ├── Dockerfile #docker-образ
│   ├── src/           # Исходный код React-компонентов
│   ├── package.json
│   └── README.md      # Подробная документация фронтенда
├── compose.yaml # Оркестрация всех сервисов
├── gitignore # индексы неотслеживаемых файлов
├── .env.example       # Пример переменных окружения
├── .env               # Файл с реальными данными (gitignore)
└── README.md          # Этот файл
```

## Ссылка на видеодемонстрацию(также видео есть в архиве)
https://disk.yandex.ru/i/83ZtmzbSUh_A4A

## Роли в команде и выполненные задачи

# AI-ingeneer и backend developer(Дарья)
- Спроектировал архитектуру агента 
- Добавила авторизацию на основе HTTP Only Cookie 
- Реализовала ReAct сусду для одного хода агента 
- Реализовала RAG pipeline 
- Осуществила соединение с базой данных 
- Использовала Open-AI совместимыq клиент для подключения к облачным моделям 
- Внедрила Redis для сохранение в кэше авторизованный сессии на один день 
- Настроила CORS во избежание блокировок запросов с разных портов
- Покрытие модульными тестами на 100% 
- Капитан команды - коммуникации, подбор людей

# Web-disigner(Нина)
- Разрабатывала дизайн интерфейса AI-Buddy
- Продумывала структуру и расположение элементов на основных экранах.
- Разрабатывала экраны авторизации, чата, личного кабинета стажёра и кабинета руководителя.
- Прорабатывала разделы с задачами, документами, FAQ и информацией о команде.
- Подбирала цвета, шрифты и оформление интерфейса с учётом фирменного стиля Авито.
- Использовала основные цвета бренда и старалась сохранить визуальное сходство с продуктами компании.
- Разрабатывала дизайн презентации проекта в том же стиле.

# Frontend developer(Арсен)
- Сверстал интерфейс AI-Buddy по макету на React и TypeScript.
- Собрал экраны авторизации, чата, личного кабинета стажёра и кабинета руководителя.
- Сверстал разделы задач, документов, частых вопросов и карточки команды.
- Подключил вход и выход к бэкенду через HTTP-only cookie.
- Закрыл страницы кабинета: без авторизации открывается вход, меню зависит от роли.
- Настроил чат стажёра через SSE: ответ приходит потоком, карточки сотрудников показываются отдельно.
- Сохранял переписку на время сессии и очищал её при выходе.
- Подключил задачи и документы из Markdown и сделал их просмотр в интерфейсе.
- Покрыл прикладной код тестами больше чем на 80%.
- Подготовил запуск фронтенда локально и в Docker.

# Продуктовый и BI аналитик(Владимир)
-Формирование скоупа MVP
-Создание системного промта
-Создание синтетики - json-профили стажера и команды, документы для RAG-системы, задачи, FAQ, тестовые вопросы
-Создание структуры продуктовой презентации, ее текста
-Создание метрик в кабинете руководителя
-Организация рабочих процессов, контролирование сроков

# Продуктовый аналитик(Анна)
-Разработала пользовательские сценарии для стажёра и руководителя
-Описала в сценариях логику взаимодействия с AI-buddy в нестандартных ситуациях (эскалация, отказ)