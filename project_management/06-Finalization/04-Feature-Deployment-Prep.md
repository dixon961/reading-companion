## Фича 4: Подготовка к развертыванию (Production-ready build)

- [ ] (DevOps): Задача 4.1: Оптимизировать `backend/Dockerfile`, используя multi-stage build, чтобы финальный образ содержал только скомпилированный бинарный файл.
- [ ] (Frontend): Задача 4.2: Убедиться, что команда `npm run build` в директории `frontend` успешно собирает статические файлы в папку `dist/`.
- [ ] (DevOps): Задача 4.3: Создать `nginx/Dockerfile` и `nginx/default.conf` для обслуживания статических файлов фронтенда.
- [ ] (DevOps): Задача 4.4: Создать файл `docker-compose.prod.yml`, который описывает сервисы `backend`, `postgres` и `nginx`.
- [ ] (DevOps): Задача 4.5: В `docker-compose.prod.yml` настроить `nginx` так, чтобы он обслуживал статику фронтенда и проксировал запросы `/api/*` на сервис `backend`.
- [ ] (DevOps): Задача 4.6: Добавить в `Makefile` цели `build-prod` и `run-prod` для сборки и запуска production-окружения.