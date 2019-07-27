package controller

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"

	"github.com/Maxbrain0/echo_mongo/model"
	"github.com/labstack/echo/v4"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
)

// Users holds reference to a database collection and is the receiver of various
// endpoint controllers which will need mongoDB collection access
type Users struct {
	Collection *mongo.Collection
}

// CreateUser creates a user in mongo dB and returns a response on success
func (user *Users) CreateUser(c echo.Context) error {
	u := new(model.User)

	if err := c.Bind(u); err != nil {
		return err
	}

	// make sure username and password are in request
	if len(u.UserName) < 1 || len(u.Password) < 1 {
		return echo.NewHTTPError(http.StatusBadRequest, "Please provide a user name and password")
	}

	// determine if userName already exists from find count
	filter := bson.M{"userName": u.UserName}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, cErr := user.Collection.CountDocuments(ctx, filter)

	if cErr != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not add user")
	}

	if count != 0 {
		return echo.NewHTTPError(http.StatusConflict, "User already exists")
	}

	// Create a hashed password
	hashedPW, err := bcrypt.GenerateFromPassword([]byte(u.Password), 10)

	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not add user")
	}

	// attempt to insert into the database
	ctx2, cancel2 := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel2()
	res, err := user.Collection.InsertOne(ctx2, bson.M{"userName": u.UserName, "password": string(hashedPW), "email": u.Email})

	if err != nil {
		fmt.Println(err)
		return echo.NewHTTPError(http.StatusInternalServerError, "Could not add user")
	}

	oid := res.InsertedID.(primitive.ObjectID)

	response := &model.User{
		ID:       oid,
		UserName: u.UserName,
	}

	return c.JSON(http.StatusCreated, response)
}

// Login receives the username and password from from the json request body and determines if the user exist
// It then compares hashed password, and if successful, returns userName and jwt
// func (user *Users) Login(c echo.Context) error {
// 	u := new(model.User)

// 	if err := c.Bind(u); err != nil {
// 		return err
// 	}
// }
