//https://github.com/dgrijalva/jwt-go/blob/master/http_example_test.go
package manyangcloud_config
import (
	"fmt"
	"crypto/rsa"
	
	jwtgo "github.com/dgrijalva/jwt-go"
	"manyangcloud_fs"
	//"manyangcloud_jwt"
)
var (
	PubKeyFile	*rsa.PublicKey
	PrivKeyFile *rsa.PrivateKey	
)
//these can also be set...
const (
	PKPWD = "ManyangKawai"
	
	KeyCertPath = "/var/www/keycertz/"
	PrivKeyPath = "/var/www/keycertz/mykey.pem"
	PubKeyPath  = "/var/www/keycertz/mykey.pub"	
	
	//dont forget to escape characters like @ w/ %40
	MongoHost = "127.0.0.1"
	MongoUser = "mongod"
	MongoPassword = "SOMEHARDPASSWORD"
	MongoDb = "admin"
	
	RedisRP = "SOMELOGASSPASSWORD HERE"
	
	MysqlPass = "ANOTHER-HARD-PASSOWRD"	
)

func init() {
	f,ok,err := manyangcloud_fs.ReadFile(PubKeyPath)

	if (!ok || err != nil) { fmt.Println(err) } else {
		//PubKeyFile, err = manyangcloud_jwt.ParseRSAPublicKeyFromPEM(f)
		PubKeyFile, err = jwtgo.ParseRSAPublicKeyFromPEM(f)
		if err != nil { fmt.Println(err) }	
	}
 
	f,ok,err = manyangcloud_fs.ReadFile(PrivKeyPath)	
	if (!ok || err != nil) { fmt.Println(err) } else {
		//PrivKeyFile, err = manyangcloud_jwt.ParseRSAPrivateKeyFromPEMWithPassword(f, PKPWD)
		PrivKeyFile, err = jwtgo.ParseRSAPrivateKeyFromPEMWithPassword(f, PKPWD)
		if err != nil { fmt.Println(err) }
	}	
}
