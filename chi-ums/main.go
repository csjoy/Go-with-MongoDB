package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoDB = "users"
const mongoURL = "mongodb://localhost:27017/" + mongoDB

var ctx = context.Background()
var collection *mongo.Collection

type status map[string]interface{}

type User struct {
	ID     string `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string `json:"name" bson:"name"`
	Gender string `json:"gender" bson:"gender"`
	Age    int    `json:"age" bson:"age"`
}

func main() {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURL))
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer client.Disconnect(ctx)

	mdb := client.Database(mongoDB)
	collection = mdb.Collection("users")

	r := chi.NewRouter()
	r.Use(middleware.Logger)
	r.Post("/user", createUser)
	r.Get("/user", readUser)
	r.Put("/user/{id}", updateUser)
	r.Delete("/user/{id}", deleteUser)
	log.Fatal(http.ListenAndServe(":8080", r))
}

func createUser(w http.ResponseWriter, r *http.Request) {
	var user User
	err := json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	user.ID = ""
	res, err := collection.InsertOne(ctx, user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "_id", Value: res.InsertedID}}

	var createdUser User
	err = collection.FindOne(ctx, filter).Decode(&createdUser)
	if err == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(createdUser)
}

func readUser(w http.ResponseWriter, r *http.Request) {
	query := bson.D{{}}
	cur, err := collection.Find(ctx, query)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	var users []User
	err = cur.All(ctx, &users)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(users)
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	var user User
	err = json.NewDecoder(r.Body).Decode(&user)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "_id", Value: userID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "name", Value: user.Name},
				{Key: "gender", Value: user.Gender},
				{Key: "age", Value: user.Age},
			},
		},
	}
	var updatedUser User
	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedUser)
	if err == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	updatedUser.ID = id
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(updatedUser)
}

func deleteUser(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")
	userID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "_id", Value: userID}}

	var deletedUser User
	err = collection.FindOneAndDelete(ctx, filter).Decode(&deletedUser)
	if err == mongo.ErrNoDocuments {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	} else if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(status{"error": err.Error()})
		return
	}

	deletedUser.ID = id
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(deletedUser)
}
