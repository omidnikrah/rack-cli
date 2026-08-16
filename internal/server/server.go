package server

import (
	"fmt"
	"rack-cli/internal/entity"
)

type Storage interface {
	Load() ([]entity.Server, error)
	Save(server entity.Server) error
}

type InputReader interface {
	ReadInput(label string) string
	ReadInputWithDefault(label, defaultValue string) string
	ReadList(label string) []string
}

func Add(storage Storage, in InputReader) {
	label := in.ReadInput("Label")
	ip := in.ReadInput("IP address")
	rootUser := in.ReadInputWithDefault("Root user", "root")
	password := in.ReadInput("Password")
	purpose := in.ReadInput("Purpose")
	tags := in.ReadList("Tags")

	newServer := entity.Server{
		Label:    label,
		IP:       ip,
		RootUser: rootUser,
		Password: password,
		Purpose:  purpose,
		Tags:     tags,
	}

	sErr := storage.Save(newServer)
	if sErr != nil {
		fmt.Println(sErr)
		return
	}
	fmt.Printf("Added server: %+v\n", newServer)
}

func Edit(storage Storage, server entity.Server) {
	fmt.Printf("You chose edit server: %d \n", server.ID)
}

func Delete() {
	fmt.Println("You chose delete server")
}

func View() {
	fmt.Println("You chose view server")
}

func Ping() {
	fmt.Println("You chose ping server")
}

func SSH() {
	fmt.Println("You chose ssh server")
}

func List() {
	fmt.Println("You chose list servers")
}
