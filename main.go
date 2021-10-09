package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Users struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

type Posts struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Captions  string             `json:"captions,omitempty" bson:"captions,omitempty"`
	ImageURL  string             `json:"image_url,omitempty" bson:"image_url,omitempty"`
	UserID    string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Timestamp time.Time          `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}

var client *mongo.Client

func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	fmt.Println("Create Accessed")
	response.Header().Add("content-type", "application/json")
	var users Users
	json.NewDecoder(request.Body).Decode(&users)
	fmt.Println(users)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	collection := client.Database("instagram").Collection("users")
	fmt.Println(users)
	result, err := collection.InsertOne(context.TODO(), users)
	fmt.Println(result.InsertedID)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err.Error())
	}
	json.NewEncoder(response).Encode(result)
}

func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	fmt.Println("Get Accessed")
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var result Users
	var filter = Users{ID: id}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	collection := client.Database("instagram").Collection("users")
	err := collection.FindOne(context.TODO(), filter).Decode(&result)
	if err != nil {
		fmt.Println("error")
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

func CreatePostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var post Posts
	json.NewDecoder(request.Body).Decode(&post)
	fmt.Println(request.Body, "\n", post)
	post.Timestamp = time.Now()
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	collection := client.Database("instagram").Collection("posts")
	// ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	// result, _ := collection.InsertOne(ctx, post)
	fmt.Println(post)
	result, err := collection.InsertOne(context.TODO(), post)
	fmt.Println(result.InsertedID)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err.Error())
	}
	// json.NewEncoder(response).Encode(result.InsertedID)
	json.NewEncoder(response).Encode(result)
}

func GetPostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var post Posts
	var filter = Posts{ID: id}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	collection := client.Database("instagram").Collection("posts")
	err := collection.FindOne(context.TODO(), filter).Decode(&post)
	if err != nil {
		fmt.Println("error")
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(post)
}

func GetPostsByUsersEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	// id, _ := primitive.ObjectIDFromHex(params["id"])
	var id = params["id"]
	var posts []*Posts
	var filter = Posts{UserID: id}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	collection := client.Database("instagram").Collection("posts")
	findOptions := options.Find()
	cur, err := collection.Find(context.TODO(), filter, findOptions)
	if err != nil {
		log.Fatal(err)
	}
	for cur.Next(context.TODO()) {
		// create a value into which the single document can be decoded
		var post Posts
		err := cur.Decode(&post)
		if err != nil {
			log.Fatal(err)
		}
		posts = append(posts, &post)
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.TODO())
	json.NewEncoder(response).Encode(posts)

}

func main() {
	fmt.Println("Starting the application...")
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, err := mongo.Connect(context.TODO(), clientOptions)
	if err != nil {
		fmt.Println("Error connecting to Mongo")
		log.Fatal(err)
	}

	// Check the connection
	err = client.Ping(context.TODO(), nil)

	if err != nil {
		fmt.Println("Error Pinging to Mongo")
		log.Fatal(err)
	}

	fmt.Println("Connected to MongoDB!")
	router := mux.NewRouter()
	router.HandleFunc("/users", CreateUserEndpoint).Methods("POST")
	router.HandleFunc("/users/{id}", GetUserEndpoint).Methods("GET")
	router.HandleFunc("/posts", CreatePostEndpoint).Methods("POST")
	router.HandleFunc("/posts/{id}", GetPostEndpoint).Methods("GET")
	router.HandleFunc("/posts/users/{id}", GetPostsByUsersEndpoint).Methods("GET")
	http.ListenAndServe(":12345", router)
}
