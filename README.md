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

# Endpoints/ эндпоинты

Создание подписки:POST /subscriptions

Формат:
{
  "service_name": "Netflix",
  "price": 599,
  "user_id": "user123",
  "start_date": "01-2026"
}

Получение всех доступных подписок: GET /subscriptions?user_id=user123&service_name=Netflix

Получение подписок по конкретному ID : GET /subscriptions/{id}

Обновление: PUT /subscriptions/{id}

Удаление: DELETE /subscriptions/{id}

Общий прайс подписки: POST /subscriptions/total-cost

Формат:
{
  "user_id": "user123",
  "service_name": "Netflix",
  "period_start": "01-2026",
  "period_end": "12-2026"
}
Ответ: { "total_cost": 7188 }


# Примечание: Примечание
Дата указывается в формате MM-YYYY (например, "01-2026")
Все цены — целые числа (рубли)
Проект не использует Docker, миграции, фреймворки — только стандартная библиотека Go по причинам :
1.Минимум зависимостей — проще запустить и поддерживать
2.Полный контроль над SQL — видно, какие запросы выполняются
3.Быстрая сборка и запуск — не нужно собирать образы или настраивать контейнеры
