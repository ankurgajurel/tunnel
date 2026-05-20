package main

import (
	"fmt"

	"github.com/ankurgajurel/tunnel/internal/server"
)

func main() {
	srv := server.New(":8080")

	fmt.Println("server is listening in :8080")

	err := srv.ListenAndServe()
	if err != nil {
		fmt.Println("server error", err)
	}
}
