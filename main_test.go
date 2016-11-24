package main

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogPrintln(t *testing.T) {
	log.Println("Starting tests!")
	os.Setenv("TESTING", "YES")
}

func TestPrintObject(t *testing.T) {
	data := []string{"test", "data"}
	result := printObject(data)
	ass := assert.New(t)
	ass.NotEmpty(result)
}
