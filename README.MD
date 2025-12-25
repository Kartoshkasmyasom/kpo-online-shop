# KPO Online Shop — Orders + Payments + Frontend (Go) + Kafka + Postgres

Проект состоит из 3 сервисов:
- **orders** (порт хоста: `8081`) — создание заказа, список заказов, статус заказа
- **payments** (порт хоста: `8082`) — аккаунт пользователя, пополнение, баланс
- **frontend** (порт хоста: `8080`) — минимальный UI для вызова API

Инфраструктура:
- **Postgres** (порт хоста: `5432`)
- **Kafka + ZooKeeper**
- **kafka-init** — одноразовый сервис для создания топиков

---

## Архитектура

### Orders Service
- `POST /create` создаёт заказ со статусом `NEW` и пишет событие в **orders_outbox** (в одной транзакции)
- Outbox publisher отправляет событие в Kafka topic **payments.request**
- Kafka consumer читает **payments.result** и обновляет `orders.status` на `FINISHED` или `CANCELLED`

### Payments Service
- Kafka consumer читает **payments.request**
- **Transactional Inbox:** вставляет `message_id` в `payments_inbox`
  - если сообщение уже было — повторная обработка не происходит
- Списывает деньги и пишет таблицу `payments` (идемпотентно по `order_id`)
- **Transactional Outbox:** пишет событие результата в `payments_outbox`
- Outbox publisher отправляет событие в Kafka topic **payments.result**

---

## Требования

- Docker Desktop

---

## Запуск

Из корня проекта:

```powershell
docker compose up -d --build
```
Для тестирования проекта в браузере зайдите на:
```bash
http://localhost:8080/
