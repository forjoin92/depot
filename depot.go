package main

import (
	"flag"
	"fmt"
)

func main() {
	cluster := flag.String("cluster", "127.0.0.1:30304", "Comma separated cluster peers")
	fmt.Println(cluster)
}
