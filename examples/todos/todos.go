package main

import (
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/CzarSimon/jsonstore-go-client/jsonstore"
)

const (
	ListCommand     = "ls"
	AddCommand      = "add"
	CompleteCommand = "complete"
	DeleteCommand   = "delete"
	TodoKey         = "todos"
)

type Env struct {
	db       *jsonstore.HttpClient
	metadata Metadata
}

func getEnv() *Env {
	db := jsonstore.NewClient(getStoreToken())
	var metadata Metadata
	err := db.Get("metadata", &metadata)
	if err != nil {
		fmt.Printf("Could not get todos metadata. Error: %s\n", err)
		os.Exit(1)
	}
	return &Env{
		db:       db,
		metadata: metadata,
	}
}

func getStoreToken() string {
	token := os.Getenv("JSONSTORE_TOKEN")
	if token == "" {
		fmt.Println("No jsonstore token found")
		os.Exit(1)
	}
	return token
}

func main() {
	_ = getEnv()
	subCommand, err := getSubCommand()
	if err != nil {
		fmt.Println(err)
	}
	switch subCommand {
	case ListCommand:
		fmt.Println(subCommand)
	case AddCommand:
		fmt.Println(subCommand)
	case CompleteCommand:
		fmt.Println(subCommand)
	case DeleteCommand:
		fmt.Println(subCommand)
	default:
		printHelp()
	}
}

type Todo struct {
	ID    int       `json:"id"`
	Title string    `json:"title"`
	Done  bool      `json:"done"`
	Date  time.Time `json:"date"`
}

type Metadata struct {
	NextId int `json:"nextId"`
}

func getSubCommand() (string, error) {
	if len(os.Args) < 2 {
		return "", errors.New("No subcommand provided")
	}
	return os.Args[1], nil
}

func printHelp() {
	fmt.Println("todos - simple todo tool to demonstrate jsonstore go client")
	fmt.Printf("\n%s       - lists all todos active todos\n", ListCommand)
	fmt.Printf("%s      - adds a new todo\n", AddCommand)
	fmt.Printf("%s - marks a todo as completed\n", CompleteCommand)
	fmt.Printf("%s   - deletes a todo\n", DeleteCommand)
}
