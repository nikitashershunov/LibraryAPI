# Library REST API

Микросервис для управления библиотекой книг на Go с RESTful API.

## Основные функции

- Полноценные CRUD операции для книг (создание, чтение, обновление, удаление)
- Фильтрация по:
  - Названию
  - Жанрам
- Сортировка по:
  - ID
  - Названию
  - Году издания
  - Количеству страниц
- Пагинация результатов
- Подробное логирование в JSON формате

## API Endpoints

### Основные
| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/v1/books` | Получить список книг (с фильтрацией) |
| `POST` | `/v1/books` | Добавить новую книгу |
| `GET` | `/v1/books/:id` | Получить книгу по ID |
| `PATCH` | `/v1/books/:id` | Обновить данные книги |
| `DELETE` | `/v1/books/:id` | Удалить книгу |

### Системные
| Метод | Путь | Описание |
|-------|------|----------|
| `GET` | `/v1/healthcheck` | Проверка состояния сервера |


## Предварительные требования
- Go 1.20+
- PostgreSQL 12+
- Утилита `migrate`

## Установка
1. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/nikitashershunov/LibraryAPI.git
   cd LibraryAPI
   ```

2. Настройте подключение к БД (создайте переменную окружения BOOKS_DB_DSN со строкой DSN или передайте значение DSN через флаг --db-dsn):
   ```bash
   export BOOKS_DB_DSN="postgres://{USERNAME}:{PASSWORD}@localhost/{DATABASE_NAME}"
   ```

3. Примените миграции:
   ```bash
   make db/migrations/up
   ```

4. Запустите сервер:
   ```bash
   make run/api
   ```

## Конфигурация сервера

Настройки сервера можно задать через флаги командной строки:

| Параметр          | По умолчанию       | Описание                          |
|-------------------|--------------------|-----------------------------------|
| `--port`          | 4000               | Порт сервера                      |
| `--env`           | development        | Окружение (development/production)|
| `--db-dsn`        | BOOKS_DB_DSN       | Строка подключения к PostgreSQL (DSN)|
| `--db-max-idle-conns` | 25           | Макс. количество idle-соединений  |
| `--db-max-open-conns` | 25           | Макс. количество соединений с БД  |

## Цели Makefile

```bash
make help                       # Показать доступные команды
make run/api                    # Запустить сервер
make db/psql                    # Подключиться к БД через psql
make db/migrations/up           # Применить миграции
make db/migrations/new name=$1  # Создать новые миграции
```

## Структура проекта

```
.
├── cmd
│   └── api            # Основное приложение
├── internal
│   ├── data           # Модели и работа с БД
│   ├── jsonlog        # Логирование в JSON
│   └── validator      # Валидация данных
├── migrations         # SQL-миграции
├── Makefile           # Автоматизация команд для разработки
└── README.md          # Документация
```
