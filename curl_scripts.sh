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