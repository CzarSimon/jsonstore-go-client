package main

import (
	"errors"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/CzarSimon/jsonstore-go-client/jsonstore"
)

const (
	ListCommand     = "ls"
	AddCommand      = "add"
	CompleteCommand = "complete"
	DeleteCommand   = "delete"
	HelpCommand     = "help"
	TodoKey         = "todos"
)

type Env struct {
	jsonstore jsonstore.Client
	metadata  Metadata
}

func (env *Env) nextId() int {
	nextId := env.metadata.NextId
	err := env.jsonstore.Put("metadata/nextId", nextId+1)
	if err == nil {
		env.metadata.NextId = nextId + 1
	}
	return nextId
}

func (env *Env) addTodo() error {
	title, err := getCommandAt(2)
	if err != nil {
		return errors.New("No todo title provided")
	}
	todoId := env.nextId()
	todo := NewTodo(todoId, title)
	err = env.jsonstore.Post(fmt.Sprintf("todos/%d", todoId), todo)
	if err != nil {
		return err
	}
	fmt.Printf("Todo: '%s' added\n", title)
	return nil
}

func (env *Env) listTodos() error {
	todos := make([]Todo, 0)
	err := env.jsonstore.Get("todos", &todos)
	if err != nil {
		return err
	}
	for _, todo := range filterTodos(todos) {
		fmt.Println(todo)
	}
	return nil
}

func (env *Env) completeTodo() error {
	ID := getIdFromArgs()
	err := env.jsonstore.Put(fmt.Sprintf("todos/%d/done", ID), true)
	if err != nil {
		return err
	}
	var todo Todo
	err = env.jsonstore.Get(fmt.Sprintf("todos/%d", ID), &todo)
	if err != nil {
		fmt.Printf("Todo with id %d set to done\n", ID)
	}
	fmt.Printf("'%s' set to done\n", todo.Title)
	return nil
}

func (env *Env) deleteTodo() error {
	ID := getIdFromArgs()
	err := env.jsonstore.Delete(fmt.Sprintf("todos/%d", ID))
	if err != nil {
		return err
	}
	fmt.Println("Todo deleted")
	return nil
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
		jsonstore: db,
		metadata:  metadata,
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
	env := getEnv()
	subCommand, err := getCommandAt(1)
	if err != nil {
		subCommand = ListCommand
		err = nil
	}
	switch subCommand {
	case ListCommand:
		err = env.listTodos()
	case AddCommand:
		err = env.addTodo()
	case CompleteCommand:
		err = env.completeTodo()
	case DeleteCommand:
		err = env.deleteTodo()
	case HelpCommand:
		printHelp()
	default:
		fmt.Printf("Unknown command: '%s'\n", subCommand)
		printHelp()
	}
	if err != nil {
		fmt.Println(err)
	}
}

type Todo struct {
	ID    int       `json:"id"`
	Title string    `json:"title"`
	Done  bool      `json:"done"`
	Date  time.Time `json:"date"`
}

func (todo Todo) String() string {
	return fmt.Sprintf("%d - %s", todo.ID, todo.Title)
}

func NewTodo(id int, title string) Todo {
	return Todo{
		ID:    id,
		Title: title,
		Done:  false,
		Date:  time.Now(),
	}
}

type Metadata struct {
	NextId int `json:"nextId"`
}

func getCommandAt(index int) (string, error) {
	if len(os.Args) < index+1 {
		return "", fmt.Errorf("No command at index: %d provided", index)
	}
	return os.Args[index], nil
}

func printHelp() {
	fmt.Println("todos - simple todo tool to demonstrate jsonstore go client")
	fmt.Printf("\n%s       - lists all todos active todos\n", ListCommand)
	fmt.Printf("%s      - adds a new todo\n", AddCommand)
	fmt.Printf("%s - marks a todo as completed\n", CompleteCommand)
	fmt.Printf("%s   - deletes a todo\n", DeleteCommand)
}

func getIdFromArgs() int {
	idStr, err := getCommandAt(2)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	ID, err := strconv.Atoi(idStr)
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	return ID
}

func filterTodos(todos []Todo) []Todo {
	filteredTodos := make([]Todo, 0)
	for _, todo := range todos {
		if !todo.Done && todo.Title != "" {
			filteredTodos = append(filteredTodos, todo)
		}
	}
	return filteredTodos
}
