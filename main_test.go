package main

import (
	"log"
	"os"
	"testing"
)

func TestLogPrintln(t *testing.T) {
	log.Println("Starting tests!")
	os.Setenv("TESTING", "YES")
}
