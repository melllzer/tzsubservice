# Docker-развёртывание 
Если вы хотите запустить проект через Docker —  используйте следующую инструкцию.


# Требования

- Docker
- Docker Compose

# Запуск

1. Убедиться, что в корень проекта добавлены:
   - `Dockerfile`
   - `docker-compose.yml`

2. Выполнить в терминале:
```bash
docker-compose up --build
