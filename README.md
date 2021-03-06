### **Агрегатор новостей**

----
#### **Описание**
Программа, которая собирает публикации с нескольких информационных сайтов и показывает пользователям агрегированную ленту самых последних новостей.
  
****
#### **Возможности**

| № | Описание |
| :----------------: | :---------------- |
| **1** | Приложение имеет веб-интерфейс с отображением **десяти** последних по времени публикаций.|
| **2** | Приложение принимает на вход конфигурационный файл в формате JSON с массивом ссылок на RSS-ленты информационных сайтов и периодом опроса в минутах.|
| **3** | Приложение регулярно выполняет обход всех переданных в конфигурации RSS-лент.|
| **4** | Приложение выполняет чтение каждой RSS-ленты в отдельном потоке выполнения (горутине).|
| **5** | Приложение сохраняет публикации в БД.|
| **6** | Приложение состоит из сервера приложений, базы данных и веб-интерфейса пользователя.|
| **7** | Веб-интерфейс получает от сервера приложений данные в формате JSON.|
| **8** | Сервер приложения предоставляет API, посредством которого осуществляется взаимодействие сервера и веб-интерфейса.|
| **9** | API предоставляет метод для получения заданного количества новостей. Требуемое количество публикаций указывается в пути запроса метода API.|
| **10** | Агрегатор хранит следующий набор данных для каждой публикации: **Заголовок (title)**, **Описание (description)**, **Дата публикации (pubDate)**, **Ссылка на источник (link)**|

****
#### **Использование**

```bash
git clone https://github.com/rtemka/news.git
cd ./news
```

##### **Для прогона тестов:**

```bash
export POSTGRES_TEST_DB_URL="postgres://[user:password]@localhost:5432/[testdb-name]"
```

```bash
export MONGO_TEST_DB_URL="mongodb://[user:password]@localhost:27017"
```

Если эти переменные окружения не установлены, то тесты **будут пропущены**.

```bash
go test -v ./...
```

##### **Настройка Базы данных**

Для полноценного запуска приложения необходимо иметь на хост-машине установленный **Postgres/Mongodb**, а также **прописать переменные окружения** для подключения к БД.

Для рабочей БД, например:
```bash
export NEWS_DB_CONN_STRING="postgres://[user:password]@host.docker.internal:5432/[db-name]"
```
... или 
```bash
export NEWS_DB_CONN_STRING="mongodb://[user:password]@host.docker.internal:27017"
```

**Схему** БД можно найти **[тут](pkg/storage/postgres/schema.sql)**

##### **Docker**
Собираем образ и запускаем контейнер

```bash
docker build -t news .
```
```bash
docker run --rm -it -p 5432:5432 -p 27017:27017 -p 8080:8080 --env NEWS_DB_CONN_STRING --name news news
```

##### **Из исходника**
```bash
go build -o ./cmd/news/news ./cmd/news/news.go
./cmd/news/news ./cmd/news/config.json
```