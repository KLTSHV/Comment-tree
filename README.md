# CommentTree

Сервис для работы с древовидными комментариями (неограниченная вложенность) с HTTP API, постраничным выводом, сортировкой, полнотекстовым поиском и web-интерфейсом (HTML + JS).

## Возможности

- **POST /comments** — создание комментария (в т.ч. ответа с `parent_id`)
- **GET /comments?parent={id}** — получение дерева комментариев
  - корни: постранично + каждое поддерево целиком
  - конкретный parent: сам parent + его прямые дети постранично (каждый ребёнок раскрывается вниз полностью)
- **DELETE /comments/{id}** — удаление комментария и всего поддерева
- **GET /comments/search** — поиск по `text` и `author` (постранично + сортировка)
- **Web UI** — просмотр дерева, создание/ответы, удаление, поиск

> Хранение: **in-memory** (данные исчезают при перезапуске).

---

## Установка и запуск

Из корня проекта:

```bash
go mod tidy
go run ./cmd/api
```

## Сервис поднимется на:
	•	Web UI: http://localhost:8080/
	•	API: http://localhost:8080/comments

⸻

## Параметры сортировки и пагинации

Для GET /comments и GET /comments/search:
	•	page — номер страницы (по умолчанию 1)
	•	limit — размер страницы (по умолчанию 20 для API, в UI обычно 10)
	•	sort — сортировка по времени создания:
	•	created_at_asc
	•	created_at_desc

⸻

## API

1) Создать комментарий

POST /comments

Тело запроса (JSON):

{
  "parent_id": 0,
  "author": "Senpai",
  "text": "Привет, мир"
}

	•	parent_id = 0 — корневой комментарий
	•	parent_id > 0 — ответ на существующий комментарий

Пример:

curl -i -X POST http://localhost:8080/comments \
  -H "Content-Type: application/json" \
  -d '{"parent_id":0,"author":"Senpai","text":"Привет, мир"}'

Ответ: 201 Created, JSON объекта комментария.

⸻

2) Получить дерево комментариев

2.1 Корневые комментарии (постранично) + полные поддеревья
GET /comments?page=1&limit=10&sort=created_at_asc

Пример:

curl "http://localhost:8080/comments?page=1&limit=10&sort=created_at_asc"

Ответ:
	•	items — массив корневых CommentNode
	•	у каждого узла рекурсивно есть children
	•	meta.total — всего корневых комментариев

⸻

2.2 Получить конкретную ветку (parent) + прямых детей постранично
GET /comments?parent=1&page=1&limit=5&sort=created_at_desc

Пример:

curl "http://localhost:8080/comments?parent=1&page=1&limit=5&sort=created_at_desc"

Ответ:
	•	parent — узел комментария #1
	•	parent.children — только прямые дети (постранично), но каждый ребёнок раскрывается вниз целиком

⸻

3) Удалить комментарий и поддерево

DELETE /comments/{id}

Пример:

curl -i -X DELETE "http://localhost:8080/comments/1"

Ответ:

{
  "id": 1,
  "deleted": 5
}

deleted — сколько узлов удалено (комментарий + все вложенные).

⸻

4) Поиск по комментариям

GET /comments/search?q=...&page=1&limit=20&sort=created_at_desc

Ищет в text и author (case-insensitive).

Пример:

curl "http://localhost:8080/comments/search?q=ответ&page=1&limit=20&sort=created_at_desc"

Ответ:
	•	items — плоский список Comment
	•	meta.total — всего совпадений

⸻

## Web UI

Откройте: http://localhost:8080/

Доступно:
	•	просмотр дерева с визуальной вложенностью
	•	создание корневых комментариев и ответов (кнопка «Ответить» выставляет parent_id)
	•	удаление комментария и его поддерева
	•	поиск по ключевым словам

