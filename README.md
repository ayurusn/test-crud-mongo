# mongo-crud

Реализован простой веб-сервер для взаимодействия с MongoDB. 

## Параметры
Обработка маршрутов: ```Gorilla Mux```.

Хранения данных: ```MongoDB```.

Драйвер для MongoDB: ```https://github.com/mongodb/mongo-go-driver```.

Параметры для запуска веб-сервера хрянятся в файле конфигурации ```config.json```

Обрабатываются обьекты следующего вида:
``` Golang
type Object struct {
	ID          string  `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}
```

## Маршруты
Веб-сервер предоставляет следующие маршруты:

1. ```GET /objects``` - возвращает список всех объектов в базе данных
2. ```GET /objects/{id}``` - возвращает объект с заданным идентификатором
3. ```POST /objects``` - создает новый объект
4. ```PUT /objects/{id}``` - обновляет объект с заданным идентификатором
5. ```DELETE /objects/{id}``` - удаляет объект с заданным идентификатором


## Запуск:

 ``` bash
 docker run --name some-mongo -p 27017:27017 --rm mongo:latest
 ```

 ``` bash
 go build
 ./mongo-crud 
 ```

## Взаимодействие:

Работу с сервером можно производить с помощью curl-скриптов, которые находятся в файле ```curl_scripts.sh```

``` bash
# Получить список всех объектов
curl http://localhost:8000/objects/

# Получить объект с определенным идентификатором
curl http://localhost:8000/objects/{id}

# Создать новый объект
curl -X POST -H "Content-Type: application/json" -d '{"name": "Example 2", "description": "Another example object"}' http://localhost:8000/objects

# Обновить объект с определенным идентификатором
curl -X PUT -H "Content-Type: application/json" -d '{"name": "Updated Example 2", "description": "Updated example object"}' http://localhost:8000/objects/{id}

# Удалить объект с определенным идентификатором
curl -X DELETE http://localhost:8000/objects/{id}
```

## Тесты
 Написаны unit-тесты для всех handler-функций.

Запуск тестов: ```go test ./...```

 ##### ВАЖНО: 

 Для работы тестов неодходимо запустить контейнер с MongoDB.

 ```bash
  docker run --name some-mongo -p 27017:27017 --rm mongo:latest
 ```