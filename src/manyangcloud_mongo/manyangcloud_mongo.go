package manyangcloud_mongo

import(
	"fmt"
	//"sync"
	"context"
	//"net/http"
	//"encoding/json"
	
	//"github.com/gorilla/websocket"
	
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"	
	
	"manyangcloud_data"
	"manyangcloud_utils"
	"manyangcloud_config"	
)

type key string
//https://stackoverflow.com/questions/54627542/official-mongo-go-driver-using-sessions
const (
	HostKey     = key("hostKey")
	UsernameKey = key("usernameKey")
	PasswordKey = key("passwordKey")
	DatabaseKey = key("databaseKey")	
)

var ctx context.Context;
var client *mongo.Client;

func init()  {
	ctx = context.Background()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	
	ctx = context.WithValue(ctx, HostKey, manyangcloud_config.MongoHost)
	ctx = context.WithValue(ctx, UsernameKey, manyangcloud_config.MongoUser)
	ctx = context.WithValue(ctx, PasswordKey, manyangcloud_config.MongoPassword)
	ctx = context.WithValue(ctx, DatabaseKey, manyangcloud_config.MongoDb)

	uri := fmt.Sprintf(`mongodb://%s:%s@%s/%s`,
		ctx.Value(UsernameKey).(string),
		ctx.Value(PasswordKey).(string),
		ctx.Value(HostKey).(string),
		ctx.Value(DatabaseKey).(string),
	)
	clientOptions := options.Client().ApplyURI(uri)
	
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	
	// Check the connection
	err = client.Ping(ctx, nil)
	if err != nil { fmt.Println(err); } else { fmt.Println("Mongo Connected"); }
}

func MongoTryUser(u []byte, p []byte) (bool,*manyangcloud_data.AUser,error) {
	var xdoc manyangcloud_data.AUser
	collection := client.Database("api").Collection("users");
	filter := bson.D{{"user", string(u)}}
	if err := collection.FindOne(ctx, filter).Decode(&xdoc); err != nil {
	    return false,nil, err
    } else {
		bres, err := manyangcloud_utils.ValidateUserPassword(p, []byte(xdoc.Password))
	    if err != nil { return false, nil, err } else {  return bres, &xdoc, nil }
    }		    
}
