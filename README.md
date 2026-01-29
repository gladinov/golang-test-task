# Numbers Microservice

Микросервис на Go с REST API для работы с числами.  
Позволяет добавлять числа в базу и получать все сохранённые числа в отсортированном виде.

---

## Запуск

### Требования

- Go ≥ 1.24
- Docker + Docker Compose
- make

---


### Использование Makefile

В проекте используется `Makefile`, который инкапсулирует основные сценарии запуска, тестирования и проверки качества кода.
```markdown
> На Windows рекомендуется использовать WSL.
```

#### Основные команды

| Команда | Описание |
|-------|---------|
| `make up` | Запуск продакшн-контейнеров |
| `make build-up` | Сборка и запуск продакшн-контейнеров |
| `make down` | Остановка продакшн-контейнеров и удаление томов |
| `make up-test` | Запуск тестовой БД |
| `make down-test` | Остановка тестовой БД |
| `make test-unit` | Запуск **unit-тестов** (без БД) |
| `make test-integration` | Запуск **интеграционных тестов** (с тестовой БД) |
| `make coverage-all` | Запуск **unit + integration тестов** и сбор общего покрытия |

---

### Примеры запуска

#### Продакшн

##### Запуск

```bash
make build-up
```
##### Остановка:
```bash
make down
```

### Тесты

#### Unit-тесты (без БД):
```bash
make test-unit
```

#### Интеграционные тесты (с тестовой БД):
```bash
make test-integration
```

### Покрытие тестами (unit + integration)
```bash
make coverage-all
```

После запуска сервис будет доступен по адресу:
`http://localhost:8080`


## Примеры
### Добавить число 1
```bash
curl -X POST http://localhost:8080/numbers -H "Content-Type: application/json" -d "{\"number\": 1}"
```
**Ответ (200 OK):**
```json
[1]
```

### Добавить число 3
```bash
curl -X POST http://localhost:8080/numbers -H "Content-Type: application/json" -d "{\"number\": 3}"
```
**Ответ (200 OK):**
```json
[1,3]
```

### Добавить число 2
```bash
curl -X POST http://localhost:8080/numbers -H "Content-Type: application/json" -d "{\"number\": 2}"
```
**Ответ (200 OK):**
```json
[1,2,3]
```