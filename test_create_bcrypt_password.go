package main

import(
	"fmt"

	"manyangcloud_utils"
)

func main()  {
	passwd := manyangcloud_utils.GenerateUserPassword("ManyangKawai")
	fmt.Println(passwd)
}