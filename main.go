package main

import (
	"fmt"
	"github.com/janosgyerik/portping"
	"time"
)

func main() {

	timeout := time.Duration(1) * time.Second
	r :=portping.Ping("tcp","127.0.0.1:80",  timeout)
	fmt.Println(r)
	r =portping.Ping("tcp","127.0.0.1:443",  timeout)
	fmt.Println(r)
	r =portping.Ping("tcp","127.0.0.1:8990",  timeout)
	fmt.Println(r)
}
