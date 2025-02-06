# Описание проекта

Этот проект предназначен для ведения учёта задач. Он запускает сервер, который по адресу `http://localhost:7540` (при стандартных настройках) предоставляет пользователю доступ к удобному интерфейсу управления задачами.

---

## Список выполненных заданий со звёздочкой

Все задания со звёздочкой были успешно выполнены. Ниже представлен список всех выполненных задач:

  1-2. Поддержка переменных окружения.  
  3. Реализация правил со звёздочкой в функции `NextDate`.  
  4. Возможность поиска задач.  
  5. Реализация авторизации.  
  6. Сборка Docker-образа.  
  7. Общее оформление проекта.

---

## Структура проекта

```plaintext
- /cmd
  - /app
    - main.go         - Точка входа в приложение.
- /data              - Директория для хранения базы данных.
- /internal
  - /handlers        - Логика HTTP-обработчиков.
    - auth.go        - Реализация авторизации.
    - handlers.go    - Основные обработчики.
    - respond.go     - Форматирование ответа.
  - /models
    - model.go       - Модели данных, используемых проектом.
  - /repository      - Работа с базой данных.
    - dbinit.go      - Инициализация базы данных.
    - tasks.go       - Методы для взаимодействия с базой данных.
  - /service         - Бизнес-логика приложения.
    - next_date.go   - Функция расчёта следующей даты для задач.
    - validator.go   - Валидация данных.
- /migration
  - scheduler.sql    - SQL-скрипт для инициализации базы данных.
- /tests             - Тесты для проекта.
- /web               - Фронтенд часть приложения.
- .dockerignore      - Файлы и папки, игнорируемые при сборке Docker-образа.
- .env               - Переменные окружения проекта.
- .gitignore         - Игнорируемые файлы для Git.
- Dockerfile         - Файл для сборки Docker-образа.
- go.mod             - Файл модулей Go.
- go.sum             - Контрольная сумма зависимостей.
```

## Локальный запуск
Если вы собираетесь использовать программу локально, то можете убрать // в начале 19 строки в файле main.go.
Используйте .env файл в котором будете хранить значения:
  1) TODO_PORT=7540 (Порт на котором хотите запускать)
  2) TODO_DBFILE=data/scheduler.db (путь к файлу базы данных)
  3) TODO_PASSWORD=gofinalproject (пароль для авторизации)
  4) TODO_SECRET=secretkey (секретный ключ для генерации токена)
Вызывайте localhost или 127.0.0.1 и через двоеточик указывайте выбранный порт.

## Инструкция по запуску тестов

Если вы меняете или дорабатываете существующий код, то для проверки текущего функционала можете использовать тестыв ./tests.
запустить все тесты вы можете командой go test ./... в главной дирректории проекта. Я старался называть все функции-обработчики так, 
как они описаны в тестах. 

Подробно про файл settings.go в /tests
Файл settings.go содержит глобальные переменные для тестов такие как: Port, DBFile и Token.
Для тестов подставляйте свои значения в эти параметры.

И так же две переменные булингового типа Search и FullNextDate. Они служат для того чтобы тесты проверяли возможности
поиска задачь и назначения следющей даты для заданий со звёздочкой.

## Инструкция по сборке Docker и запуску контейнера

Все стандартные переменные окружения передаются в докерфайле. Для сборки контейнера введите в корневой папке проекта в bash
команду: `docker build -t <свойтэг>:<версию> .` 
ВАЖНО в докерфайле используется переменная EXPOSE которая открывает в контейнере 7540 порт. Можете при запуске использовать флаги
-p <локальный хост>:7540 или -P которая автоматически сопоставит порт контейнера со стандартным на хосте.
Переменные окружения можете указывать при помощи флага -e. Пример комманды со всеми переменными окружения: 
`docker run -p 7540:7540 \
  -e TODO_PORT=7540 \
  -e TODO_DBFILE=database/scheduler.db \
  -e TODO_PASSWORD=gofinalproject \
  -e TODO_SECRET=secretkey \
  scheduler:v1.0.0`

Переменные описаны выше в инстукции к локальному запуску.