package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gorilla/mux"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"golang.org/x/crypto/bcrypt"
)

//--------------------------------------- Collection Model for users Collection ---------------------//
type Users struct {
	ID       primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Name     string             `json:"name,omitempty" bson:"name,omitempty"`
	Email    string             `json:"email,omitempty" bson:"email,omitempty"`
	Password string             `json:"password,omitempty" bson:"password,omitempty"`
}

//--------------------------------------- Collection Model for posts Collection ---------------------//
type Posts struct {
	ID        primitive.ObjectID `json:"_id,omitempty" bson:"_id,omitempty"`
	Captions  string             `json:"captions,omitempty" bson:"captions,omitempty"`
	ImageURL  string             `json:"image_url,omitempty" bson:"image_url,omitempty"`
	UserID    string             `json:"userid,omitempty" bson:"userid,omitempty"`
	Timestamp time.Time          `json:"timestamp,omitempty" bson:"timestamp,omitempty"`
}

//------------------------------ Function Module to Encrypt passwords ----------------------------------//

func encrypt(pwd []byte) string {
	hash, err := bcrypt.GenerateFromPassword(pwd, bcrypt.MinCost)
	if err != nil {
		log.Println(err)
	}
	return string(hash)
}

//------------------------------ Function Module to Encrypt passwords ----------------------------------//

func comparePasswords(hashedPwd string, plainPwd []byte) bool {
	byteHash := []byte(hashedPwd)
	err := bcrypt.CompareHashAndPassword(byteHash, plainPwd) //compare stored hash password with another password
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

var client *mongo.Client

//------------------------------ Function Module to Create a User ----------------------------------//

func CreateUserEndpoint(response http.ResponseWriter, request *http.Request) {
	fmt.Println("Create Accessed")
	response.Header().Add("content-type", "application/json")
	var users Users
	json.NewDecoder(request.Body).Decode(&users)
	// var temp = users.Password
	users.Password = encrypt([]byte(users.Password))
	// fmt.Println(comparePasswords(users.Password, []byte(temp)))
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017") //DB credentials
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)     //closes connection to DB after 10 seconds
	client, err := mongo.Connect(ctx, clientOptions)                        //DB connection establishment
	collection := client.Database("instagram").Collection("users")
	result, err := collection.InsertOne(ctx, users) //Insert Database Operation
	if err != nil {
		log.Fatal(err)
		fmt.Println(err.Error())
	}
	json.NewEncoder(response).Encode(result)
}

//------------------------------ Function Module to get a user using User's ID ----------------------------------//

func GetUserEndpoint(response http.ResponseWriter, request *http.Request) {
	fmt.Println("Get Accessed")
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request) //to get the param ID
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var result Users
	var filter = Users{ID: id} //filter to find the coressponding data from the MongoDB
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(ctx, clientOptions)
	collection := client.Database("instagram").Collection("users")
	err := collection.FindOne(ctx, filter).Decode(&result) //Find Database Operation
	if err != nil {
		fmt.Println("error")
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(result)
}

//------------------------------ Function Module to Create a Post ----------------------------------//

func CreatePostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	var post Posts
	json.NewDecoder(request.Body).Decode(&post)
	post.Timestamp = time.Now() //generate timestamp for the post while creation
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ := mongo.Connect(ctx, clientOptions)
	collection := client.Database("instagram").Collection("posts")
	result, err := collection.InsertOne(ctx, post)
	if err != nil {
		log.Fatal(err)
		fmt.Println(err.Error())
	}
	json.NewEncoder(response).Encode(result)
}

//------------------------------ Function Module to get a Post using Post's ID ----------------------------------//

func GetPostEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	id, _ := primitive.ObjectIDFromHex(params["id"])
	var post Posts
	var filter = Posts{ID: id}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	ctx, _ := context.WithTimeout(context.Background(), 10*time.Second)
	client, _ := mongo.Connect(ctx, clientOptions)
	collection := client.Database("instagram").Collection("posts")
	err := collection.FindOne(ctx, filter).Decode(&post)
	if err != nil {
		fmt.Println("error")
		response.WriteHeader(http.StatusInternalServerError)
		response.Write([]byte(`{"message": "` + err.Error() + `" }`))
		return
	}
	json.NewEncoder(response).Encode(post)
}

//------------------------------ Function Module to Get posts by Specfic UserID ----------------------------------//

func GetPostsByUsersEndpoint(response http.ResponseWriter, request *http.Request) {
	response.Header().Add("content-type", "application/json")
	params := mux.Vars(request)
	var id = params["id"]
	var posts []*Posts
	var filter = Posts{UserID: id}
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017")
	client, _ := mongo.Connect(context.TODO(), clientOptions)
	collection := client.Database("instagram").Collection("posts")
	var page, _ = strconv.Atoi(request.URL.Query().Get("page")) //access page query from URL
	var perPage int64 = 5                                       //Items per page Set to 5
	// fmt.Println(page, perPage)
	findOptions := options.Find()
	findOptions.SetLimit(perPage)
	findOptions.SetSkip((int64(page)) * perPage) //setting offset for data search
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
		posts = append(posts, &post) //Add all results to the response
	}
	if err := cur.Err(); err != nil {
		log.Fatal(err)
	}
	cur.Close(context.TODO()) //cursor closes
	json.NewEncoder(response).Encode(posts)

}

//------------------------------ Main Function Module handles routes -------------------------------------------//

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
	router := mux.NewRouter() //Router Set up for API endpoints
	router.HandleFunc("/users", CreateUserEndpoint).Methods("POST")
	router.HandleFunc("/users/{id}", GetUserEndpoint).Methods("GET")
	router.HandleFunc("/posts", CreatePostEndpoint).Methods("POST")
	router.HandleFunc("/posts/{id}", GetPostEndpoint).Methods("GET")
	router.HandleFunc("/posts/users/{id}", GetPostsByUsersEndpoint).Methods("GET")
	http.ListenAndServe(":12345", router)
}

//Author: Abhijeet Chatterjee
