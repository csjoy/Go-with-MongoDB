package main

import (
	"context"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

const mongoDB = "hrms"
const mongoURI = "mongodb://localhost:27017/" + mongoDB

var ctx = context.Background()
var collection *mongo.Collection

type Employee struct {
	ID     string  `json:"id,omitempty" bson:"_id,omitempty"`
	Name   string  `json:"name" bson:"name"`
	Salary float64 `json:"salary" bson:"salary"`
	Age    int64   `json:"age" bson:"age"`
}

func main() {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatal("Error connecting to database:", err)
	}
	defer client.Disconnect(ctx)

	mdb := client.Database(mongoDB)
	collection = mdb.Collection("employees")

	r := gin.Default()
	r.POST("/employee", createEmployee)
	r.GET("/employee", readEmployee)
	r.PUT("/employee/:id", updateEmployee)
	r.DELETE("/employee/:id", deleteEmployee)
	log.Fatal(r.Run("localhost:8080"))
}

func createEmployee(c *gin.Context) {
	var employee Employee
	err := c.BindJSON(&employee)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	employee.ID = ""
	res, err := collection.InsertOne(ctx, employee)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	filter := bson.D{{Key: "_id", Value: res.InsertedID}}

	var createdEmployee Employee
	err = collection.FindOne(ctx, filter).Decode(&createdEmployee)
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusCreated, createdEmployee)
}

func readEmployee(c *gin.Context) {
	query := bson.D{{}}
	cur, err := collection.Find(ctx, query)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer cur.Close(ctx)

	var employees []Employee
	err = cur.All(ctx, &employees)
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.IndentedJSON(http.StatusOK, employees)
}

func updateEmployee(c *gin.Context) {
	id := c.Param("id")
	employeeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var employee Employee
	err = c.BindJSON(&employee)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "_id", Value: employeeID}}
	update := bson.D{
		{
			Key: "$set",
			Value: bson.D{
				{Key: "name", Value: employee.Name},
				{Key: "salary", Value: employee.Salary},
				{Key: "age", Value: employee.Age},
			},
		},
	}

	var updatedEmployee Employee
	err = collection.FindOneAndUpdate(ctx, filter, update).Decode(&updatedEmployee)
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	updatedEmployee.ID = id
	c.IndentedJSON(http.StatusOK, updatedEmployee)
}

func deleteEmployee(c *gin.Context) {
	id := c.Param("id")
	employeeID, err := primitive.ObjectIDFromHex(id)
	if err != nil {
		c.IndentedJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	filter := bson.D{{Key: "_id", Value: employeeID}}

	var deletedEmployee Employee
	err = collection.FindOneAndDelete(ctx, filter).Decode(&deletedEmployee)
	if err == mongo.ErrNoDocuments {
		c.IndentedJSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	} else if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	deletedEmployee.ID = id
	c.IndentedJSON(http.StatusOK, deletedEmployee)
}
