package main

import(
	"fmt"
	"manyangcloud_fs"
	"manyangcloud_config"
	"manyangcloud_genkeys"
)

func main() {
	
	pk, err := manyangcloud_genkeys.PrivateKeyToEncryptedPEM(1028, "ManyangKawai")
	if err != nil {
		fmt.Println(err);
	}

	f, ok, err := manyangcloud_fs.CreateFile(manyangcloud_config.KeyCertPath, "mykey.pem")
	if !ok {
		fmt.Println(err)
	} else {
		manyangcloud_fs.WriteFile(f, pk)
	}

}