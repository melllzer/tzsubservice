# Subscription Service — Test Task

REST API для управления подписками пользователей.  


# Возможности

- Создание, чтение, обновление и удаление подписок (CRUD)
- Фильтрация по `user_id` и `service_name`
- Расчёт общей стоимости активных подписок за период (`MM-YYYY`)
- Хранение данных в PostgreSQL

# Технологии

- Go 1.25.6
- PostgreSQL
- Встроенный `net/http`
- Библиотека `github.com/lib/pq` для работы с БД

# Запуск

1. Должен быть запущен PostgreSQL
2. Создание базы данных `subservice_db`
3. Пароль от БД "3228@
4. Выполнение: 

```bash
go run cmd/main.go
