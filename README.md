
# Учебный REST API задач на Go

Этот проект показывает, как можно строить REST API для списка задач на Go и PostgreSQL.
Он сделан как учебный пример: сначала полезно понять границы между пакетами, затем самостоятельно
написать недостающие части и исправить ошибки, перечисленные в конце README.

## 1. Общая схема приложения

Запрос к API обычно проходит такой путь:

```text
Клиент (браузер, curl, Postman)
        |
        v
HTTP-маршрут
        |
        v
Handler: принимает запрос и формирует HTTP-ответ
        |
        v
TaskStore: выполняет SQL-запрос
        |
        v
PostgreSQL: хранит задачи
```

Главная идея разделения: HTTP-код не должен знать детали SQL, а код базы данных не должен
зависеть от HTTP. Модель `Task` является общим описанием данных задачи.

## 2. Структура папок

```text
RestApi/
|-- api/main.go                    # точка запуска приложения
|-- internal/models/task.go        # структуры данных и JSON-вход
|-- internal/db/database.go        # подключение к PostgreSQL
|-- internal/db/tasks.go           # SQL-операции с задачами
|-- internal/handlers/handlers.go  # HTTP-обработчики и JSON-ответы
|-- sql/init.sql/init.sql          # создание таблицы и стартовые данные
|-- docker-compose.yml             # запуск PostgreSQL в Docker
|-- go.mod                         # имя Go-модуля и зависимости
`-- go.sum                         # контрольные суммы зависимостей
```

Папка `internal` имеет специальное значение в Go: пакеты внутри нее можно импортировать только
из кода, находящегося внутри родительского модуля `RestApi`. Это помогает скрыть внутреннюю
реализацию приложения от внешних проектов.

## 3. Файл `internal/models/task.go`

Здесь описаны данные, с которыми работают остальные слои.

### `Task`

`Task` представляет строку таблицы `tasks` и объект, который API отдаёт клиенту.

- `ID int` хранит целочисленный идентификатор.
- `Title string` хранит обязательный заголовок.
- `Description string` хранит описание.
- `Completed bool` показывает, завершена ли задача.
- `CreatedAt time.Time` хранит дату создания.
- `UpdatedAt time.Time` хранит дату последнего изменения.
- Тег `json:"id"` говорит `encoding/json`, под каким именем поле попадёт в JSON.
- Тег `db:"id"` нужен `sqlx`, чтобы сопоставить колонку SQL с полем Go.

Пример результата API:

```json
{
  "id": 1,
  "title": "Изучить Go",
  "description": "Изучить язык программирования Go",
  "completed": true,
  "created_at": "2026-09-09T12:00:00Z",
  "updated_at": "2026-09-09T12:00:00Z"
}
```

### `CreateTaskInput`

Это отдельная структура для данных, которые клиент присылает при создании задачи. Дата и ID
не принимаются от клиента: их должна назначать база данных или сервер.

### `UpdateTaskInput`

Указатели `*string` и `*bool` позволяют отличить два случая:

- поле отсутствует в JSON: указатель равен `nil`, значение не меняем;
- поле передано: указатель не равен `nil`, значение обновляем.

Это основа частичного обновления `PATCH`.

## 4. Файл `internal/db/database.go`

Функция `Connect(databaseURL string)` получает строку подключения и возвращает `*sqlx.DB`.

- `sqlx.Connect("postgres", databaseURL)` открывает подключение к PostgreSQL и проверяет его.
- `*sqlx.DB` представляет пул соединений, а не одно отдельное соединение.
- `SetMaxOpenConns(20)` ограничивает максимальное число одновременно открытых соединений.
- `SetMaxIdleConns(5)` оставляет до пяти свободных соединений для повторного использования.
- Возврат `nil, err` прекращает запуск, если база недоступна.

Для работы драйвера PostgreSQL в точке входа обычно нужен импорт:

```go
_ "github.com/lib/pq"
```

Пустой импорт запускает регистрацию драйвера, хотя его имя напрямую в коде не используется.

## 5. Файл `internal/db/tasks.go`

`TaskStore` хранит ссылку на базу и предоставляет методы работы с задачами. Это слой доступа к
данным: здесь находятся SQL-запросы, но нет HTTP-ответов.

### `NewTaskStore`

Конструктор получает готовое подключение и возвращает `TaskStore`. Конструктор делает создание
объекта единообразным и не даёт другим пакетам напрямую собирать его внутренние поля.

### `GetAll`

1. Создаётся срез `tasks`.
2. Многострочная строка с SQL выбирает все колонки.
3. `s.db.Select(&tasks, query)` выполняет запрос и заполняет срез.
4. При ошибке возвращаются `nil` и ошибка.
5. При успехе возвращается список задач.

`Select` подходит, когда запрос должен вернуть несколько строк.

### `GetByID`

Метод получает ID, передаёт его в SQL как `$1` и возвращает одну задачу. Параметры вроде `$1`
важны: они отделяют SQL-команду от пользовательского значения и помогают избежать SQL-инъекций.

Для одной строки нужно использовать переменную типа `models.Task`, а не `[]models.Task`:

```go
var task models.Task
err := s.db.Get(&task, query, id)
return &task, err
```

`sql.ErrNoRows` означает, что задача не найдена. Это отдельный случай, который HTTP-слой может
превратить в ответ `404 Not Found`.

### `CreateTask`

Функция принимает `CreateTaskInput`, создаёт время `now`, выполняет `INSERT` и получает созданную
строку благодаря `RETURNING`. Такой подход позволяет сразу вернуть клиенту настоящий ID и даты.

В текущем проекте функция оформлена как обычная функция с первым аргументом `*TaskStore`, хотя
остальные операции оформлены методами. Для единообразия её можно переписать как метод:

```go
func (s *TaskStore) CreateTask(input models.CreateTaskInput) (*models.Task, error)
```

### `UpdateTask`

Функция сначала читает задачу, затем изменяет только поля, которые пришли во входной структуре,
обновляет `UpdatedAt` и выполняет `UPDATE ... RETURNING`.

Это понятная учебная реализация, но два SQL-запроса подряд могут привести к конфликту при
параллельных изменениях. В более серьёзном приложении используют транзакцию или один SQL-запрос.

### `DELETEtask`

Функция выполняет `DELETE`, получает число удалённых строк через `RowsAffected` и сообщает об
ошибке, если ID не существовал. В Go имена экспортируемых функций обычно пишутся в стиле
`DeleteTask`, поэтому `DELETEtask` стоит переименовать.

## 6. Файл `internal/handlers/handlers.go`

Handler является связующим слоем между HTTP и `TaskStore`.

### `Handler` и `NewHandler`

Структура хранит `store`, чтобы каждый обработчик мог вызвать операции базы данных. Конструктор
принимает зависимость извне. Это называется dependency injection: обработчику не нужно самому
создавать подключение к базе, и его легче тестировать.

### `respondWithJSON`

Функция выполняет общий алгоритм HTTP JSON-ответа:

1. устанавливает `Content-Type`;
2. отправляет HTTP-статус через `WriteHeader`;
3. кодирует `payload` в JSON и пишет его в ответ.

Общий помощник убирает копирование этих трёх строк из каждого handler.

### `respondWithError`

Функция оборачивает текст ошибки в объект `{ "error": "..." }` и повторно использует общий
JSON-ответ.

### `GetAllTasks`

Обработчик вызывает `h.store.GetAll()`. Если база вернула ошибку, клиент получает `500 Internal
Server Error`; при успехе получает список задач со статусом `200 OK`.

### `GetTask`

Обработчик извлекает часть пути, преобразует её из строки в число через `strconv.Atoi`, вызывает
`GetByID` и возвращает задачу. Неправильный ID должен давать `400 Bad Request`, отсутствующая
задача — `404 Not Found`.

Важно: текущий код сам разбирает путь и ожидает, что в `r.URL.Path` первой частью будет число.
Если маршрут будет `/tasks/1`, нужно сначала убрать префикс `/tasks/` или использовать роутер.

## 7. Файл `sql/init.sql/init.sql`

SQL-скрипт запускается PostgreSQL при первом создании Docker volume.

- `DROP TABLE IF EXISTS tasks` удаляет старую таблицу, если она существует.
- `CREATE TABLE` создаёт таблицу и задаёт ограничения.
- `SERIAL PRIMARY KEY` создаёт автоувеличиваемый ID.
- `NOT NULL` запрещает обязательному полю быть пустым.
- `DEFAULT FALSE` и `DEFAULT CURRENT_TIMESTAMP` задают значения автоматически.
- `INSERT` добавляет три стартовые задачи.

Описание `description TEXT` сейчас допускает `NULL`, а Go-модель использует `string`. Для простого
учебного проекта можно добавить `description TEXT NOT NULL DEFAULT ''` или использовать
`sql.NullString` в модели.

## 8. Файл `docker-compose.yml`

Compose запускает только PostgreSQL:

- `image` выбирает готовый образ PostgreSQL 15 Alpine;
- `environment` создаёт базу, пользователя и пароль;
- `ports` делает базу доступной на `localhost:5432`;
- `volumes` сохраняет данные и подключает SQL-скрипт;
- `healthcheck` проверяет готовность PostgreSQL командой `pg_isready`.

Запуск:

```powershell
cd RestApi
docker compose up -d
```

Остановка:

```powershell
docker compose down
```

Чтобы удалить и данные, и volume, используйте это только в учебной базе:

```powershell
docker compose down -v
```

SQL-скрипт из `docker-entrypoint-initdb.d` выполняется только при первом создании volume.
Изменение `init.sql` само по себе не перезапускает инициализацию.

## 9. Как должен запускаться API

Сейчас `api/main.go` пустой, поэтому полноценный HTTP-сервер ещё не собран. Типичный порядок
в точке входа такой:

1. импортировать `net/http`, `database/sql` и драйвер PostgreSQL;
2. создать DSN подключения;
3. вызвать `database.Connect`;
4. создать `TaskStore`;
5. создать `Handler`;
6. зарегистрировать маршруты;
7. вызвать `http.ListenAndServe`.

Пример будущей последовательности без готового роутера:

```go
store := database.NewTaskStore(db)
handler := handlers.NewHandler(store)
http.HandleFunc("/tasks", handler.GetAllTasks)
http.HandleFunc("/tasks/", handler.GetTask)
log.Fatal(http.ListenAndServe(":8080", nil))
```

## 10. Проверка проекта

Из папки `RestApi`:

```powershell
go test ./...
go vet ./...
go fmt ./...
```

`go test ./...` сейчас проверяет компиляцию, но тестов нет. Поэтому успешная команда ещё не
доказывает, что SQL-запросы работают. `go vet` полезен для проверки подозрительных конструкций,
например неправильных struct tags.

## 11. Ошибки, которые нужно исправить самостоятельно

Это хороший список упражнений после чтения:

1. В `GetAll` исправить `orderd by` на `ORDER BY`.
2. В `GetByID` заменить срез `[]models.Task` на одну `models.Task`.
3. В `UpdateTask` добавить запятую после `title = $1`.
4. Проверить, что все теги модели имеют формат `db:"id"` без пробела после двоеточия.
5. Добавить импорт драйвера `_ "github.com/lib/pq"`.
6. Заполнить `api/main.go`, чтобы сервер действительно запускался.
7. Добавить маршруты для `POST`, `PATCH` и `DELETE`.
8. Разделять ошибку «нет строки» и настоящую ошибку базы данных.
9. Проверять входной JSON и обязательный `title`.
10. Добавить тесты для store и HTTP handlers.

## 12. Как учиться по этому проекту

Сначала закрой README и попробуй по памяти объяснить четыре слоя: модель, база, handler и
точка входа. Затем напиши `main.go` самостоятельно. После этого реализуй один endpoint от
начала до конца: сначала SQL, потом метод store, потом handler, потом маршрут и тест.

Полезный цикл обучения:

```text
прочитал функцию -> закрыл файл -> написал по памяти -> запустил gofmt -> проверил тестом
```

Не обязательно запоминать каждую строку дословно. Важно понимать, какой слой отвечает за действие,
какие данные входят и выходят, где обрабатывается ошибка и почему функция возвращает именно такие
типы.

## 13. Актуальная проверка

`UpdateTask` получает идентификатор через `parseTaskIDFromPath`. Поэтому запросы к конкретной
задаче используют путь `/tasks/{id}`; некорректный путь возвращает `400 Bad Request`.

Перед отправкой изменений в репозиторий выполните:

```powershell
go build ./...
go test ./...
```
=======
# Git lab Ci



## Getting started

To make it easy for you to get started with GitLab, here's a list of recommended next steps.

Already a pro? Just edit this README.md and make it your own. Want to make it easy? [Use the template at the bottom](#editing-this-readme)!

## Add your files

* [Create](https://docs.gitlab.com/user/project/repository/web_editor/#create-a-file) or [upload](https://docs.gitlab.com/user/project/repository/web_editor/#upload-a-file) files
* [Add files using the command line](https://docs.gitlab.com/topics/git/add_files/#add-files-to-a-git-repository) or push an existing Git repository with the following command:

```
cd existing_repo
git remote add origin http://192.168.56.111/gitlab/git-lab-ci.git
git branch -M main
git push -uf origin main
```

## Integrate with your tools

* [Set up project integrations](http://192.168.56.111/gitlab/git-lab-ci/-/settings/integrations)

## Collaborate with your team

* [Invite team members and collaborators](https://docs.gitlab.com/user/project/members/)
* [Create a new merge request](https://docs.gitlab.com/user/project/merge_requests/creating_merge_requests/)
* [Automatically close issues from merge requests](https://docs.gitlab.com/user/project/issues/managing_issues/#closing-issues-automatically)
* [Enable merge request approvals](https://docs.gitlab.com/user/project/merge_requests/approvals/)
* [Set auto-merge](https://docs.gitlab.com/user/project/merge_requests/auto_merge/)

## Test and Deploy

Use the built-in continuous integration in GitLab.

* [Get started with GitLab CI/CD](https://docs.gitlab.com/ci/quick_start/)
* [Analyze your code for known vulnerabilities with Static Application Security Testing (SAST)](https://docs.gitlab.com/user/application_security/sast/)
* [Deploy to Kubernetes, Amazon EC2, or Amazon ECS using Auto Deploy](https://docs.gitlab.com/topics/autodevops/requirements/)
* [Use pull-based deployments for improved Kubernetes management](https://docs.gitlab.com/user/clusters/agent/)
* [Set up protected environments](https://docs.gitlab.com/ci/environments/protected_environments/)

***

# Editing this README

When you're ready to make this README your own, just edit this file and use the handy template below (or feel free to structure it however you want - this is just a starting point!). Thanks to [makeareadme.com](https://www.makeareadme.com/) for this template.

## Suggestions for a good README

Every project is different, so consider which of these sections apply to yours. The sections used in the template are suggestions for most open source projects. Also keep in mind that while a README can be too long and detailed, too long is better than too short. If you think your README is too long, consider utilizing another form of documentation rather than cutting out information.

## Name
Choose a self-explaining name for your project.

## Description
Let people know what your project can do specifically. Provide context and add a link to any reference visitors might be unfamiliar with. A list of Features or a Background subsection can also be added here. If there are alternatives to your project, this is a good place to list differentiating factors.

## Badges
On some READMEs, you may see small images that convey metadata, such as whether or not all the tests are passing for the project. You can use Shields to add some to your README. Many services also have instructions for adding a badge.

## Visuals
Depending on what you are making, it can be a good idea to include screenshots or even a video (you'll frequently see GIFs rather than actual videos). Tools like ttygif can help, but check out Asciinema for a more sophisticated method.

## Installation
Within a particular ecosystem, there may be a common way of installing things, such as using Yarn, NuGet, or Homebrew. However, consider the possibility that whoever is reading your README is a novice and would like more guidance. Listing specific steps helps remove ambiguity and gets people to using your project as quickly as possible. If it only runs in a specific context like a particular programming language version or operating system or has dependencies that have to be installed manually, also add a Requirements subsection.

## Usage
Use examples liberally, and show the expected output if you can. It's helpful to have inline the smallest example of usage that you can demonstrate, while providing links to more sophisticated examples if they are too long to reasonably include in the README.

## Support
Tell people where they can go to for help. It can be any combination of an issue tracker, a chat room, an email address, etc.

## Roadmap
If you have ideas for releases in the future, it is a good idea to list them in the README.

## Contributing
State if you are open to contributions and what your requirements are for accepting them.

For people who want to make changes to your project, it's helpful to have some documentation on how to get started. Perhaps there is a script that they should run or some environment variables that they need to set. Make these steps explicit. These instructions could also be useful to your future self.

You can also document commands to lint the code or run tests. These steps help to ensure high code quality and reduce the likelihood that the changes inadvertently break something. Having instructions for running tests is especially helpful if it requires external setup, such as starting a Selenium server for testing in a browser.

## Authors and acknowledgment
Show your appreciation to those who have contributed to the project.

## License
For open source projects, say how it is licensed.

## Project status
If you have run out of energy or time for your project, put a note at the top of the README saying that development has slowed down or stopped completely. Someone may choose to fork your project or volunteer to step in as a maintainer or owner, allowing your project to keep going. You can also make an explicit request for maintainers.
>>>>>>> 5949f998778350ac4205da540727e8525c70f4d8
