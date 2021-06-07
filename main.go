package main

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
	"net/http"
	"os"
)

type MongoDB struct {
	Event *mongo.Collection
	Session *mongo.Client
}

func connect() (MongoDB) {
	client, err := mongo.Connect(context.Background(), options.Client().ApplyURI(string(os.Getenv("MONGODB_URL"))))
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(context.Background(), nil)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Connected to MongoDB!")

	return MongoDB{
		Session: client,
	}
}
func (db MongoDB) CloseDB() {
	err := db.Session.Disconnect(context.Background())
	if err != nil {
		log.Fatal(err)
	}
}

type User struct {
	ID				primitive.ObjectID	`json:"_id,omitempty" bson:"_id,omitempty"`
	FullName		string				`json:"fullName,omitempty" bson:"fullName,omitempty"`
	Email			string				`json:"email,required" bson:"email,required"`
	PhoneNo			string				`json:"phoneNo,omitempty" bson:"phoneNo,omitempty"`
	Password		string				`json:"password,required" bson:"password,required"`
}

func getUser(email string, client *mongo.Client) (*User, error)	{
	var result User
	query := bson.D{{"email", email}}
	collection := client.Database("hackon").Collection("users")
	err := collection.FindOne(context.TODO(), query).Decode(&result)
	if err != nil {
		return nil, err
	}
	return &result, nil
}

func createUser(user *User, client *mongo.Client) (string, error){
	collection := client.Database("hackon").Collection("users")
	insertResult, err := collection.InsertOne(context.TODO(), user)
	if err != nil {
		log.Fatal(err)
		return "",err
	}
	id := fmt.Sprintf("%s", insertResult.InsertedID)
	return id,nil

}
func main()	{
	godotenv.Load(".env")
	dbConnection := connect()
	defer dbConnection.CloseDB()
	client := dbConnection.Session

	http.Handle("/", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request){
		fmt.Fprint(rw, "live")
	}))
	http.Handle("/signup", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request){
		var user User
		err := json.NewDecoder(r.Body).Decode(&user)
		if err != nil {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			payload := struct {
					Error string `json:"error"`
			}{Error: "Invalid or Incomplete fields"}

			json.NewEncoder(rw).Encode(payload)
			return
		}
		_, err = getUser(user.Email, client)
		if err == nil {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)

			payload := struct {
				Error string `json:"error"`
			}{Error: "user with this email already exists"}

			json.NewEncoder(rw).Encode(payload)
			return
		}
		_, err = createUser(&user, client)
		if err != nil {
			fmt.Println(err)
			rw.WriteHeader(http.StatusInternalServerError)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusCreated)
		json.NewEncoder(rw).Encode(user)
	}))
	http.Handle("/login", http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request){
		var data struct {
			Email 		string 	`json:"email"`
			Password	string	`json:"password"`
		}
		err := json.NewDecoder(r.Body).Decode(&data)
		if err != nil {
			rw.WriteHeader(http.StatusBadRequest)
			return
		}
		user, err := getUser(data.Email, client)
		if err != nil {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			payload := struct {
				Error string `json:"error"`
			}{Error: "Invalid email or password"}

			json.NewEncoder(rw).Encode(payload)
			return
		}
		if (*user).Password != data.Password {
			rw.Header().Set("Content-Type", "application/json")
			rw.WriteHeader(http.StatusBadRequest)
			payload := struct {
				Error string `json:"error"`
			}{Error: "Invalid password"}

			json.NewEncoder(rw).Encode(payload)
			return
		}
		rw.Header().Set("Content-Type", "application/json")
		rw.WriteHeader(http.StatusOK)
		json.NewEncoder(rw).Encode(user)
	}))
	log.Println("HTTP server started on :4000")
	err := http.ListenAndServe(":4000", nil)
	if err != nil {
		log.Fatal("Error starting server:", err)
	}
}