package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDB configuration
type Config struct {
	Mongo struct {
		Host       string `json:"host"`
		Port       int    `json:"port"`
		Database   string `json:"database"`
		Collection string `json:"collection"`
	} `json:"mongo"`
}

type Object struct {
	ID          string  `json:"id,omitempty"`
	Name        *string `json:"name,omitempty"`
	Description *string `json:"description,omitempty"`
}

var Collection *mongo.Collection

func main() {
	// Read config from file
	configFile, err := os.Open("config.json")
	if err != nil {
		log.Fatalf("unable to open configuration file: %s", err)
	}
	defer configFile.Close()

	configData, err := ioutil.ReadAll(configFile)
	if err != nil {
		log.Fatalf("unable to read configuration file: %s", err)
	}

	var config Config
	err = json.Unmarshal(configData, &config)
	if err != nil {
		log.Fatalf("unable to parse configuration: %s", err)
	}

	// Connect to MongoDB
	mongoURI := fmt.Sprintf("mongodb://%s:%d", config.Mongo.Host, config.Mongo.Port)
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("unable to connect to MongoDB: %s", err)
	}

	// Ping the MongoDB server to verify the connection
	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}

	// Get the collection reference
	Collection = client.Database(config.Mongo.Database).Collection(config.Mongo.Collection)

	// Create a new router using Gorilla Mux
	router := mux.NewRouter()

	// API endpoints
	router.HandleFunc("/objects", createObject).Methods("POST")
	router.HandleFunc("/objects/", getAllObjects).Methods("GET")
	router.HandleFunc("/objects/{id}", getObject).Methods("GET")
	router.HandleFunc("/objects/{id}", updateObject).Methods("PUT")
	router.HandleFunc("/objects/{id}", deleteObject).Methods("DELETE")

	// Start the HTTP server
	log.Println("Server listening on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}

func writeError(w http.ResponseWriter, status int, format string, a ...any) {
	s := fmt.Sprintf(format, a...)
	w.WriteHeader(status)
	w.Header().Set("Content-Type", "application/json")
	body, _ := json.Marshal(map[string]string{"error": s})
	w.Write(body)
}

func getObjectByID(id string) (*Object, error) {
	var obj Object
	filter := bson.M{"id": id}
	err := Collection.FindOne(context.Background(), filter).Decode(&obj)
	if err != nil {
		return nil, err
	}
	return &obj, nil
}

// Handler for creating a object
func createObject(w http.ResponseWriter, r *http.Request) {
	var object Object
	err := json.NewDecoder(r.Body).Decode(&object)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "object creation failed: %s", err)
		return
	}

	// Generate a new unique ID for the object
	object.ID = primitive.NewObjectID().Hex()

	// Insert the object into the collection
	_, err = Collection.InsertOne(context.Background(), object)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "object creation failed: %s", err)
		return
	}

	// Return the created object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(object)
}

// Handler for getting a object by ID
func getObject(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	objectID := params["id"]

	object, err := getObjectByID(objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "object not found")
		} else {
			writeError(w, http.StatusInternalServerError, "object extracrion failed")
		}
		return
	}

	// Return the object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(object)
}

// Handler for updating a object
func updateObject(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	objectID := params["id"]

	// Create an update filter to match the object ID
	filter := bson.M{"id": objectID}

	currentObj, err := getObjectByID(objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "object for update not found")
		} else {
			writeError(w, http.StatusInternalServerError, "update failed")
		}
		return
	}

	// object from request, new data for update
	var updateObj Object
	err = json.NewDecoder(r.Body).Decode(&updateObj)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "object updating failed: %s", err)
		return
	}

	// check new object fields for nil
	// if nil then get current value of the field
	if updateObj.Name == nil {
		updateObj.Name = currentObj.Name
	}
	if updateObj.Description == nil {
		updateObj.Description = currentObj.Description
	}

	// Create an update document with the updated fields
	update := bson.M{"$set": bson.M{
		"name":        updateObj.Name,
		"description": updateObj.Description,
	}}

	// Perform the update operation
	_, err = Collection.UpdateOne(context.Background(), filter, update)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "object updating failed: %s", err)
		return
	}

	// Return the updated object
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(updateObj)
}

// Handler for deleting a object
func deleteObject(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	objectID := params["id"]

	// Check existing of object before deleting
	_, err := getObjectByID(objectID)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			writeError(w, http.StatusNotFound, "object for deleting not found")
		} else {
			writeError(w, http.StatusInternalServerError, "deletion failed")
		}
		return
	}
	// Delete the object from the collection
	filter := bson.M{"id": objectID}
	_, err = Collection.DeleteOne(context.Background(), filter)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "deletion failed: %s", err)
		return
	}

	// Return a success message
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("object deleted successfully"))
}

// Handler for getting a list of all objects
func getAllObjects(w http.ResponseWriter, r *http.Request) {
	// Find all objects in the collection
	cursor, err := Collection.Find(context.Background(), bson.M{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "getting objects failed: %s", err)
		return
	}
	defer cursor.Close(context.Background())

	objects := make([]Object, 0)
	for cursor.Next(context.Background()) {
		var object Object
		err := cursor.Decode(&object)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "getting objects failed: %s", err)
			return
		}
		objects = append(objects, object)
	}

	if err := cursor.Err(); err != nil {
		writeError(w, http.StatusInternalServerError, "getting objects failed: %s", err)
		return
	}

	// Return the list of objects
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(objects)
}
