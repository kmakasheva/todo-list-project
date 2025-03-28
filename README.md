![Логотип проекта](https://cdn-icons-png.flaticon.com/512/6619/6619606.png "Логотип проекта")

# To-Do List  
To-Do List — это backend-приложение на Go, которое позволяет управлять списком задач через API.  
Задачи можно создавать, редактировать, удалять, переносить на новую дату по заданному интервалу и искать по тексту.  

Данные хранятся в SQLite, реализована локальная авторизация, а логирование ведется через slog.  
В конце разработки проект был контейнеризирован с Docker, а также частично настроен CI/CD.  

![GitHub Workflow Status](https://img.shields.io/github/actions/workflow/status/kmakasheva/todo-list-project/ci.yml?branch=main)  
![GitHub go.mod Go version](https://img.shields.io/github/go-mod/go-version/kmakasheva/todo-list-project)  
![GitHub last commit](https://img.shields.io/github/last-commit/kmakasheva/todo-list-project)  

## 🛠 Стек технологий  
- **Язык**: Go  
- **База данных**: SQLite  
- **Аутентификация**: Локальная авторизация  
- **Логирование**: slog (уровни INFO, ERROR)  
- **Контейнеризация**: Docker  
- **CI/CD**: GitHub Actions  

## 🚀 Развертывание  
### 1. Локальный запуск  
Склонировать репозиторий:  
```sh
git clone https://github.com/kmakasheva/todo-list-project.git  
cd todo-list-project  
```  
Установить зависимости:  
```sh
go mod tidy  
```  
Запустить проект:  
```sh
go run ./cmd/*.go  
```  

### 2. Запуск в Docker  
Собрать и запустить контейнер:  
```sh
docker-compose up --build  
```  

## 📌 Функциональность  
✅ Работа с задачами (создание, редактирование, удаление)  
✅ Поиск задач по тексту  
✅ Перенос задач на новую дату по заданному интервалу  
✅ Локальная авторизация  
✅ Хранение данных в SQLite  
✅ Логирование действий через slog (INFO, ERROR)  
✅ Контейнеризация с Docker  
✅ Частично настроенный CI/CD (GitHub Actions)  

## 📄 API-маршруты  
| Метод  | Эндпоинт        | Описание |
|--------|----------------|----------|
| POST   | /api/task      | Создать задачу |
| GET    | /api/tasks     | Получить список задач |
| GET    | /api/task      | Получить задачу |
| UPDATE | /api/task      | Редактирование задачи |
| DELETE | /api/task      | Удалить задачу |
| POST   | /api/signin    | Авторизация |
| POST   | /api/task/done | Удаление неактуальной задачи |

## 📌 Файл `.env`  
```
TODO_PORT=7540  
TODO_DBFILE=./  
CONFIG_PATH=./config/local.yaml  
TODO_PASSWORD=kkk123  
SECRET=todosecret  
```  
(Нужно авторизоваться через сервер, получить токен через `F12 -> Applications`, скопировать его и вставить в `tests/settings.go` в поле `token`)  

## 🧪 Запуск тестов  
1. Укажите необходимые параметры в `tests/settings.go`  
2. Запустите тесты командой:  
```sh
go test ./tests  
```  

## 📚 Планы по доработке  
- Подключить PostgreSQL вместо SQLite для более надежного хранения данных.  
- Оптимизировать обработку ошибок, добавив кастомные ошибки и единый обработчик.  
- Пересмотреть CI/CD — например, добавить автоматические тесты при пуше.  

## 📎 Дополнительно  
- **Разработчик**: [@kmakasheva](https://github.com/kmakasheva)  
- **GitHub**: [todo-list-project](https://github.com/kmakasheva/todo-list-project)  
