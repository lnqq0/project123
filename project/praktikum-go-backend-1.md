# Бэкенд на Go: практикум и проект

Раздатка к паре: упражнения, рабочий пример целиком, разбор механики и задание на проект.

**Не переписывайте код с проектора — копируйте отсюда.** Если на паре что-то осталось непонятным или мы пропустили слайд с фиолетовой полосой — смотрите [Часть 3](#часть-3-почему-это-работает), там разобрано, почему каждая конструкция работает именно так.

- [Часть 1. Упражнения на паре](#часть-1-упражнения-на-паре)
- [Часть 2. Рабочий пример целиком](#часть-2-рабочий-пример-целиком)
- [Часть 3. Почему это работает](#часть-3-почему-это-работает) — разбор механики, без магии
- [Часть 4. Проект на три пары](#часть-4-проект-на-три-пары)
- [Часть 5. Чеклист перед сдачей](#часть-5-чеклМист-перед-сдачей)
- [Часть 6. Домашнее задание](#часть-6-домашнее-задание)

---

## Часть 1. Упражнения на паре

Всё выполняется **в браузере**, ничего устанавливать не нужно.

- Go: **https://go.dev/play**
- PostgreSQL: **https://pgplayground.com** или **https://extendsclass.com/postgresql-online.html**

### Упражнение 1. Почини JSON (5 минут)

Скопируйте в Playground и нажмите Run.

```go
package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type Task struct {
	ID    int64
	Title string
	done  bool
}

func main() {
	t := Task{ID: 1, Title: "Сверстать карточку", done: false}

	out, err := json.Marshal(t)
	if err != nil {
		fmt.Println("ошибка:", err)
		os.Exit(1)
	}
	fmt.Println(string(out))
}
```

Сейчас программа печатает:

```json
{"ID":1,"Title":"Сверстать карточку"}
```

**Задача:** добиться ровно такого вывода:

```json
{"id":1,"title":"Сверстать карточку","done":false}
```

Нужно два изменения. Подсказка: одно — про регистр буквы, второе — про теги структуры.

**Вопрос:** почему компилятор ни словом не пожаловался на потерянное поле?

<details>
<summary>Ответ</summary>

Поле `done` начинается с маленькой буквы, значит оно неэкспортируемое — пакет `encoding/json` физически не может его прочитать. Для компилятора это совершенно нормальная ситуация, ошибки здесь нет. Правильно так:

```go
type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}
```

</details>

### Упражнение 2. Хендлер без сервера (5 минут)

В Playground нельзя открыть порт — зато хендлер можно вызвать напрямую. Именно так пишутся настоящие тесты.

```go
package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
)

type Task struct {
	ID    int64  `json:"id"`
	Title string `json:"title"`
	Done  bool   `json:"done"`
}

func tasksHandler(w http.ResponseWriter, r *http.Request) {
	tasks := []Task{
		{ID: 1, Title: "Сверстать карточку", Done: true},
		{ID: 2, Title: "Написать хендлер", Done: false},
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(tasks)
}

func main() {
	req := httptest.NewRequest("GET", "/tasks", nil)
	rec := httptest.NewRecorder()

	tasksHandler(rec, req)

	res := rec.Result()
	fmt.Println("статус:", res.StatusCode)
	fmt.Println("Content-Type:", res.Header.Get("Content-Type"))
	fmt.Println("тело:", rec.Body.String())
}
```

**Задачи:**

1. Добавьте в ответ третью задачу.
2. Сделайте так, чтобы хендлер возвращал `201`, а не `200`.
3. Поставьте `w.WriteHeader(...)` **перед** `w.Header().Set(...)`. Что изменилось в выводе? Почему?

`httptest.NewRecorder()` — это подставной `ResponseWriter`: вместо сети он пишет в буфер, который потом можно прочитать.

**Если получили такую ошибку:**

```text
res.StatusCode undefined (type func() *http.Response has no field or method StatusCode)
res.Header undefined (type func() *http.Response has no field or method Header)
```

Забыты скобки: `rec.Result` вместо `rec.Result()`. Без скобок вы кладёте в `res` **саму функцию**, а не результат её вызова — отсюда и `type func() *http.Response` в тексте ошибки. Go честно говорит, какой тип получился: прочитайте его в скобках, и причина видна сразу.

Это общее правило: в Go имя функции без скобок — это значение-функция, которое можно передать или сохранить. Мы на этом же свойстве построили `h.list` в роутере (3.16) — там скобки не нужны намеренно.

**Короткий вариант без `Result()`.** У `ResponseRecorder` есть и свои поле с методом, они дают то же самое:

```go
fmt.Println("статус:", rec.Code)
fmt.Println("Content-Type:", rec.Header().Get("Content-Type"))
fmt.Println("тело:", rec.Body.String())
```

`rec.Result()` возвращает настоящий `*http.Response` — такой же, какой получил бы реальный клиент. В настоящих тестах обычно пишут именно так.

### Упражнение 3. Роутер (бонус, если успели)

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
)

func main() {
	mux := http.NewServeMux()

	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, "список задач")
	})
	mux.HandleFunc("GET /tasks/{id}", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintf(w, "задача с id=%s\n", r.PathValue("id"))
	})
	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprintln(w, "задача создана")
	})

	try := func(method, path string) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		fmt.Printf("%-6s %-12s -> %d %s", method, path, rec.Code, rec.Body.String())
	}

	try("GET", "/tasks")
	try("GET", "/tasks/42")
	try("POST", "/tasks")
	try("DELETE", "/tasks/42")
	try("GET", "/unknown")
}
```

Вывод:

```text
GET    /tasks       -> 200 список задач
GET    /tasks/42    -> 200 задача с id=42
POST   /tasks       -> 201 задача создана
DELETE /tasks/42    -> 405 Method Not Allowed
GET    /unknown     -> 404 404 page not found
```

Обратите внимание: `405` и `404` роутер вернул сам, вы не написали для этого ни строчки.

**Задача:** добавьте маршрут `DELETE /tasks/{id}`, который отвечает `204` без тела.

### Упражнение 4. Своя схема в PostgreSQL (7 минут)

Откройте онлайн-песочницу Postgres и выполните по шагам.

```sql
CREATE TABLE users (
    id    BIGSERIAL PRIMARY KEY,
    name  TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE tasks (
    id         BIGSERIAL   PRIMARY KEY,
    title      TEXT        NOT NULL,
    done       BOOLEAN     NOT NULL DEFAULT FALSE,
    user_id    BIGINT      REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_tasks_user_id ON tasks(user_id);
```

Дальше:

1. Добавьте двух пользователей и по три задачи каждому.

   ```sql
   INSERT INTO users (name, email) VALUES ('Аня', 'anya@example.com') RETURNING id, name;
   INSERT INTO tasks (title, user_id) VALUES ('Сверстать карточку', 1) RETURNING id, done, created_at;
   ```

2. Выберите все невыполненные задачи вместе с именем автора:

   ```sql
   SELECT t.id, t.title, t.done, u.name AS author
   FROM tasks t
   JOIN users u ON u.id = t.user_id
   WHERE t.done = FALSE
   ORDER BY t.id;
   ```

3. Отметьте одну задачу выполненной и удалите другую:

   ```sql
   UPDATE tasks SET done = TRUE WHERE id = 1;
   DELETE FROM tasks WHERE id = 2;
   ```

4. **Сломайте базу.** Попробуйте вставить задачу несуществующему пользователю:

   ```sql
   INSERT INTO tasks (title, user_id) VALUES ('Задача-призрак', 999);
   ```

   Вы получите:

   ```text
   ERROR: insert or update on table "tasks" violates foreign key constraint "tasks_user_id_fkey"
   DETAIL: Key (user_id)=(999) is not present in table "users".
   ```

   Запомните эту формулировку. Это не поломка — это база не дала вам испортить данные.

5. **Посмотрите на SQL-инъекцию.** Представьте, что сервер склеил запрос строками, а пользователь прислал в поле поиска `' OR 1=1 --`:

   ```sql
   SELECT * FROM tasks WHERE title = '' OR 1=1 --';
   ```

   Запрос вернёт **всю таблицу**. Ровно поэтому в Go всегда `$1`, а не конкатенация.

---

## Часть 2. Рабочий пример целиком

Это полноценный REST API для задач. Скомпилирован и проверен — можно копировать файл за файлом.

### Эндпоинты

| Метод | Путь | Тело запроса | Ответ |
|---|---|---|---|
| `GET` | `/health` | — | `200` `{"status":"ok"}` |
| `GET` | `/tasks` | — | `200` массив задач |
| `GET` | `/tasks/{id}` | — | `200` задача / `404` / `400` |
| `POST` | `/tasks` | `{"title":"..."}` | `201` задача / `400` / `422` |
| `DELETE` | `/tasks/{id}` | — | `204` / `404` / `400` |

### schema.sql

```sql
CREATE TABLE IF NOT EXISTS users (
    id    BIGSERIAL PRIMARY KEY,
    name  TEXT NOT NULL,
    email TEXT NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS tasks (
    id         BIGSERIAL   PRIMARY KEY,
    title      TEXT        NOT NULL,
    done       BOOLEAN     NOT NULL DEFAULT FALSE,
    user_id    BIGINT      REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS idx_tasks_user_id ON tasks(user_id);
```

### Создание проекта

```bash
mkdir tasks-api && cd tasks-api
go mod init example.com/tasks
go get github.com/jackc/pgx/v5/pgxpool
```

### model.go

```go
package main

import "time"

type Task struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Done      bool      `json:"done"`
	CreatedAt time.Time `json:"created_at"`
}
```

### store.go — здесь и только здесь живёт SQL

```go
package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("not found")

type TaskStore struct {
	db *pgxpool.Pool
}

func NewTaskStore(db *pgxpool.Pool) *TaskStore {
	return &TaskStore{db: db}
}

func (s *TaskStore) List(ctx context.Context) ([]Task, error) {
	rows, err := s.db.Query(ctx,
		`SELECT id, title, done, created_at FROM tasks ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tasks := []Task{} // не var tasks []Task — иначе фронт получит null
	for rows.Next() {
		var t Task
		if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
			return nil, err
		}
		tasks = append(tasks, t)
	}
	return tasks, rows.Err()
}

func (s *TaskStore) Get(ctx context.Context, id int64) (Task, error) {
	var t Task
	err := s.db.QueryRow(ctx,
		`SELECT id, title, done, created_at FROM tasks WHERE id = $1`, id).
		Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return Task{}, ErrNotFound
	}
	if err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *TaskStore) Create(ctx context.Context, title string) (Task, error) {
	t := Task{Title: title}
	err := s.db.QueryRow(ctx,
		`INSERT INTO tasks (title) VALUES ($1) RETURNING id, done, created_at`, title).
		Scan(&t.ID, &t.Done, &t.CreatedAt)
	if err != nil {
		return Task{}, err
	}
	return t, nil
}

func (s *TaskStore) SetDone(ctx context.Context, id int64, done bool) error {
	tag, err := s.db.Exec(ctx, `UPDATE tasks SET done = $1 WHERE id = $2`, done, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (s *TaskStore) Delete(ctx context.Context, id int64) error {
	tag, err := s.db.Exec(ctx, `DELETE FROM tasks WHERE id = $1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
```

### handlers.go — здесь и только здесь живёт HTTP

```go
package main

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
)

type TaskHandler struct {
	store *TaskStore
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}

func writeError(w http.ResponseWriter, status int, msg string) {
	writeJSON(w, status, map[string]string{"error": msg})
}

func (h *TaskHandler) list(w http.ResponseWriter, r *http.Request) {
	tasks, err := h.store.List(r.Context())
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, tasks)
}

func (h *TaskHandler) getOne(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	task, err := h.store.Get(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusOK, task)
}

type createTaskRequest struct {
	Title string `json:"title"`
}

func (h *TaskHandler) create(w http.ResponseWriter, r *http.Request) {
	var req createTaskRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid JSON")
		return
	}

	req.Title = strings.TrimSpace(req.Title)
	if req.Title == "" {
		writeError(w, http.StatusUnprocessableEntity, "title is required")
		return
	}

	task, err := h.store.Create(r.Context(), req.Title)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	writeJSON(w, http.StatusCreated, task)
}

func (h *TaskHandler) remove(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		writeError(w, http.StatusBadRequest, "id must be a number")
		return
	}

	err = h.store.Delete(r.Context(), id)
	if errors.Is(err, ErrNotFound) {
		writeError(w, http.StatusNotFound, "task not found")
		return
	}
	if err != nil {
		writeError(w, http.StatusInternalServerError, "database error")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
```

### main.go — запуск, конфиг, маршруты, мидлвары

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

func mustEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func withCORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PATCH, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)
		log.Printf("%s %s %s", r.Method, r.URL.Path, time.Since(start))
	})
}

func main() {
	ctx := context.Background()

	dsn := mustEnv("DATABASE_URL", "postgres://app:secret@localhost:5432/app?sslmode=disable")
	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		log.Fatalf("cannot create pool: %v", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		log.Fatalf("cannot reach database: %v", err)
	}

	h := &TaskHandler{store: NewTaskStore(pool)}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	mux.HandleFunc("GET /tasks", h.list)
	mux.HandleFunc("POST /tasks", h.create)
	mux.HandleFunc("GET /tasks/{id}", h.getOne)
	mux.HandleFunc("DELETE /tasks/{id}", h.remove)

	addr := ":" + mustEnv("PORT", "8080")
	log.Printf("listening on %s", addr)

	srv := &http.Server{
		Addr:         addr,
		Handler:      withLogging(withCORS(mux)),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
	log.Fatal(srv.ListenAndServe())
}
```

### docker-compose.yml

```yaml
services:
  db:
    image: postgres:16
    environment:
      POSTGRES_USER: app
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: app
    ports:
      - "5432:5432"
    volumes:
      - pgdata:/var/lib/postgresql/data
      - ./schema.sql:/docker-entrypoint-initdb.d/schema.sql
    healthcheck:
      test: ["CMD-SHELL", "pg_isready -U app -d app"]
      interval: 5s
      timeout: 3s
      retries: 10

  adminer:
    image: adminer
    ports:
      - "8081:8080"
    depends_on:
      - db

volumes:
  pgdata:
```

`schema.sql` из папки проекта применится автоматически при **первом** создании базы. Если поменяли схему и хотите начать заново: `docker compose down -v` (это удалит данные).

Adminer открывается на **http://localhost:8081** — система `PostgreSQL`, сервер `db`, пользователь `app`, пароль `secret`, база `app`. Там видно, как ваши POST-запросы превращаются в строки.

### .env.example, .gitignore

```text
# .env.example — коммитим
DATABASE_URL=postgres://app:secret@localhost:5432/app?sslmode=disable
PORT=8080
```

```text
# .gitignore
.env
tasks-api
*.exe
```

### Запуск

```bash
docker compose up -d          # поднять базу и adminer
go run .                      # запустить сервер
```

Если база не поднялась, `go run .` честно упадёт с `cannot reach database` — так и задумано.

### Шпаргалка по curl

```bash
# жив ли сервер
curl -i localhost:8080/health

# список
curl localhost:8080/tasks

# создать
curl -i -X POST localhost:8080/tasks \
  -H 'Content-Type: application/json' \
  -d '{"title":"Подключить бэк к фронту"}'

# одна задача
curl -i localhost:8080/tasks/1

# удалить
curl -i -X DELETE localhost:8080/tasks/1

# только код ответа
curl -o /dev/null -w "%{http_code}\n" localhost:8080/tasks/999
```

Флаг `-i` показывает заголовки и код ответа. Без него вы видите только тело и не заметите, что вернулась ошибка.

Реальные ответы этого примера:

```text
POST /tasks {"title":"Подключить бэк к фронту"}  -> 201 {"id":3,"title":"...","done":false,"created_at":"..."}
POST /tasks {"title":"  "}                       -> 422 {"error":"title is required"}
POST /tasks не-json                              -> 400 {"error":"invalid JSON"}
GET  /tasks/999                                  -> 404 {"error":"task not found"}
GET  /tasks/abc                                  -> 400 {"error":"id must be a number"}
DELETE /tasks/2                                  -> 204 (без тела)
```

---

## Часть 3. Почему это работает

Разбор механики: что это такое, из чего сделано и почему работает именно так. Порядок совпадает с лекцией — если на паре мы что-то пропустили или вы не успели записать, ищите здесь.

Весь вывод программ в этом разделе — **настоящий**, полученный прогоном.

**Оглавление раздела**

- 3.1–3.4 — инструменты и язык: модуль, версии, формат-глаголы, строки и байты
- 3.5–3.8 — HTTP: текст протокола, URL, идемпотентность, TCP
- 3.9–3.12 — как программа становится сервером: порт, горутины, `ListenAndServe`, гонки
- 3.13–3.16 — интерфейсы: `Handler`, `HandlerFunc`, `io.Writer`, ловушка `nil`
- 3.17–3.21 — **mux: что это и как устроен**
- 3.22–3.27 — хендлер: матрёшка, `Request`, `ResponseWriter`, порядок записи
- 3.28–3.31 — JSON: теги, рефлексия, `Decode`
- 3.32–3.37 — PostgreSQL: `PRIMARY KEY`, `BIGSERIAL`, внешние ключи, индексы, `EXPLAIN`, транзакции
- 3.38–3.43 — Go и база: драйвер, DSN, пул, протокол, `Scan`, `NULL`
- 3.44–3.48 — ошибки, `defer`, `context`, Docker, CORS

---

### 3.1. Пакет, модуль и `go.mod`

**Пакет** — папка с `.go`-файлами, у которых в первой строке одно и то же `package имя`. Файлы одного пакета **видят друг друга без импортов** — поэтому `main.go` спокойно вызывает функции из `store.go`.

**Модуль** — дерево пакетов с одним `go.mod` в корне. Это единица версионирования и распространения.

```bash
go mod init example.com/tasks
go get github.com/jackc/pgx/v5/pgxpool
```

```text
// go.mod
module example.com/tasks     ← как ВАС будут импортировать
go 1.27                      ← минимальная версия языка

require (
    github.com/jackc/pgx/v5 v5.7.6
    github.com/jackc/puddle/v2 v2.2.2 // indirect
)
```

`// indirect` значит «нужен не вам, а вашей зависимости».

Путь импорта — не URL, но обычно совпадает с ним: `github.com/jackc/pgx/v5` Go честно пойдёт качать с GitHub. Отсюда и `/v5` в конце — начиная со второй мажорной версии номер входит в путь, чтобы v4 и v5 могли сосуществовать в одном проекте.

### 3.2. `go.sum` и версии

`go.mod` говорит **какие** версии нужны, `go.sum` хранит **хеши** того, что было скачано:

```text
github.com/jackc/pgx/v5 v5.7.6 h1:rZa0f...
github.com/jackc/pgx/v5 v5.7.6/go.mod h1:0Ss...
```

При каждой сборке Go сверяет скачанное с хешем: если содержимое версии изменилось — сборка падает, а не тихо собирает чужой код. **Оба файла коммитятся, руками не правятся.** Нужно обновиться — `go get -u` или `go mod tidy`.

Версии семантические: `v5.7.6` = мажорная.минорная.патч. Мажорная меняется, когда ломают совместимость, — поэтому она и попадает в путь импорта.

### 3.3. Формат-глаголы

```go
t := Task{ID: 1, Title: "Сверстать карточку"}
fmt.Printf("%v\n",  t)   // {1 Сверстать карточку}
fmt.Printf("%+v\n", t)   // {ID:1 Title:Сверстать карточку}
fmt.Printf("%#v\n", t)   // main.Task{ID:1, Title:"Сверстать карточку"}
fmt.Printf("%T\n",  t)   // main.Task
fmt.Printf("%q\n", "текст")  // "текст"
```

| Глагол | Что делает | Когда нужен |
|---|---|---|
| `%v` | значение как есть | обычный вывод |
| `%+v` | структура с именами полей | **отладка** — самый полезный |
| `%#v` | как валидный код на Go | скопировать значение в тест |
| `%T` | тип значения | «что вообще в этом интерфейсе лежит?» |
| `%q` | строка в кавычках | видно пробелы и пустые строки |
| `%w` | завернуть ошибку (только в `fmt.Errorf`) | чтобы работал `errors.Is` |

Практический совет: когда что-то «не сохраняется», первым делом напечатайте `%+v` разобранной структуры. В девяти случаях из десяти там пусто, и сразу видно, что JSON не разобрался.

### 3.4. Строка — это байты, а не буквы

```go
s := "Привет"
fmt.Printf("len(s)  (байты): %d\n", len(s))          // 12
fmt.Printf("len([]rune(s)) (буквы): %d\n", len([]rune(s)))  // 6
fmt.Printf("первые 4 байта: %v\n", []byte(s)[:4])    // [208 159 209 128]
fmt.Printf("s[0]: %d — это байт\n", s[0])            // 208
```

Строка в Go — последовательность байт в UTF-8. Латинская буква занимает 1 байт, кириллическая — 2, эмодзи — до 4. `rune` — один символ Unicode.

Отсюда три следствия:

- `Content-Length: 41` при видимых 25 символах — в заголовке байты, а не символы;
- `Write([]byte)` принимает байты, потому что сеть и файлы оперируют байтами;
- `s[0:10]` на кириллице может разрезать букву пополам. Ограничивайте длину по `[]rune`.

### 3.5. HTTP — это текст

```http
POST /tasks HTTP/1.1
Host: api.example.com
Content-Type: application/json
Content-Length: 41

{"title":"Подключить бэк к фронту"}
```

Четыре части: стартовая строка · заголовки · **пустая строка** · тело.

Пустая строка обязательна, потому что сервер читает байты из сокета и не знает заранее, где кончатся заголовки. После неё он смотрит в `Content-Length` и читает ровно столько байт тела. Никакой другой границы в протоколе нет: JSON на уровне HTTP — просто строка, и что это JSON, получатель узнаёт только из `Content-Type`.

Заголовки в Go — буквально `map[string][]string`:

```go
type Header map[string][]string        // из стандартной библиотеки

r.Header.Get("Content-Type")           // первое значение или ""
w.Header().Set("Content-Type", "application/json")
w.Header().Add("Set-Cookie", "a=1")    // добавить ещё одно
```

`Get` и `Set` приводят имя к каноническому виду, поэтому `content-type` и `Content-Type` — один ключ. Если лезть в map напрямую, регистр вдруг начнёт иметь значение.

### 3.6. URL по частям

```go
u, _ := url.Parse("https://api.example.com:8443/tasks/42?done=false&page=2#top")
```

```text
Scheme   = "https"
Host     = "api.example.com:8443"
Hostname = "api.example.com"
Port     = "8443"
Path     = "/tasks/42"
RawQuery = "done=false&page=2"
Fragment = "top"                ← на сервер НЕ отправляется
Query().Get("page") = "2"
```

**Путь или query?** Путь отвечает на вопрос «что за ресурс» (`/tasks/42`), query — «как его отфильтровать и порезать» (`?done=false&page=2`). Идентификатор в query (`?id=42`) — признак, что автор не разобрался.

Fragment (после `#`) на сервер не уходит вообще — это чисто браузерная штука.

### 3.7. Идемпотентность

Запрос **идемпотентен**, если повторить его десять раз — то же самое, что выполнить один раз. От этого зависит, можно ли автоматически повторить запрос при обрыве связи.

| Метод | Безопасный | Идемпотентный | Почему |
|---|---|---|---|
| `GET` | да | да | только читает |
| `PUT` | нет | да | кладёт объект целиком |
| `DELETE` | нет | да | второй раз удалять уже нечего |
| `PATCH` | нет | как повезёт | `{"done":true}` — да; «прибавь 1» — нет |
| `POST` | нет | **нет** | каждый раз создаёт новую запись |

Поэтому в серьёзных API на `POST` присылают **ключ идемпотентности** — случайный идентификатор запроса; сервер, увидев тот же ключ второй раз, возвращает прежний результат вместо создания дубля. Так работают платёжные системы.

### 3.8. Под HTTP лежит TCP

HTTP описывает, **что написать** в поток байт. Сам поток обеспечивает TCP: доставку, целостность и порядок.

Установка соединения стоит денег: трёхэтапное рукопожатие — минимум один полный оборот до сервера и обратно, для HTTPS сверху ещё TLS. Поэтому соединения переиспользуют (**keep-alive**), и поэтому в Go принято держать один `http.Client`, а не создавать новый на каждый запрос.

Та же логика у базы данных — отсюда пул соединений (3.40).

### 3.9. Что значит «слушать порт»

```go
ln, err := net.Listen("tcp", ":8080")   // 1. занять порт у ОС
for {
	conn, err := ln.Accept()             // 2. ждать подключения
	go handle(conn)                      // 3. обслужить и вернуться к ожиданию
}
```

Порт — просто число от 1 до 65535, которым ОС различает программы на одной машине. Занять его может только одна: отсюда `address already in use`.

`Accept` блокируется — программа буквально стоит и ждёт.

`":8080"` без хоста значит «слушать на всех интерфейсах». `"localhost:8080"` — только с этой машины; внутри Docker это означает, что снаружи к вам не достучаться.

### 3.10. Что такое горутина

Горутина — функция, выполняющаяся одновременно с остальными, но **не поток операционной системы**. Планировщик Go раскладывает тысячи горутин на несколько потоков ОС и переключает их сам, не обращаясь к ядру.

```text
горутин до старта: 1,       память: 63 КБ
горутин после:     100001,  память: 57552 КБ
запуск 100000 горутин занял 249.7ms
стеки горутин: ~2050 байт на горутину
```

- **2 КБ** стартового стека против 1–8 **МБ** у потока ОС;
- сто тысяч горутин поднялись за четверть секунды; сто тысяч потоков ОС не поднялись бы вообще;
- переключение не требует системного вызова.

Поэтому модель «горутина на каждый запрос» — норма, а не расточительство. И поэтому в Go можно писать обычный последовательный код без колбэков и `async/await`: блокирующий вызов блокирует только свою горутину.

### 3.11. Что делает `ListenAndServe`

```text
1. net.Listen("tcp", addr)                — занять порт
2. в бесконечном цикле: ln.Accept()       — дождаться соединения
3. на каждое соединение — go c.serve()    — ОТДЕЛЬНАЯ ГОРУТИНА
4. внутри горутины:
     прочитать байты до пустой строки     — это заголовки
     разобрать их в структуру http.Request
     прочитать Content-Length байт тела
     спросить у mux, какой хендлер отвечает за этот метод и путь
     вызвать хендлер, передав ему w и r
     отправить байты ответа обратно
```

Два вывода. Первый: ваш хендлер — не колбэк, а обычная функция, которую обычным вызовом вызывает библиотечный код. Второй: `ListenAndServe` блокирует навсегда, поэтому её оборачивают в `log.Fatal(...)` — иначе ошибка запуска пройдёт молча.

### 3.12. Гонки в хендлере

Сто одновременных запросов — сто горутин, выполняющих одну и ту же вашу функцию одновременно.

```go
var counter int  // одна переменная на все запросы

func handler(w http.ResponseWriter, r *http.Request) {
	counter++    // НЕ одна операция: прочитать, прибавить, записать
}
```

```text
$ go run -race .
WARNING: DATA RACE
Found 1 data race(s)

запросов: 200
counter = 200     ← и при этом результат «правильный»
```

Вот чем это коварно: на ноутбуке гонка почти никогда не проявляется (пять запусков — пять раз ровно 200), а под нагрузкой два счёта совпадут по времени и одно увеличение потеряется. Ловится только `go run -race`.

**Вывод для проекта:** не держите изменяемое состояние в глобальных переменных. Всё, что нужно хендлеру, — в базу.

### 3.13. Интерфейс — это пара «тип + значение»

Интерфейсная переменная хранит две вещи: какой конкретный тип в ней лежит и само значение. Отсюда работает `%T` и вызов метода.

В Go интерфейсы удовлетворяются **неявно**: нигде не пишут `implements`. Если у типа есть все нужные методы — он подходит.

Следствие для дизайна: интерфейсы объявляют там, где их **используют**, а не там, где реализуют. Поэтому они маленькие — один-два метода: `Handler`, `Writer`, `error`.

### 3.14. Ловушка `nil`-интерфейса

Раз интерфейс это пара, то он пуст только когда пусты **обе** части:

```text
пустой интерфейс:       <nil>         | == nil: true
положили *bytes.Buffer: *bytes.Buffer | == nil: false
положили nil-указатель: *bytes.Buffer | == nil: false  ← ЛОВУШКА
```

Где укусит: функция объявлена как `func f() error`, внутри есть `var myErr *MyError`, оставшийся `nil`, и его возвращают. Снаружи `if err != nil` **срабатывает**, хотя ошибки не было. Правило: возвращайте `nil` буквально, а не переменную конкретного типа-указателя.

### 3.15. `http.Handler` — один интерфейс на всю библиотеку

```go
// вот он целиком, из стандартной библиотеки
type Handler interface {
	ServeHTTP(w ResponseWriter, r *Request)
}
```

Один метод. Его реализуют ваш хендлер, `ServeMux`, мидлвара, файловый сервер. Хендлером может быть и структура:

```go
type greeter struct{ name string }

func (g greeter) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "привет из структуры, %s", g.name)
}

var h http.Handler = greeter{name: "Аня"}
```

В `net/http` нет системы плагинов, нет регистрации, нет рефлексии — есть один интерфейс, и на нём всё стыкуется.

### 3.16. `http.HandlerFunc` — три строки, объясняющие всю «магию»

Мы пишем хендлеры функциями, а интерфейс требует метод. Соединяют их три строки стандартной библиотеки:

```go
type HandlerFunc func(ResponseWriter, *Request)

func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)          // метод просто вызывает саму функцию
}
```

Это объявление именованного типа на основе функции и метод, объявленный **на самом типе функции**. В Go метод можно повесить на любой именованный тип своего пакета — не только на структуру.

```go
var h http.Handler = http.HandlerFunc(plain)   // приведение типа, а не вызов
```

И `mux.HandleFunc(p, f)` внутри — просто `mux.Handle(p, HandlerFunc(f))`.

Тот же приём с `io.Writer`:

```go
type Writer interface {              // io.Writer
	Write(p []byte) (n int, err error)
}
```

У `http.ResponseWriter` метод `Write` есть — значит он и есть `io.Writer`. Поэтому `json.NewEncoder(w)`, `fmt.Fprintln(w, ...)` и `io.Copy(w, file)` работают.

---

### 3.17. Что такое `mux`

`mux` — сокращение от **multiplexer**, мультиплексор. Термин из электроники: устройство, у которого **один вход и много выходов**, и которое по управляющему сигналу решает, на какой выход направить сигнал.

```text
                        ┌─────────┐   ──→  h.list      ← GET /tasks
   HTTP-запрос  ──────→ │   mux   │   ──→  h.create    ← POST /tasks
   (один вход)          │метод+путь│  ──→  h.getOne    ← GET /tasks/{id}
                        └─────────┘   ──→  h.remove    ← DELETE /tasks/{id}
```

Управляющий сигнал здесь — **метод и путь запроса**. Один поток входящих запросов, десяток хендлеров, и кто-то должен на каждый запрос решить, кому его отдать. Этот кто-то и называется mux.

Синонимы, которые встретятся в статьях: **роутер**, **router**, **диспетчер**. Это одно и то же. В Go прижилось `mux`, потому что так называется тип в стандартной библиотеке — `ServeMux`.

### 3.18. Что у mux внутри

Регистрация маршрута — это **запись в таблицу**. Никакого кода при этом не выполняется, хендлер просто запоминается:

```go
mux := http.NewServeMux()

mux.HandleFunc("GET /tasks",         h.list)
mux.HandleFunc("POST /tasks",        h.create)
mux.HandleFunc("GET /tasks/{id}",    h.getOne)
mux.HandleFunc("DELETE /tasks/{id}", h.remove)
```

После этих четырёх строк внутри `mux` лежит примерно такое:

```text
"GET /tasks"          → функция h.list
"POST /tasks"         → функция h.create
"GET /tasks/{id}"     → функция h.getOne
"DELETE /tasks/{id}"  → функция h.remove
```

**Две обязанности и всё:**

1. **Запомнить:** `HandleFunc(шаблон, хендлер)` — положить пару в таблицу.
2. **Найти:** `ServeHTTP(w, r)` — взять метод и путь из запроса, найти подходящий шаблон, вызвать его хендлер.

И сам mux — **тоже `Handler`**: у него есть метод `ServeHTTP`, значит его можно передать в `http.Server` или завернуть в мидлвару, как любой другой хендлер.

### 3.19. Свой mux за 30 строк

Это работает. Скопируйте в Playground и запустите.

```go
package main

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
)

// Упрощённый роутер: ровно то, что делает http.ServeMux, без деталей
type myMux struct {
	routes map[string]http.HandlerFunc // ключ: "GET /tasks"
}

func newMyMux() *myMux {
	return &myMux{routes: map[string]http.HandlerFunc{}}
}

func (m *myMux) HandleFunc(pattern string, h http.HandlerFunc) {
	m.routes[pattern] = h // регистрация — это просто запись в map
}

// вот он, весь роутинг
func (m *myMux) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	key := r.Method + " " + r.URL.Path

	if h, ok := m.routes[key]; ok {
		h(w, r) // нашли — вызываем
		return
	}

	// путь есть, но с другим методом?
	for pattern := range m.routes {
		if strings.HasSuffix(pattern, " "+r.URL.Path) {
			http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
			return
		}
	}

	http.NotFound(w, r)
}

func main() {
	mux := newMyMux()
	mux.HandleFunc("GET /tasks", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "список задач")
	})
	mux.HandleFunc("POST /tasks", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		fmt.Fprint(w, "создана")
	})

	try := func(method, path string) {
		rec := httptest.NewRecorder()
		mux.ServeHTTP(rec, httptest.NewRequest(method, path, nil))
		fmt.Printf("%-7s %-8s -> %d %s\n", method, path, rec.Code, strings.TrimSpace(rec.Body.String()))
	}

	try("GET", "/tasks")
	try("POST", "/tasks")
	try("DELETE", "/tasks")
	try("GET", "/unknown")
}
```

```text
GET     /tasks   -> 200 список задач
POST    /tasks   -> 201 создана
DELETE  /tasks   -> 405 Method Not Allowed
GET     /unknown -> 404 404 page not found
```

Наш самодельный mux уже ведёт себя как настоящий — отдаёт 405 на неизвестный метод и 404 на неизвестный путь. **Никакой магии в роутере нет: `map`, поиск по ключу и вызов функции.**

Обратите внимание: `myMux` нигде не говорит, что он `http.Handler`. У него просто есть метод `ServeHTTP` — и этого достаточно, интерфейсы в Go неявные.

> **Упражнение на дом:** добавьте в этот mux поддержку `{id}` — чтобы `GET /tasks/{id}` ловил `/tasks/42`, а значение можно было получить в хендлере.

### 3.20. Чем настоящий `ServeMux` отличается от нашего

| Возможность | Наш | Настоящий |
|---|---|---|
| Точное совпадение метода и пути | да | да |
| Параметры пути `{id}` | нет | **да**, через `r.PathValue("id")` |
| Префиксы `/static/` | нет | да, ловит всё вложенное |
| Приоритет более специфичного шаблона | нет | да, и не зависит от порядка регистрации |
| Заголовок `Allow` при `405` | нет | да |
| Чистка пути (`/a//b/../c` → `/a/c`) | нет | да, с редиректом `301` |
| Поиск за один проход, а не перебором | нет | да, дерево вместо цикла |

Идея ровно та же: таблица шаблонов, поиск по методу и пути, вызов найденного хендлера. Всё остальное — аккуратность и скорость.

### 3.21. Шаблоны, приоритет и `DefaultServeMux`

| Шаблон | Что ловит |
|---|---|
| `/tasks` | любой метод на `/tasks` |
| `GET /tasks` | только `GET`; `POST` получит `405` |
| `GET /tasks/{id}` | один сегмент: `/tasks/42`, но не `/tasks/42/comments` |
| `GET /files/{path...}` | весь остаток пути |
| `/static/` | слеш на конце = префикс, всё вложенное |
| `GET /{$}` | **только** корень `/` |
| `GET /` | **всё подряд** — это префикс из одного слеша |

Главная ловушка — последние две строки. `"/"` это не «главная страница», а «любой путь». Для точного корня есть `"/{$}"`.

**Приоритет:** выигрывает более специфичный шаблон, порядок строк в коде значения не имеет.

```go
mux.HandleFunc("GET /tasks/{id}", h.getOne)
mux.HandleFunc("GET /tasks/new",  h.newForm)   // зарегистрирован ПОЗЖЕ
```

`GET /tasks/new` попадёт в `h.newForm` — точный литерал специфичнее параметра. А если два шаблона одинаково специфичны (`GET /a/{x}/c` и `GET /a/b/{y}`), `ServeMux` считает это **конфликтом** и роняет программу паникой при регистрации — сразу при старте, а не на живом трафике.

**`DefaultServeMux`.** Когда вы пишете `http.HandleFunc(...)` и `http.ListenAndServe(":8080", nil)`, используется глобальная переменная пакета `net/http`:

```go
var DefaultServeMux = &defaultServeMux

func HandleFunc(pattern string, handler func(ResponseWriter, *Request)) {
	DefaultServeMux.HandleFunc(pattern, handler)
}
```

В неё может дописать **любая** подключённая библиотека. Классика: подключили `net/http/pprof` для профилирования — он в своей `init()` молча зарегистрировал `/debug/pprof/` в вашем сервере, и снаружи стали доступны дампы памяти вашего приложения.

**Как правильно:** всегда свой `mux := http.NewServeMux()` и явная передача его в сервер.

---

### 3.22. Полная картина: кто кого вызывает

```text
http.Server               принял соединение, разобрал запрос, вызвал .ServeHTTP у своего Handler
  └─ withLogging          засёк время, вызвал .ServeHTTP у next
      └─ withCORS         поставил заголовки, вызвал .ServeHTTP у next
          └─ mux          нашёл шаблон "GET /tasks", вызвал .ServeHTTP у найденного
              └─ HandlerFunc(h.list)   вызвала саму функцию h.list
                  └─ h.list           ваш код: спросил store, записал JSON в w
```

Каждая стрелка — обычный вызов метода `ServeHTTP`. Все пять уровней реализуют один и тот же интерфейс `Handler`, поэтому и складываются друг в друга.

Отсюда понятно, как отлаживать: не сработал CORS — посмотрите, обёрнут ли `mux`; не попал в хендлер — проблема в шаблоне; не видно в логах — логирование стоит внутри чего-то, что обрывает запрос раньше.

### 3.23. Мидлвара — это матрёшка из обычных вызовов

```go
func wrap(name string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println(name, "— до")
		next.ServeHTTP(w, r)      // передаём управление дальше
		fmt.Println(name, "— после")
	})
}

h := wrap("логирование", wrap("CORS", mux))
```

```text
логирование — до
CORS — до
  сам хендлер
CORS — после
логирование — после
```

Никакой «цепочки обработчиков» как отдельного механизма не существует — это обычный стек вызовов. Следствия: мидлвара может **не вызвать** `next` и оборвать запрос (так работает проверка авторизации и наш `OPTIONS` в CORS); код до `next` видит запрос, код после — уже отправленный ответ; порядок обёрток важен.

### 3.24. `Request`: что где лежит

| Поле | Что в нём |
|---|---|
| `r.Method` | `"GET"` |
| `r.URL.Path` | `"/tasks/42"` |
| `r.URL.Query().Get("page")` | параметр из query |
| `r.PathValue("id")` | параметр из шаблона mux |
| `r.Header.Get("Authorization")` | заголовок |
| `r.Body` | тело как **поток** |
| `r.Context()` | «клиент ещё ждёт?» |
| `r.RemoteAddr` | адрес клиента |

**Почему `r` со звёздочкой, а `w` без.** `Request` — большая структура, копировать её на каждый вызов незачем, и библиотека хочет, чтобы все видели один и тот же запрос. `ResponseWriter` — интерфейс, внутри которого уже лежит указатель; писать `*http.ResponseWriter` не нужно и вредно.

Общее правило Go: звёздочка у интерфейса почти всегда признак ошибки.

### 3.25. `r.Body` — это поток

Тело запроса — `io.ReadCloser`: труба, из которой байты вычитываются по мере надобности. Поэтому прочитать его можно **один раз**:

```text
первый  Decode: {Title:Сверстать карточку}, ошибка: <nil>
второй  Decode: {Title:}, ошибка: EOF
```

Если тело нужно дважды (залогировать и разобрать) — `io.ReadAll` в переменную, потом `r.Body = io.NopCloser(bytes.NewReader(body))`.

И ограничивайте размер: `r.Body = http.MaxBytesReader(w, r.Body, 1<<20)` — иначе любой может прислать гигабайт.

### 3.26. `ResponseWriter` доделывает работу за вас

```go
type ResponseWriter interface {
	Header() Header                  // словарь заголовков, пока не отправлен
	Write([]byte) (int, error)       // записать байты тела
	WriteHeader(statusCode int)      // зафиксировать код и отправить заголовки
}
```

```text
без заголовка:               код=200  Content-Type="text/plain; charset=utf-8"
с заголовком:                код=200  Content-Type="application/json; charset=utf-8"
заголовок после WriteHeader: код=201  Content-Type=""
```

- Не вызвали `WriteHeader` — подставится `200`.
- Не поставили `Content-Type` — Go **угадает по первым 512 байтам тела** (`http.DetectContentType`) и решит, что ваш JSON это простой текст.
- Поставили заголовок после `WriteHeader` — поздно, заголовки уже ушли по сети.

Правильный порядок всегда один: `Header().Set` → `WriteHeader` → `Write`.

### 3.27. Забытый `return`

```text
http: superfluous response.WriteHeader call from main.writeJSON
```

Первый вызов уже отправил заголовки, второй отклонён — но **тело всё равно дописалось** в тот же ответ. Клиент получает склеенные два JSON, которые не разберёт ни один парсер. Правило: каждая ветка, которая отвечает клиенту, заканчивается `return`.

**`strconv.ParseInt(s, 10, 64)`** — строка, основание системы счисления, сколько бит должно влезть:

```text
"42"                   -> 42, err = <nil>
"abc"                  -> 0,  invalid syntax
"99999999999999999999" -> 9223372036854775807, value out of range
""                     -> 0,  invalid syntax
"ff" (основание 16)    -> 255
```

Возвращается **и значение, и ошибка**. Проигнорируете `err` — пойдёте в базу за задачей номер ноль.

**`any`** в `writeJSON(w, status int, v any)` — это псевдоним для `interface{}`, «интерфейс без единого метода», которому удовлетворяет любой тип. Появился в Go 1.18 просто чтобы не писать фигурные скобки.

---

### 3.28. Теги структур: кто их читает

Backticks после поля — не синтаксис JSON. Это произвольная строка, которую компилятор кладёт в метаданные типа, а `encoding/json` достаёт **в рантайме** через рефлексию:

```go
type Task struct {
	Title string `json:"title" db:"title_col"`
}

ft, _ := reflect.TypeOf(Task{}).FieldByName("Title")
fmt.Printf("%q\n", ft.Tag)              // "json:\"title\" db:\"title_col\""
fmt.Printf("%q\n", ft.Tag.Get("json"))  // "title"
```

**Рефлексия** — способность программы в рантайме узнать про значение, какого оно типа и какие у него поля, не зная типа на этапе компиляции. Так `json.Marshal` умеет сериализовать любую структуру.

Отсюда: опечатка в теге (`jsn:"title"`) — **не ошибка компиляции**; теги разных библиотек живут в одной строке через пробел; вы увидите не ошибку, а неправильный JSON. `go vet` умеет ловить кривые теги.

По той же причине поле с маленькой буквы не попадает в JSON: `encoding/json` — **другой пакет**, и рефлексия честно соблюдает правила доступа.

### 3.29. `omitempty`, `-` и указатели

```go
type Task struct {
	ID     int64  `json:"id"`
	Title  string `json:"title"`
	Note   string `json:"note,omitempty"`  // исчезнет из JSON, если пустое
	Secret string `json:"-"`               // не попадёт в JSON никогда
	Done   *bool  `json:"done"`            // различает «прислали false» и «не прислали»
}
```

```text
Note и Done пустые:     {"id":1,"title":"Сверстать карточку","done":null}
Note и Done заполнены:  {"id":1,"title":"Сверстать карточку","note":"срочно","done":true}
```

Осторожно с `omitempty` на `bool`: `false` — валидное значение, но `omitempty` его выбросит.

### 3.30. `Decode` и его тонкости

`Decode(&req)` требует указатель, потому что должен **записать** в вашу переменную.

**Неизвестные поля молча игнорируются:** клиент пришлёт `{"titel":"..."}` — ошибки не будет, поле останется пустым. Это причина половины багов «я отправляю, а не сохраняется». Строгий режим:

```go
dec := json.NewDecoder(r.Body)
dec.DisallowUnknownFields()
```

**Отсутствующее поле ошибкой не считается** — оно просто останется нулевым. Поэтому «обязательность» поля проверяете вы сами, а не `Decode`.

### 3.31. Структура запроса ≠ модель

```go
var t Task
json.NewDecoder(r.Body).Decode(&t)   // ОПАСНО
```

Клиент присылает `{"id":999,"done":true,"created_at":"2001-01-01T00:00:00Z"}` — и вы покорно пытаетесь это сохранить. Это отдельный класс уязвимостей, называется mass assignment.

```go
type createTaskRequest struct {      // ПРАВИЛЬНО
	Title string `json:"title"`
}
```

Всё остальное, что бы клиент ни прислал, просто не попадёт в структуру: полей для него нет.

**Валидация — до базы:**

```go
req.Title = strings.TrimSpace(req.Title)
if req.Title == "" {
	writeError(w, http.StatusUnprocessableEntity, "title is required")
	return
}
if len([]rune(req.Title)) > 200 {
	writeError(w, http.StatusUnprocessableEntity, "title is too long")
	return
}
```

`TrimSpace` до проверки обязателен: строка из одних пробелов не пустая для `== ""`, но пустая по смыслу. И `[]rune`, а не `len(строка)`, — иначе 200 байт кириллицы это всего сто букв.

---

### 3.32. База — отдельная программа

PostgreSQL — не библиотека внутри вашего процесса, а **сервер**, который слушает порт 5432, ровно как ваш Go-сервер слушает 8080. Ваше приложение для базы такой же клиент, как браузер для вас.

Отсюда: задержка на каждый запрос (один `JOIN` вместо ста запросов — вот откуда скорость), свой двоичный протокол поверх TCP, и необходимость пула соединений.

Гарантия «данные переживут выключение питания» обеспечивается **журналом упреждающей записи** (WAL): прежде чем менять данные, база записывает на диск «я собираюсь сделать вот это».

### 3.33. Что создают `PRIMARY KEY` и `UNIQUE`

```text
=> \d users
Indexes:
    "users_pkey" PRIMARY KEY, btree (id)
    "users_email_key" UNIQUE CONSTRAINT, btree (email)
Referenced by:
    TABLE "tasks" CONSTRAINT "tasks_user_id_fkey"
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
```

Оба ограничения — это **индекс** (B-дерево) плюс правило «значения не повторяются». Отсюда:

- поиск по `id` и `email` быстрый **бесплатно** — индекс уже есть;
- проверка уникальности при вставке — это поиск по тому же индексу;
- `PRIMARY KEY` = `UNIQUE` + `NOT NULL`, и он один на таблицу.

**А внешний ключ индекс НЕ создаёт.** Видно в выводе: у `users` есть запись «на меня ссылается `tasks`», но индекса на `tasks.user_id` не появилось. Его создают руками — это самая частая причина «у нас всё тормозит на джойнах».

### 3.34. `BIGSERIAL` изнутри

```text
=> \d tasks
   Column   |           Type           | Nullable |              Default
------------+--------------------------+----------+-----------------------------------
 id         | bigint                   | not null | nextval('tasks_id_seq'::regclass)
 done       | boolean                  | not null | false
 created_at | timestamp with time zone | not null | now()
```

Тип колонки — обычный `bigint`, «автоинкремент» — это `DEFAULT nextval(...)`, вызов функции у последовательности, которая живёт в базе сама по себе.

Последовательность **не откатывается**: если транзакция с `INSERT` откатится, номер всё равно израсходован. Поэтому в `id` бывают дыры, и поэтому `MAX(id)` нельзя использовать как «сколько у нас записей» — считает только `COUNT(*)`.

### 3.35. Внешний ключ: когда проверяется

| Операция | Что делает база |
|---|---|
| `INSERT` в `tasks` | ищет такой `id` в `users`; не нашла — отклоняет всю операцию |
| `UPDATE` `tasks.user_id` | то же самое |
| `DELETE` из `users` | смотрит, что сказано в `ON DELETE` |

- `ON DELETE CASCADE` — удалить и все задачи пользователя. Удобно и опасно: молча уносит данные.
- `ON DELETE RESTRICT` — запретить удаление, пока есть задачи. Безопаснее по умолчанию.
- `ON DELETE SET NULL` — оставить задачи, обнулив автора.

`DELETE` пользователя при `CASCADE` вынужден найти все его задачи — и **без индекса на `tasks.user_id`** это полный просмотр таблицы на каждое удаление.

### 3.36. Индексы и `EXPLAIN ANALYZE`

Индекс — B-дерево: значения хранятся отсортированными, и поиск за несколько шагов вместо полного перебора. Двести тысяч строк — порядка 18 сравнений вместо 200 000.

Таблица `tasks`, 200 000 строк, ищем одну по названию:

```text
БЕЗ ИНДЕКСА
Seq Scan on tasks  (actual time=10.990..18.177 rows=1)
  Filter: (title = 'Задача 123456')
  Rows Removed by Filter: 199999
Execution Time: 18.219 ms

С ИНДЕКСОМ
Index Scan using idx_tasks_title on tasks  (actual time=0.035..0.035 rows=1)
  Index Cond: (title = 'Задача 123456')
Execution Time: 0.076 ms
```

**18.2 мс → 0.076 мс, в 240 раз быстрее**, и разрыв растёт вместе с таблицей.

Что читать в плане: `Seq Scan` — полный просмотр, `Index Scan` — поиск по индексу. **`Rows Removed by Filter`** — сколько строк база прочитала зря; большое число здесь всегда значит «не хватает индекса». `actual time` — реально потраченное время; `cost` — условные единицы планировщика, между запросами их сравнивать бессмысленно.

Индекс — компромисс: быстрее чтение, медленнее запись (его надо обновлять при каждой вставке). Поэтому его добавляют под конкретный запрос, а не «на всякий случай».

**Проблема N+1.** Взять список задач, а потом в цикле запросить автора для каждой — это 101 поход в базу вместо одного. Решение — `JOIN`. Название от количества запросов: один за списком плюс по одному на каждый из N элементов.

### 3.37. Транзакции

Несколько операций как одна: списать у одного и зачислить другому обязаны произойти **обе или ни одной**.

```sql
-- если хотите выполнить это в песочнице, сначала заведите таблицу:
-- CREATE TABLE accounts (id BIGSERIAL PRIMARY KEY, balance NUMERIC(10,2) NOT NULL);
-- INSERT INTO accounts (balance) VALUES (500), (0);

BEGIN;
  UPDATE accounts SET balance = balance - 100 WHERE id = 1;
  UPDATE accounts SET balance = balance + 100 WHERE id = 2;
COMMIT;      -- или ROLLBACK, и тогда как будто ничего не было
```

Четыре гарантии (ACID): **атомарность** — целиком или никак; **согласованность** — ограничения схемы соблюдены; **изолированность** — параллельные транзакции не видят половинчатых состояний; **долговечность** — после `COMMIT` данные переживут выключение питания.

Где нужно в вашем проекте: везде, где одно действие меняет **больше одной** таблицы — заказ и его позиции, выдача книги и изменение остатка, регистрация и создание профиля.

**Пагинация.** `LIMIT`/`OFFSET` прост, но у него две проблемы: `OFFSET 100000` заставляет базу прочитать и выбросить 100 000 строк, а вставка новой записи сдвигает границы страниц. Решение для больших данных — пагинация по курсору (`WHERE id > 1040 ORDER BY id LIMIT 20`). **Для учебного проекта `LIMIT`/`OFFSET` совершенно нормален** — важно знать, где он ломается.

---

### 3.38. Драйвер и DSN

У PostgreSQL свой двоичный протокол поверх TCP. **Драйвер** — библиотека, которая умеет на нём разговаривать.

- `database/sql` — стандартный интерфейс для **любой** SQL-базы. Сам ничего не умеет, нужен драйвер. Плюс: сменить базу почти не трогая код. Минус: общий знаменатель всех баз.
- `pgx` — родной драйвер PostgreSQL. Умеет и через `database/sql`, и **напрямую**. Мы используем второй способ: быстрее и даёт доступ к типам PostgreSQL (`JSONB`, массивы, `UUID`), `COPY`, `LISTEN/NOTIFY`.

```text
postgres://app:secret@localhost:5432/app?sslmode=disable
└──┬───┘   └┬┘ └──┬─┘ └───┬───┘ └┬─┘ └┬┘ └──────┬──────┘
   │        │     │       │      │    │         │
 схема   польз. пароль   хост   порт база   параметры
```

| Часть | Частая ошибка |
|---|---|
| `app:secret` | это логин-пароль **базы**; спецсимволы надо кодировать по правилам URL |
| `localhost:5432` | в Docker **не** `localhost`, а имя сервиса — `db` |
| `/app` | один сервер держит много баз — не перепутайте |
| `?sslmode=disable` | годится **только** для локальной разработки |

### 3.39. Соединение и пул

Соединение с базой — полноценная сессия: TCP-рукопожатие, аутентификация, согласование параметров, десятки миллисекунд. Плюс на стороне PostgreSQL на каждое соединение заводится **отдельный процесс** — поэтому соединений не может быть много.

Пул держит открытыми несколько соединений и выдаёт свободное тому, кто попросил, забирая обратно после запроса.

```go
pool, err := pgxpool.New(ctx, os.Getenv("DATABASE_URL"))
defer pool.Close()
if err := pool.Ping(ctx); err != nil {
	log.Fatalf("cannot reach database: %v", err)
}
```

`pgxpool.New` — **один раз при старте**, не в хендлере. `Ping` сразу, потому что `New` соединений ещё не открывает, и неверный пароль иначе всплывёт только на первом запросе пользователя.

**Отсюда критичность `defer rows.Close()`:** `rows` — открытый курсор, он держит соединение, пока вы не дочитали. Не закрыли — соединение не вернулось в пул. Десяток таких запросов, и сервер висит без единой ошибки в логе.

Ещё две вещи в том же цикле:

```go
tasks := []Task{}           // не var tasks []Task
for rows.Next() {
	var t Task
	if err := rows.Scan(&t.ID, &t.Title, &t.Done, &t.CreatedAt); err != nil {
		return nil, err
	}
	tasks = append(tasks, t)
}
return tasks, rows.Err()    // ошибка могла случиться в середине
```

- **`rows.Err()` обязателен.** Если связь оборвалась на середине выборки, `Next()` просто вернёт `false` — цикл закончится как будто нормально, и вы отдадите клиенту половину данных как полный ответ.
- **`tasks := []Task{}`, а не `var tasks []Task`.** Второй при пустом результате даёт `nil`-срез, который превращается в JSON как `null`, а не `[]`, и `.map()` на клиенте падает.

### 3.40. Почему `$1` физически нельзя сломать кавычкой

По сети уходит не одна строка, а последовательность сообщений расширенного протокола:

```text
Parse    "SELECT id, title FROM tasks WHERE id = $1"   ← ТОЛЬКО текст запроса
         база разбирает синтаксис и строит план
Bind     $1 = 42                                        ← ТОЛЬКО значение
Execute                                                 ← выполнить
```

К моменту, когда база получает значение, разбор SQL уже **закончен**, план построен. Что бы ни лежало в `$1`, это данные, и структуру запроса они изменить не могут. Это не экранирование, это **разделение по времени**.

А при склейке строк данные пользователя **становятся текстом запроса**:

```sql
SELECT * FROM tasks WHERE title = '' OR 1=1 --'
```

Апостроф закрыл строку, `OR 1=1` стал частью условия, `--` закомментировал хвост.

Важное ограничение: плейсхолдером можно подставить **только значение**, не имя таблицы или колонки — план строится до подстановки. Нужна динамическая сортировка? Сверяйте имя колонки со списком допустимых в коде.

### 3.41. `Scan`: указатели и порядок

Из базы прилетают байты плюс код типа каждой колонки. `Scan` берёт **адреса** ваших переменных и раскладывает по ним разобранные значения: `text` → `string`, `timestamptz` → `time.Time`.

Сопоставление идёт **по порядку, а не по имени**. Поменяли местами колонки в `SELECT` — и, если типы совпали, значения молча разъедутся по чужим полям. Компилятор здесь бессилен: для него это просто набор указателей.

`QueryRow` не возвращает ошибку сам — она приходит из `Scan`. Поэтому `errors.Is(err, pgx.ErrNoRows)` идёт после `Scan`, и «строка не найдена» — штатный случай, а не сбой.

Для `Exec` проверяйте `RowsAffected()`: `DELETE` несуществующей строки — **не ошибка** для базы, `err` будет `nil`. Но для вашего API это `404`, а не `204`.

### 3.42. `NULL` не помещается в обычный тип Go

```text
Scan в int64:  can't scan into dest[0] (col: user_id): cannot scan NULL into *int64
Scan в *int64: err=<nil>, значение=<nil>
COALESCE:      err=<nil>, значение=0
```

Варианты: сканировать в указатель (`var userID *int64`), подставить значение в SQL (`SELECT COALESCE(user_id, 0)`) или — лучше всего — объявить колонку `NOT NULL` в схеме.

Заодно понятно, чем опасен `LEFT JOIN`: он подставляет `NULL` в колонки, объявленные `NOT NULL`, и `Scan` падает на запросе, который «раньше работал».

### 3.43. Транзакция в Go

```go
tx, err := pool.Begin(ctx)
if err != nil {
	return err
}
defer tx.Rollback(ctx)   // ← страховка на все ветки выхода

_, err = tx.Exec(ctx, `INSERT INTO tasks (title, user_id) VALUES ($1, $2)`, "первая", 1)
_, err = tx.Exec(ctx, `INSERT INTO tasks (title, user_id) VALUES ($1, $2)`, "вторая", 999)
if err != nil {
	return err           // выходим — сработает defer
}

return tx.Commit(ctx)
```

```text
до начала: 0
после первой вставки, ошибка: <nil>
вторая вставка упала — выходим, Rollback сработает по defer

$ psql -c "SELECT count(*) ..."   →   0        ← первая вставка тоже отменена
```

Приём `defer tx.Rollback(ctx)` сразу после `Begin`: из функции можно выйти десятком разных `return`, и на любом откат случится сам. А если дошли до `Commit`, последующий `Rollback` уже ничего не делает и ошибкой не считается.

**Внутри транзакции запросы идут через `tx`, а не через `pool`** — все они выполняются на одном соединении, иначе база не сочла бы их одной транзакцией. Это то, что чаще всего забывают, и тогда транзакция молча не работает.

---

### 3.44. Ошибки: `%w` и `errors.Is`

В Go нет исключений: ошибка — обычное значение обычного интерфейса `error`.

```go
var ErrNotFound = errors.New("not found")

func serviceGet(id int64) error {
	if err := storeGet(id); err != nil {
		return fmt.Errorf("get task %d: %w", id, err)   // заворачиваем
	}
	return nil
}
```

```text
текст ошибки:                get task 42: not found
err == ErrNotFound:          false      ← это уже ДРУГОЕ значение
errors.Is(err, ErrNotFound): true       ← Is разворачивает цепочку
errors.Unwrap(err):          not found
```

`errors.Is` сравнивает с целью, и если не совпало — вызывает `Unwrap` и сравнивает снова, пока цепочка не кончится. С `%v` вместо `%w` исходная ошибка превращается в текст, и `Is` её не найдёт.

**Store не возвращает `pgx.ErrNoRows` наружу** — это деталь реализации. Свой `ErrNotFound` — граница между слоями. И не отдавайте наружу текст ошибки базы: в нём имена таблиц и колонок. В лог — всё, в ответ — «database error».

### 3.45. `defer`: когда он срабатывает

```go
func main() {
	x := 1
	defer fmt.Println("defer 1, x был равен", x)
	x = 2
	defer fmt.Println("defer 2, x был равен", x)
	x = 3
	fmt.Println("конец функции, x =", x)
}
```

```text
конец функции, x = 3
defer 2, x был равен 2
defer 1, x был равен 1
```

- Выполняется при выходе **из функции** — из любого `return` и даже при панике. Не в конце блока и не в конце итерации цикла.
- Порядок обратный: последний отложенный срабатывает первым.
- **Аргументы вычисляются сразу**, в момент записи `defer`.

`defer` внутри цикла — ловушка: все вызовы накопятся и выполнятся только при выходе из функции.

### 3.46. `context`: провод отмены

Когда браузер закрывает вкладку, `net/http` отменяет `r.Context()`, и ваш запрос к базе может прекратиться, не доделав работу впустую.

```go
func slowQuery(ctx context.Context) (string, error) {
	select {
	case <-time.After(2 * time.Second):
		return "результат из базы", nil
	case <-ctx.Done():
		return "", ctx.Err()      // клиент ушёл или истёк таймаут
	}
}

ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
defer cancel()
res, err := slowQuery(ctx)
```

```text
результат: "", ошибка: context deadline exceeded
```

Поэтому в store первым аргументом идёт `ctx`, а в хендлере передаём `r.Context()`. `context.Background()` — контекст без отмены, его место в `main`.

### 3.47. Docker: образ, контейнер, том

Нам нужен PostgreSQL. Ставить его на двадцать пять разных ноутбуков — это двадцать пять разных версий и конфликтов портов.

- **Образ** — готовый слепок файловой системы с установленной программой: `postgres:16` это Linux плюс PostgreSQL 16. У всех абсолютно одинаковый.
- **Контейнер** — запущенный экземпляр образа, изолированный процесс со своей файловой системой и сетью.
- **Том (volume)** — папка, которая живёт **дольше контейнера**. Данные базы кладут в том, иначе при пересоздании контейнера они исчезнут.

Это **не виртуальная машина**: контейнер использует ядро вашей ОС и стартует за секунду.

**Проброс портов `"5432:5432"`** читается как «порт **моей машины** : порт **внутри контейнера**». Если 5432 у вас занят — пишите `"5433:5432"` и меняете порт в `DATABASE_URL`.

**Имя сервиса — это сетевое имя.** Внутри контейнера `localhost` это сам контейнер, а не ваша машина и не соседний контейнер. Поэтому в `docker-compose` приложение обращается к базе по хосту `db`.

**`schema.sql` применяется только при создании базы.** Скрипты из `/docker-entrypoint-initdb.d/` выполняются один раз, при первом старте. Поменяли схему и ничего не изменилось — потому что том уже есть. Пересоздать с нуля: `docker compose down -v` (ключ `-v` удаляет том **вместе с данными**).

### 3.48. CORS — ограничение браузера, а не сервера

```text
Access to fetch at 'http://localhost:8080/tasks' from origin
'http://localhost:3000' has been blocked by CORS policy
```

Сервер ответил нормально: запрос дошёл, код выполнился, ответ вернулся. Это **браузер** отказался отдать ответ вашему JS, потому что страница открыта с одного адреса, а запрос ушёл на другой. Защита от того, чтобы чужой сайт читал ваши данные вашими же куками. В `curl` тот же запрос работает — там нет браузера.

Сервер не «включает» CORS, он лишь **сообщает браузеру разрешение** заголовками. Перед «непростым» запросом (свой заголовок, метод `DELETE`, тип `application/json`) браузер сам шлёт предварительный `OPTIONS`. Не ответите разрешением — основного запроса просто не будет.

---

## Часть 4. Проект на три пары

### Правила

- Команды по **3–4 человека**.
- Один человек в команде — **сборщик**: у него лежит проект целиком, он собирает куски от остальных. Остальные пишут свои файлы и передают ему (гит, мессенджер, флешка — как удобно).
- Сдаёт команда, оценка одна на всех. Поэтому договаривайтесь на берегу, кто что делает.

### Роли

| Роль | Отвечает за | Файлы |
|---|---|---|
| Архитектор данных | схема таблиц, связи, тестовые данные | `schema.sql`, `seed.sql` |
| Бэкендер А | SQL-запросы, работа с базой | `store.go`, `model.go` |
| Бэкендер Б | хендлеры, коды ответов, валидация | `handlers.go`, `main.go` |
| Интегратор | контракт API, проверка через curl, подключение вёрстки | `README.md`, фронт |

В команде из трёх человек роли интегратора и архитектора совмещаются.

### Темы на выбор

Берите одну. Можно предложить свою — согласуйте со мной на паре.

| # | Тема | Минимум сущностей |
|---|---|---|
| 1 | Заявки в IT-отдел вуза | `users`, `tickets`, `comments` |
| 2 | Расписание занятий | `groups`, `teachers`, `subjects`, `lessons` |
| 3 | Трекер сдачи лабораторных | `students`, `labs`, `submissions` |
| 4 | Предзаказ в столовой | `users`, `dishes`, `orders`, `order_items` |
| 5 | Барахолка кампуса | `users`, `listings`, `categories`, `messages` |
| 6 | Прокат оборудования / библиотека | `users`, `items`, `loans` |

**Требования к теме:** минимум 3 таблицы, минимум одна связь «один ко многим», минимум одна сущность, у которой есть состояние (статус заявки, выполненность, выдано/возвращено).

### Что сдать сегодня, в конце этой пары

Кода не пишем. Сдаём четыре вещи.

**1. Схема данных.** Таблицы, колонки, типы, связи. На бумаге, в dbdiagram.io или сразу как `schema.sql`. Для каждой колонки решите: может ли она быть `NULL`, есть ли `DEFAULT`, нужен ли `UNIQUE`.

**2. Контракт API** в `README.md`. Таблица по шаблону:

````markdown
## API

| Метод | Путь | Тело запроса | Успех | Ошибки |
|---|---|---|---|---|
| GET | /tickets | — | 200, массив | 500 |
| GET | /tickets/{id} | — | 200, объект | 400, 404 |
| POST | /tickets | {"title":"...","user_id":1} | 201, объект | 400, 422 |
| PATCH | /tickets/{id} | {"status":"done"} | 200, объект | 400, 404, 422 |
| DELETE | /tickets/{id} | — | 204 | 400, 404 |

### Объект Ticket

```json
{"id":1,"title":"Не работает проектор","status":"new","user_id":3,"created_at":"2026-09-21T09:38:37+05:00"}
```
````

**3. Распределение ролей** — кто что делает, списком в README.

**4. Скелет проекта.** `go.mod`, `main.go` с одним эндпоинтом `/health`, `schema.sql`, `docker-compose.yml`, `.gitignore`. Запускается и отвечает `200`. Писать бизнес-логику сегодня не надо.

> Полчаса проектирования экономят пару часов переписывания. Придумывать схему таблиц, когда уже написаны хендлеры, — самый дорогой способ.

### Пара 2 — реализация

Рабочий CRUD по вашему контракту:

- все эндпоинты из README отвечают и возвращают обещанные коды;
- весь SQL — в `store.go`, весь HTTP — в `handlers.go`;
- параметры только через `$1`, никакой склейки строк;
- валидация входных данных до похода в базу;
- конфигурация через переменные окружения, `.env` не в гите;
- `seed.sql` с тестовыми данными, чтобы проект можно было посмотреть.

### Пара 3 — фронт и защита

- Ваша вёрстка ходит в ваш API через `fetch()`: список грузится с сервера, создание работает, удаление работает.
- Ошибки сервера показываются пользователю, а не молчат.
- Защита: 5 минут на команду. Показываете работающее приложение и отвечаете на вопросы по своему коду — **любому участнику команды**. Поэтому смотрите не только свою часть.

### Критерии оценки (100 баллов)

| Что | Баллов |
|---|---|
| Схема БД: типы, `NOT NULL`, внешние ключи, индекс на FK | 15 |
| README с контрактом API, соответствующим реальности | 10 |
| CRUD работает целиком | 20 |
| Корректные коды ответов (201/204/400/404/422/500) | 10 |
| Разделение слоёв: SQL не течёт в хендлеры | 15 |
| Безопасность: `$1`, нет секретов в гите, ошибки базы не наружу | 10 |
| Фронт подключён и работает | 10 |
| Защита: все участники понимают проект целиком | 10 |

Штрафы: `.env` с паролем в репозитории — минус 10. Склейка SQL строками — минус 10. Проект не запускается по вашей же инструкции — минус 15.

---

## Часть 5. Чеклист перед сдачей

Пройдите по нему **до** того, как звать преподавателя. В скобках — где разобрано, почему именно так.

**Go и HTTP**

- [ ] все поля моделей с большой буквы и с тегами `json:"..."` *(3.28)*
- [ ] `Content-Type` выставлен явно, а не отдан на угадывание *(3.26)*
- [ ] после каждого ответа об ошибке стоит `return` *(3.27)*
- [ ] `w.Header().Set(...)` идёт до `w.WriteHeader(...)` *(3.26)*
- [ ] нет маршрута `"/"` там, где нужен только корень — для этого `"/{$}"` *(3.21)*
- [ ] вместо `http.HandleFunc` используется свой `mux` *(3.21)*
- [ ] на каждый `POST`/`PATCH` — своя структура запроса, а не модель *(3.31)*
- [ ] проверяется `err` у `strconv.ParseInt`, а не только значение *(3.27)*
- [ ] в `Scan` и `Decode` перед переменной стоит `&` *(3.41)*
- [ ] нет изменяемых глобальных переменных, `go run -race` молчит *(3.12)*

**База**

- [ ] после каждого `Query` стоит `defer rows.Close()` *(3.39)*
- [ ] в конце цикла проверяется `rows.Err()` *(3.39)*
- [ ] списки инициализированы как `[]Task{}`, а не `var tasks []Task` *(3.39)*
- [ ] порядок колонок в `SELECT` совпадает с порядком в `Scan` *(3.41)*
- [ ] нигде нет конкатенации строк в SQL — только `$1`, `$2` *(3.40)*
- [ ] у каждого `UPDATE` и `DELETE` есть `WHERE`
- [ ] `RowsAffected() == 0` превращается в `404` *(3.41)*
- [ ] на каждом внешнем ключе есть индекс *(3.33)*
- [ ] нет запросов в цикле там, где нужен `JOIN` — проблема N+1 *(3.36)*
- [ ] `pgxpool.New` вызывается один раз при старте, а не в хендлере *(3.39)*

**Проект**

- [ ] обрабатывается `OPTIONS`, иначе фронт получит ошибку CORS *(3.48)*
- [ ] в `docker-compose` приложение обращается к базе по имени сервиса, а не по `localhost` *(3.47)*
- [ ] `.env` в `.gitignore`, в репозитории только `.env.example`
- [ ] `schema.sql` лежит в репозитории и применяется с нуля
- [ ] проект поднимается по инструкции из README на чужой машине
- [ ] тексты ошибок базы не уходят клиенту, а уходят в лог *(3.44)*

---

## Часть 6. Домашнее задание

К следующей паре установите всё, что нужно. **На паре ставить не будем** — придёте без этого, будете смотреть, как работают другие.

### 1. Go

Скачайте с **https://go.dev/dl** последнюю версию (1.27.x). Проверка:

```bash
go version
# go version go1.27.1 windows/amd64
```

### 2. Docker Desktop

**https://www.docker.com/products/docker-desktop**. Проверка:

```bash
docker --version
docker compose version
```

Docker нужен, чтобы получить PostgreSQL одной командой и не ставить его руками.

### 3. Редактор

VS Code + расширение **Go** (от Go Team at Google). При первом открытии `.go`-файла он предложит доустановить инструменты — соглашайтесь. Альтернатива: GoLand (бесплатен по студенческой лицензии).

### 4. Проверьте, что всё вместе работает

Соберите пример из части 2 целиком и убедитесь, что получаете:

```bash
$ docker compose up -d
$ go run .
2026/09/21 09:38:32 listening on :8080

$ curl localhost:8080/health
{"status":"ok"}
```

Если получили `{"status":"ok"}` — вы готовы к следующей паре.

### Если что-то не ставится

Напишите мне **до** пары, с текстом ошибки целиком и скриншотом. За час до пары — поздно.

---

## Ссылки

- **https://go.dev/play** — Go в браузере
- **https://pgplayground.com**, **https://extendsclass.com/postgresql-online.html** — PostgreSQL в браузере
- **https://dbdiagram.io** — нарисовать схему БД
- **https://pkg.go.dev/net/http** — документация `net/http`
- **https://github.com/jackc/pgx** — драйвер PostgreSQL для Go
- **https://postgrespro.ru/docs/postgresql** — документация PostgreSQL на русском
- **https://go.dev/doc/effective_go** — как писать на Go идиоматично
