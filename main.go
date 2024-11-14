package main

import (
	"fmt"
	"net/http"
)

func main() {
	http.HandleFunc("/post", handlePostRequest)
	http.HandleFunc("/get", handleGetRequest)
	fmt.Println("Server listening on :8088")
	http.ListenAndServe(":8088", nil)
}

//func main() {
//	cli := CLI{}
//	cli.Run()
//
//}
