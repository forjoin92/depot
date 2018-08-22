package service

import (
	"testing"
	"strings"
	"fmt"
)

func TestParseUrl(t *testing.T) {
	url := "127.0.0.1:30444"
	port := "90" + strings.Split(url, ":")[1][3:]
	fmt.Println(port)
}
