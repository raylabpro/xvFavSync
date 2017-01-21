package main

import (
	"log"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var okForTest bool

func init() {
	applicationExitFunction = func(c int) { okForTest = false }
}

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

//TestFailedInitConfigs - negative test
func TestFailedInitConfigsWhenFileNotExist(t *testing.T) {
	configPath = "./config.not.exists"
	initConfigs()
	if okForTest == true {
		t.Error("")
	}
	okForTest = true
}

//TestFailedInitConfigs - negative test
func TestFailedInitConfigs(t *testing.T) {
	configPath = "./config.wrong.yml"
	initConfigs()
	if okForTest == true {
		t.Error("")
	}
	okForTest = true
}

func TestInitConfigs(t *testing.T) {
	configPath = "./config.smpl.yml"
	initConfigs()
}
func TestCheckErrAndExit(t *testing.T) {
	err := error("error")
	checkErrAndExit(err)
	if okForTest == true {
		t.Error("")
	}
	okForTest = true

}
