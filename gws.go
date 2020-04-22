package main

import (
	"fmt"
	"log"
	"flag"
	"sync"
	"net/http"
	//"io/ioutil"
	"encoding/json"

	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"manyangcloud_jwt"
	"manyangcloud_utils"
	"manyangcloud_mongo"
	"manyangcloud_config"
	"manyangcloud_asyncq"	
	"manyangcloud_wspty"	
)

var addr = flag.String("addr", "0.0.0.0:1200", "http service address")
var upgrader = websocket.Upgrader{}

type msg struct {
	Jwt string `json:"jwt"`
	Type string `json:"type"`
	Data string	`json:"data"`
}

func sendMsg(j string, t string, d string, c *websocket.Conn) {
	m := msg{j, t, d};
	if err := c.WriteJSON(m); err != nil {
		fmt.Println(err)
	}
	//mm, _ := json.Marshal(m);
	//fmt.Println(string(mm));
}

func handleAPI(w http.ResponseWriter, r *http.Request) {
	//swp := r.Header.Get("Sec-Websocket-Protocol");	
	upgrader.CheckOrigin = func(r *http.Request) bool { return true }
	c, err := upgrader.Upgrade(w, r, nil) //add rh later
	
	if err != nil {
		fmt.Print("WTF @HandleAPI Ws Upgrade Error> ", err)
		return
	}
	Loop:
		for {
			in := msg{}
			
			err := c.ReadJSON(&in)
			if err != nil {
				fmt.Println("Error reading json.", err)
				c.Close()
				break Loop
			}
			switch (in.Type) {
				case "get-jwt-token":
					fmt.Println(in.Data)

					usr, pwd, err := manyangcloud_utils.B64DecodeTryUser(in.Data);
					if err != nil { fmt.Println(err);  } else {  fmt.Println(string(usr), string(pwd)) } 

					upv, auser, err := manyangcloud_mongo.MongoTryUser(usr,pwd)
					
					if err != nil { fmt.Println(err); sendMsg("noop", "invalid-credentials","noop", c); } else { 
						if upv == true { fmt.Println("A user has logged in."); }
						auser.Password = "F00"
						jauser,err := json.Marshal(auser); if err != nil { fmt.Println("error marshaling AUser.") } else {
							jwt, err := manyangcloud_jwt.GenerateJWT(manyangcloud_config.PrivKeyFile);
					 		if err != nil { fmt.Println(err);  } else  { sendMsg(jwt, "jwt-token", string(jauser), c);  }								
						}
					}	
					break;
					
				case "verify-jwt-token": fallthrough
				case "validate-stored-jwt-token":
					valid, err := manyangcloud_jwt.ValidateJWT(manyangcloud_config.PubKeyFile,in.Jwt)	
					if err != nil { fmt.Println(err); sendMsg("^vAr^", "jwt-token-invalid",err.Error(), c) } else if (err == nil && valid ) {	
					    if in.Type == "verify-jwt-token" { sendMsg("^vAr^", "jwt-token-valid","noop", c) }
					    if in.Type == "validate-stored-jwt-token" {  sendMsg("^vAr^", "stored-jwt-token-valid","noop", c) }
					}
					break;
				case "rapid-test-user-avail":
					tobj := manyangcloud_mongo.NewRapidTestUserAvailTask(in.Data, c);
					manyangcloud_asyncq.TaskQueue <- tobj					
					break;
				case "create-user":
					tobj := manyangcloud_mongo.NewCreateUserTask(in.Data, c);
					manyangcloud_asyncq.TaskQueue <- tobj
					break;
				
				default:
					break;
			}
		}
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	component := params["component"]
	subcomponent := params["subcomponent"]
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Content-Type", "application/json");
	fmt.Println(component);
	var wg sync.WaitGroup	
	wg.Add(1)
	
	tobj :=  manyangcloud_mongo.NewGetDocumentsTask("UI",component,subcomponent, w, &wg); //notice the pointer to the wait group
	manyangcloud_asyncq.TaskQueue <- tobj
		
	wg.Wait(); 
	fmt.Println("Wait Group Finished Success...");
}

func main() {
	
	manyangcloud_asyncq.StartTaskDispatcher(9)
	//go manyangcloud_sse.ConfigureSystemHeartbeat()
	//go manyangcloud_sse.StartSSE()
	flag.Parse()
	log.SetFlags(0)	
	
	//look into subrouter stuffs
	r := mux.NewRouter()

	//Websocket API
    r.HandleFunc("/api", handleAPI)
	r.HandleFunc("/pty", manyangcloud_wspty.HandleWsPty)
	//Rest API
	r.HandleFunc("/rest/api/ui/{component}/{subcomponent}", handleUI)
	
	http.ListenAndServeTLS(*addr,"/etc/letsencrypt/live/manyangcloud.com/cert.pem", "/etc/letsencrypt/live/manyangcloud.com/privkey.pem" , r)

	
}