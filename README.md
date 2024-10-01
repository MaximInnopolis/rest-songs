# rest-songs

# Описание

Этот проект реализует REST API для управления музыкальной библиотекой,
позволяя пользователям выполнять различные операции с песнями:

* Получение данных библиотеки с фильтрацией по всем полям и пагинацией
* Получение текста песни с пагинацией по куплетам
* Удаление песни
* Изменение данных песни
* Добавление новой песни


Используется Postgresql в качестве субд, Docker для контейнеризации,
mockserver в качестве внешнего API. Код покрыт info и debug логами.
Был сгенерирован swagger на реализованный API

## Требования

- Go 1.18+
- PostgreSQL
- Docker

## Установка и запуск

1. Клонируйте репозиторий:
```bash
git clone https://github.com/MaximInnopolis/rest-songs.git
cd rest-songs
```

2. Соберите докер-билд:
```bash
make up-all
```

3. Проведите миграцию:
```bash
make migrate
```

4. Для запуска mockserver:
```bash
curl -X PUT "http://localhost:1080/mockserver/expectation" -d @docs/expectation.json -H "Content-Type: application/json"
```

API доступен по адресу <http:localhost:8080>
Mockserver доступен по адресу <http:localhost:1080>
Postman коллекция доступна по следующему пути: [Postman коллекция](docs/songs.postman_collection.json)

## Примечания

Использовать дату в формате 02.01.2006
По дефолту page = 1, pageSize = 10 (10 песен на странице)