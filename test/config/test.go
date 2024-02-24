package main

import (
	"fmt"
	"lglogo"
	"log"
	"os"
)

func main() {
	var config lglogo.Config

	fd, err := os.Open("pic/info.toml")
	if err != nil {
		log.Fatalln(err)
	}
	defer fd.Close()

	config.Read(fd)

	fmt.Printf("%#v\n", config)
	fmt.Println("imagedata num : ", len(config.ImageData))
}
