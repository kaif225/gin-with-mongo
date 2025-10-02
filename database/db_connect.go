package database

import (
	"fmt"
	"log"
	"os"

	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

var Client *mongo.Client

func Connect() error {

	mongoURI := os.Getenv("MONGO_URI")
	connectionString := options.Client().ApplyURI(mongoURI)
	client, err := mongo.Connect(connectionString)
	if err != nil {
		log.Println(err)
		return err
	}
	Client = client

	fmt.Println("Connected to mongoDB")
	return nil
}

//func checkCollection()
