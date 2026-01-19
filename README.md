# Numbers Microservice

Микросервис на Go с REST API для работы с числами.  
Позволяет добавлять числа в базу и получать все сохранённые числа в отсортированном виде.

---

## Запуск

Сборка и запуск через Docker Compose:

```bash
docker compose --env-file .env up -d
```

## Примеры
### Добавить число 1
```bash
curl -X POST http://localhost:8080/numbers -H "Content-Type: application/json" -d "{\"number\": 1}"
```
#### Ответ: 
{[1]}

### Добавить число 3
```bash
curl -X POST http://localhost:8080/numbers -H "Content-Type: application/json" -d "{\"number\": 3}"
```
#### Ответ: 
{[1,3]}

### Добавить число 2
```bash
curl -X POST http://localhost:8080/numbers -H "Content-Type: application/json" -d "{\"number\": 2}"
```
#### Ответ: 
{[1,2,3]}