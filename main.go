package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		startServer()
		return
	}

	// argsWithoutProg := os.Args[1:]
	startClient()
}

func startClient() {
	panic("not implemented")
}

func startServer() {
	srv := &Server{
		port: 8080,
		usersHandler: &usersHandler{
			users: make([]User, 0),
		},
		roomHandler: &roomHandler{
			rooms: make([]*Room, 0),
		},
	}
	err := srv.Run()
	if err != nil {
		log.Fatalf("error: %s", err.Error())
	}
}

/*
- Server
*/
type Server struct {
	port int
	// list of users
	// list of rooms
	usersHandler *usersHandler
	roomHandler  *roomHandler
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	// TODO: use a better mux that allows for specifying methods
	mux.Handle("/foo", timeHandler{})
	mux.HandleFunc("/users", s.usersHandler.GetAll)
	mux.HandleFunc("/user/add", s.usersHandler.Add)
	mux.HandleFunc("/room/add", s.roomHandler.Add)
	mux.HandleFunc("/room/user/add", s.roomHandler.AddUser)
	mux.HandleFunc("/room/names", s.roomHandler.GetAllNames)
	mux.HandleFunc("/room/users", s.roomHandler.GetAllUsers)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", s.port),
		Handler: mux,
	}
	return httpServer.ListenAndServe()
}

type usersHandler struct {
	users []User
}

func (uh *usersHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var users string
	for i, user := range uh.users {
		comma := ""
		if i > 0 {
			comma = ","
		}
		users = users + comma + user.Name
	}
	w.Write([]byte(users))
}

func (uh *usersHandler) Add(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error reading body: %s", err.Error())))
		return
	}
	username := string(data)
	uh.users = append(uh.users, User{
		Name: username,
	})
	log.Printf("user added: %s, %#v", username, uh.users)
	w.WriteHeader(http.StatusNoContent)
}

type roomHandler struct {
	// TODO: need to split the handler struct from the yet to be created data struct...
	rooms []*Room

	// users    []User // TODO: set this to the same []User slice from the users handler
}

type Room struct {
	name     string
	users    []User
	messages []string
}

func (rh *roomHandler) Add(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error reading body: %s", err.Error())))
		return
	}
	roomName := string(data)

	rh.rooms = append(rh.rooms, &Room{
		name: roomName,
	})
	log.Printf("room added: %s, %#v", roomName, rh.rooms)
	w.WriteHeader(http.StatusNoContent)
}

func (rh *roomHandler) AddUser(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error reading body: %s", err.Error())))
		return
	}

	args := strings.Split(string(data), ",")
	if len(args) != 2 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("expected 2 args comma separate (roomName, username): %s", err.Error())))
		return
	}

	roomName, username := args[0], args[1]

	// TODO: check if users exists... for _, user := range users {}

	foundRoom := false
	for _, room := range rh.rooms {
		if room.name != roomName {
			continue
		}
		foundRoom = true

		// TODO: get the user struct from the main list of users
		room.users = append(room.users, User{
			Name: username,
		})
	}
	if !foundRoom {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("room does not exist: %s", roomName)))
		return
	}

	log.Printf("user %q added to room %q, %#v", username, roomName, rh.rooms)
	w.WriteHeader(http.StatusNoContent)
}

func (rh *roomHandler) GetAllNames(w http.ResponseWriter, r *http.Request) {
	var rooms string
	for i, room := range rh.rooms {
		comma := ""
		if i > 0 {
			comma = ","
		}
		rooms = rooms + comma + room.name
	}
	w.Write([]byte(rooms))
}

func (rh *roomHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error reading body: %s", err.Error())))
		return
	}
	roomName := string(data)

	var users string
	foundRoom := false
	for _, room := range rh.rooms {
		if room.name != roomName {
			continue
		}
		foundRoom = true

		for i, user := range room.users {
			comma := ""
			if i > 0 {
				comma = ","
			}
			users = users + comma + user.Name
		}
	}
	if !foundRoom {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("room does not exist: %s", roomName)))
		return
	}

	w.Write([]byte(users))
}

// TODO: replace room message strings with timeMessages
type timeMessage struct {
	time    time.Time
	message string
}

type timeHandler struct {
}

func (th timeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	tm := time.Now()
	w.Write([]byte("The time is: " + tm.String()))
}

/*
- Client
*/
type Client struct {
	serverAddress string
}

func (c *Client) getUsers() []User {
	var result []User
	return result
}

type User struct {
	Name string
}

// func getRooms() {}
