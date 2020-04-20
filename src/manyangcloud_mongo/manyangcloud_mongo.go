package manyangcloud_mongo

import(
	"fmt"
	"sync"
	"context"
	"net/http"
	"encoding/json"
	
	"github.com/gorilla/websocket"
	
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


// Task code it's run by async
type GetDocumentsTask struct {
	what string
	component string
	subcomponent string
	w http.ResponseWriter
	wg *sync.WaitGroup
} 

func NewGetDocumentsTask(what string, component string, subcomponent string, w http.ResponseWriter, wg *sync.WaitGroup) *GetDocumentsTask {
	return &GetDocumentsTask{what, component, subcomponent, w, wg}
}

func (t *GetDocumentsTask) Perform() {
	defer t.wg.Done();
	switch(t.what) {
		case "UI":
			//var xdoc interface{} *** works but adds the key value thing for everything ***/
			var xdoc map[string]interface{}
			collection := client.Database("api").Collection("ui");	
			
			filter := bson.D{};
			switch(t.subcomponent) {
				case "noop":
					filter = bson.D{{"component", t.component}}
					break;
				default: 
					filter = bson.D{{"component", t.component}, {"for",t.subcomponent}}
					break;
			}
		    if err  := collection.FindOne(ctx, filter).Decode(&xdoc); err != nil {
			    fmt.Println(err); //mongo: no document in result.. so we could do else here...
		    } else {
			    //str := fmt.Sprintf("%v", xdoc)
			    if err != nil { fmt.Println(err); }	else {			    
				    json.NewEncoder(t.w).Encode(xdoc)  
			    }
		    }		    
			break;
	}
}

//test whethert new user is available or not
type RapidTestUserAvailTask struct {
	rtu string
	ws *websocket.Conn
}

func NewRapidTestUserAvailTask(rtu string, ws *websocket.Conn)  *RapidTestUserAvailTask {
	return	&RapidTestUserAvailTask{rtu, ws}
}

func (rtuat *RapidTestUserAvailTask) Perform() {
	collection := client.Database("api").Collection("users");
	
	var xdoc interface{}
	filter := bson.D{{"user", rtuat.rtu }}
	err := collection.FindOne(ctx, filter).Decode(&xdoc); 
	// if we found same data in db, it's fail.
    if (err != nil && xdoc == nil) { 
		manyangcloud_utils.SendMsg("vAr","rapid-test-user-avail-success", "noop", rtuat.ws);
	} else {
		manyangcloud_utils.SendMsg("vAr","rapid-test-user-avail-fail", "noop", rtuat.ws);
	}		
}

// Create User code
type CreateUserTask  struct {
	jsonuserstr string
	ws *websocket.Conn
}

func NewCreateUserTask(jsonuserstr string, ws *websocket.Conn)  *CreateUserTask {
	return	&CreateUserTask {jsonuserstr, ws}
}

func (cut *CreateUserTask) Perform() {
	user := manyangcloud_data.AUser{}
    err := json.Unmarshal([]byte(cut.jsonuserstr), &user);
    if err != nil {  fmt.Println(err) }else {
	    //fmt.Printf("%+v\n", user) DEBUG REDACTED
	    
	    collection := client.Database("api").Collection("users");	
	    //check if user already exists
		
		var xdoc interface{}
		filter := bson.D{{"user", user.User }}
	    err := collection.FindOne(ctx, filter).Decode(&xdoc); 
	    if (err != nil && xdoc == nil) { 
		    fmt.Println("User Available", err);
			
			hp := manyangcloud_utils.GenerateUserPassword(user.Password);
			user.Password = hp;
			insertResult, err := collection.InsertOne(ctx, &user)
			if err != nil { fmt.Println("Error Inserting Document"); } else {
				fmt.Println("Inserted a single User: ", insertResult.InsertedID)
				manyangcloud_utils.SendMsg("vAr","toast-success", "User Created Successfully", cut.ws);
			}
		} else {
			//shouldn't get here but it means some how rapid test didn't catch this
			//modal is still open so just display modal error...
			manyangcloud_utils.SendMsg("vAr","rapid-test-user-avail-fail", "User Already Exists!", cut.ws);
		}	    
    }
}