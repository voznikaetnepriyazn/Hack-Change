# AI-BUDDY — Frontend

Frontend часть приложения AI-BUDDY. Проект построен на React, TypeScript и Vite. Для управления состоянием используется Zustand, для маршрутизации — React Router.

## Технологический стек

- **React 19** — библиотека для создания пользовательского интерфейса.
- **TypeScript** — статическая типизация JavaScript.
- **Vite** — инструмент разработки и сборки.
- **React Router 7** — маршрутизация приложения.
- **Zustand** — управление состоянием.
- **Vitest** — тестирование.
- **Nginx** — production-сервер для раздачи собранного приложения.
- **Docker** — контейнеризация приложения.

## Требования

Для локальной разработки:

- Node.js 22 или выше.
- npm.
- Docker Desktop — для запуска через Docker.

## Установка и запуск

Перейдите в директорию frontend:

```bash
cd frontend
```

Установите зависимости:

```bash
npm ci
```

Запустите сервер разработки:

```bash
npm run dev
```

По умолчанию приложение доступно по адресу:

```text
http://localhost:5173
```

## Доступные команды

| Команда | Описание |
|---|---|
| `npm run dev` | Запуск сервера разработки |
| `npm run build` | Проверка TypeScript и production-сборка |
| `npm run preview` | Локальный просмотр production-сборки |
| `npm test` | Запуск тестов Vitest |

Production-сборка создаётся в директории `dist`.

## Docker

Для frontend используется многоэтапная сборка Docker.

### Структура Docker-файлов

```text
frontend/
├── Dockerfile
├── .dockerignore
├── nginx.conf
├── package.json
├── package-lock.json
└── ...
```

### Сборка Docker-образа

Команду необходимо выполнять из корневой директории проекта:

```bash
docker build -t ai-buddy-frontend ./frontend
```

Во время сборки выполняются следующие действия:

1. Создаётся временный контейнер на основе Node.js.
2. Устанавливаются зависимости через `npm ci`.
3. Копируются исходные файлы приложения.
4. Выполняется `npm run build`.
5. Создаётся production-образ на основе Nginx.
6. Собранные файлы из `dist` копируются в директорию Nginx.

В результате production-контейнер содержит только Nginx и готовые статические файлы приложения. Node.js и исходный код в него не попадают.

### Запуск контейнера

```bash
docker run -d -p 8080:80 --name ai-buddy-frontend ai-buddy-frontend
```

Приложение будет доступно по адресу:

```text
http://localhost:8080
```

### Управление контейнером

Остановка:

```bash
docker stop ai-buddy-frontend
```

Удаление:

```bash
docker rm ai-buddy-frontend
```

Просмотр запущенных контейнеров:

```bash
docker ps
```

Просмотр логов:

```bash
docker logs ai-buddy-frontend
```

### Особенности Nginx

В `nginx.conf` настроена раздача статических файлов приложения.

Для корректной работы React Router используется:

```nginx
location / {
    try_files $uri $uri/ /index.html;
}
```

Это позволяет открывать вложенные маршруты приложения напрямую и обновлять страницы без получения ошибки 404 от Nginx.

Для файлов в `/assets/` настроено длительное кеширование.

## API и Vite Proxy

В `vite.config.ts` предусмотрен proxy для запросов к `/api`:

```typescript
server: {
  port: 5173,
  proxy: {
    '/api': {
      target: 'http://localhost:3001',
      changeOrigin: true,
      rewrite: (path) => path.replace(/^\/api/, ''),
    },
  },
},
```

Эта настройка предназначена для сервера разработки Vite.

Она перенаправляет запросы с `/api` на `http://localhost:3001`, удаляя префикс `/api`.

**Важно:** Vite proxy не используется в production-сборке и не работает в контейнере Nginx. После подключения backend потребуется отдельно настроить взаимодействие frontend и backend в Docker.

На текущем этапе frontend работает независимо от Go backend.

## Текущий статус

- [x] Создан frontend на React и TypeScript.
- [x] Настроен Vite.
- [x] Добавлена маршрутизация React Router.
- [x] Настроено управление состоянием через Zustand.
- [x] Подготовлен Dockerfile для production-сборки.
- [x] Настроен Nginx для раздачи приложения.
- [ ] Подключение frontend к Go backend.
- [ ] Настройка совместной работы frontend и backend через Docker Compose.

---

**AI-BUDDY** — frontend приложения.