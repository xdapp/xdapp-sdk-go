package main

import (
	"./register"
)

func main() {

	myReg := register.NewRegister()
	myReg.LoadService()
	myReg.InitClient()
}