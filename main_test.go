package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/require"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// Create test MongoDB
func setupTestDatabase(t *testing.T) *mongo.Database {
	assert := require.New(t)

	config := Config{
		Mongo: struct {
			Host       string `json:"host"`
			Port       int    `json:"port"`
			Database   string `json:"database"`
			Collection string `json:"collection"`
		}{
			Host:       "localhost",
			Port:       27017,
			Database:   "test_db",
			Collection: "test_collection",
		},
	}

	// Connect to MongoDB
	mongoURI := fmt.Sprintf("mongodb://%s:%d", config.Mongo.Host, config.Mongo.Port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	assert.NoError(err)

	// Create test db
	db := client.Database(config.Mongo.Database)

	// Ping the MongoDB server to verify connection
	err = client.Ping(context.Background(), nil)
	assert.NoError(err)

	return db
}

func getStringPointer(s string) *string {
	return &s
}
func TestGetAllObjects(t *testing.T) {
	assert := require.New(t)

	db := setupTestDatabase(t)
	defer db.Drop(context.Background())
	Collection = db.Collection("test_collection")

	// Get all objects from an empty database
	req, err := http.NewRequest("GET", "/objects", nil)
	assert.NoError(err)

	recorder := httptest.NewRecorder()

	// Run handler function to test
	getAllObjects(recorder, req)

	assert.Equal(http.StatusOK, recorder.Code)
	assert.Equal("[]\n", recorder.Body.String())

	// Get all objects from the filled database
	objects := []Object{
		{ID: "1", Name: getStringPointer("Object 1"), Description: getStringPointer("Description 1")},
		{ID: "2", Name: getStringPointer("Object 2"), Description: getStringPointer("Description 2")},
		{ID: "3", Name: getStringPointer("Object 3"), Description: getStringPointer("Description 3")},
	}
	for _, obj := range objects {
		_, err := Collection.InsertOne(context.Background(), obj)
		assert.NoError(err)
	}

	req, err = http.NewRequest("GET", "/objects", nil)
	assert.NoError(err)

	recorder = httptest.NewRecorder()

	// Run handler function to test
	getAllObjects(recorder, req)

	assert.Equal(http.StatusOK, recorder.Code)

	var response []Object
	err = json.Unmarshal(recorder.Body.Bytes(), &response)
	assert.NoError(err)
	assert.Equal(objects, response)
}

func TestCreateObject(t *testing.T) {
	assert := require.New(t)

	// Create test mongo
	db := setupTestDatabase(t)
	defer db.Drop(context.Background())
	Collection = db.Collection("test_collection")

	// Object to create -> JSON
	object := Object{
		Name:        getStringPointer("Test Object"),
		Description: getStringPointer("This is a test object"),
	}
	jsonObject, err := json.Marshal(object)
	assert.NoError(err)

	req, err := http.NewRequest("POST", "/objects", bytes.NewBuffer(jsonObject))
	assert.NoError(err)

	recorder := httptest.NewRecorder()

	// Run handler function to test
	createObject(recorder, req)

	assert.Equal(http.StatusOK, recorder.Code)

	var createdObject Object
	err = json.Unmarshal(recorder.Body.Bytes(), &createdObject)
	assert.NoError(err)

	assert.NotEmpty(createdObject.ID)

	dbObject, err := getObjectByID(createdObject.ID)
	assert.NoError(err)
	assert.Equal(createdObject, *dbObject)
}

func TestGetObject(t *testing.T) {
	assert := require.New(t)
	// Create test mongo
	db := setupTestDatabase(t)
	defer db.Drop(context.Background()) // Удаляем тестовую базу данных после завершения теста
	Collection = db.Collection("test_collection")

	object := Object{
		Name:        getStringPointer("Test Object"),
		Description: getStringPointer("This is a test object"),
	}

	// Get a non-existent object
	req, err := http.NewRequest("GET", "/objects/"+object.ID, nil)
	assert.NoError(err)

	recorder := httptest.NewRecorder()

	// Run handler function to test
	getObject(recorder, req)

	assert.Equal(http.StatusNotFound, recorder.Code)
	assert.Equal("{\"error\":\"object not found\"}", recorder.Body.String())

	// Get an existing object
	_, err = Collection.InsertOne(context.Background(), object)
	assert.NoError(err)

	req, err = http.NewRequest("GET", "/objects/"+object.ID, nil)
	assert.NoError(err)

	recorder = httptest.NewRecorder()

	// Run handler function to test
	getObject(recorder, req)

	assert.Equal(http.StatusOK, recorder.Code)

	var retrievedObject Object
	err = json.Unmarshal(recorder.Body.Bytes(), &retrievedObject)
	assert.NoError(err)

	assert.Equal(object, retrievedObject)
}

func TestUpdateObject(t *testing.T) {
	assert := require.New(t)

	db := setupTestDatabase(t)
	defer db.Drop(context.Background())
	Collection = db.Collection("test_collection")

	// Updating an object that does not exist in the database
	object := Object{
		Name:        getStringPointer("Test Object"),
		Description: getStringPointer("This is a test object"),
	}

	jsonObject, err := json.Marshal(object)
	assert.NoError(err)

	req, err := http.NewRequest("PUT", "/objects/"+object.ID, bytes.NewBuffer(jsonObject))
	assert.NoError(err)

	recorder := httptest.NewRecorder()

	// Run handler function to test
	updateObject(recorder, req)

	assert.Equal(http.StatusNotFound, recorder.Code)
	assert.Equal("{\"error\":\"object for update not found\"}", recorder.Body.String())

	// Regular update
	_, err = Collection.InsertOne(context.Background(), object)
	assert.NoError(err)

	// Сhange the original object
	object.Name = getStringPointer("Updated Object")
	object.Description = getStringPointer("This is an updated object")

	jsonObject, err = json.Marshal(object)
	assert.NoError(err)

	req, err = http.NewRequest("PUT", "/objects/"+object.ID, bytes.NewBuffer(jsonObject))
	assert.NoError(err)

	recorder = httptest.NewRecorder()

	// Run handler function to test
	updateObject(recorder, req)

	assert.Equal(http.StatusOK, recorder.Code)

	var updatedObject Object
	err = json.Unmarshal(recorder.Body.Bytes(), &updatedObject)
	assert.NoError(err)

	dbObject, err := getObjectByID(object.ID)
	assert.NoError(err)
	assert.Equal(object, *dbObject)

	// Incomplete update, one of the fields was not passed
	var incompleteObject Object
	incompleteObject.Name = getStringPointer("Incomplete Updated Object")

	incompleteJsonObject, err := json.Marshal(incompleteObject)
	assert.NoError(err)

	req, err = http.NewRequest("PUT", "/objects/"+object.ID, bytes.NewBuffer(incompleteJsonObject))
	assert.NoError(err)

	recorder = httptest.NewRecorder()

	// Run handler function to test
	updateObject(recorder, req)

	assert.Equal(http.StatusOK, recorder.Code)

	err = json.Unmarshal(recorder.Body.Bytes(), &updatedObject)
	assert.NoError(err)

	dbObject, err = getObjectByID(object.ID)
	assert.NoError(err)

	assert.Equal(*incompleteObject.Name, *dbObject.Name)
	assert.Equal(*object.Description, *dbObject.Description)
}

func TestDeleteObject(t *testing.T) {
	assert := require.New(t)

	db := setupTestDatabase(t)
	defer db.Drop(context.Background())
	Collection = db.Collection("test_collection")

	object := Object{
		Name:        getStringPointer("Test Object"),
		Description: getStringPointer("This is a test object"),
	}
	// Deleting a non-existent object
	req, err := http.NewRequest("DELETE", "/objects/"+object.ID, nil)
	assert.NoError(err)

	recorder := httptest.NewRecorder()

	// Run handler function to test
	deleteObject(recorder, req)

	assert.Equal(http.StatusNotFound, recorder.Code)
	assert.Equal("{\"error\":\"object for deleting not found\"}", recorder.Body.String())

	// Deleting an existing object
	_, err = Collection.InsertOne(context.Background(), object)
	assert.NoError(err)

	req, err = http.NewRequest("DELETE", "/objects/"+object.ID, nil)
	assert.NoError(err)

	recorder = httptest.NewRecorder()

	// Run handler function to test
	deleteObject(recorder, req)

	assert.Equal(http.StatusOK, recorder.Code)

	_, err = getObjectByID(object.ID)
	assert.Equal(mongo.ErrNoDocuments, err)
}
