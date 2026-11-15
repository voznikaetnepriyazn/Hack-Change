# AI-BUDDY — Frontend

Frontend приложения AI-BUDDY на React, TypeScript и Vite.

## Требования

- Node.js 22 или выше
- npm

## Локальный запуск

```bash
cd frontend
npm ci
npm run dev
```

Vite слушает `0.0.0.0:3000` (порт фиксирован через `strictPort`). Откройте
<http://localhost:3000>.

Запросы frontend идут на `/api`. Dev- и preview-серверы Vite удаляют префикс
`/api` и проксируют запрос в Go backend на `http://localhost:8080`. Другой
адрес backend задаётся переменной `BACKEND_PROXY_TARGET`.

## Проверки и сборка

| Команда | Описание |
|---|---|
| `npm run build` | Проверка TypeScript и production-сборка |
| `npm test -- --run` | Запуск тестов Vitest один раз |
| `npm run preview` | Локальный просмотр production-сборки на порту 3000 |

Production-сборка создаётся в `dist`.

## Docker

Docker-образ использует Node/Vite Preview; Nginx не используется.

Сборка из корня репозитория:

```bash
docker build -t ai-buddy-frontend ./frontend
```

Запуск с backend, доступным на хост-машине:

```bash
docker run --rm -p 3000:3000 \
  -e BACKEND_PROXY_TARGET=http://host.docker.internal:8080 \
  ai-buddy-frontend
```

Если frontend и backend находятся в одной Docker-сети, задайте
`BACKEND_PROXY_TARGET` как адрес backend-сервиса, например
`http://backend:8080`.

## Интеграция с backend

- Backend по умолчанию слушает порт `8080`; frontend обслуживается на `3000`.
- Backend CORS должен разрешать origin frontend. При проксировании через Vite
  запросы идут через тот же origin, но прямое обращение к backend требует
  корректного `CORS_ALLOWED_ORIGINS`.
- Авторизация использует HTTP-only cookie; API-запросы и SSE отправляют
  credentials.
- `/auth/me` восстанавливает пользователя по действующей cookie после
  перезагрузки приложения.
- `/chat` и `/stream` доступны только стажёру. События `employee_cards`
  отображаются в чате отдельными карточками, текстовые события остаются
  обычными сообщениями SSE.
- Подробные данные профиля стажёра в `getIntern()` пока являются демонстрационными:
  backend предоставляет только идентификатор, имя, логин и роль пользователя.
