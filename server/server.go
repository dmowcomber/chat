package server

import (
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"
)

type Server struct {
	port         int
	usersHandler *usersHandler
	roomHandler  *roomHandler
}

func NewServer(port int) *Server {
	users := &Users{
		users: make(map[string]*User),
	}
	rooms := NewRooms()
	return &Server{
		port: port,
		usersHandler: &usersHandler{
			users: users,
		},
		roomHandler: &roomHandler{
			rooms: rooms,
			users: users,
		},
	}
}

func (s *Server) Run() error {
	mux := http.NewServeMux()
	// TODO: use a better mux that allows for specifying methods
	mux.HandleFunc("/users", s.usersHandler.GetAll)
	mux.HandleFunc("/user/add", s.usersHandler.Add)
	mux.HandleFunc("/room/add", s.roomHandler.Add)
	mux.HandleFunc("/room/user/add", s.roomHandler.AddUser)
	mux.HandleFunc("/room/names", s.roomHandler.GetAllNames)
	mux.HandleFunc("/room/users", s.roomHandler.GetAllUsers)
	mux.HandleFunc("/room/message/add", s.roomHandler.AddMessage)
	mux.HandleFunc("/room/messages", s.roomHandler.GetMessages)

	httpServer := &http.Server{
		Addr:    fmt.Sprintf("0.0.0.0:%d", s.port),
		Handler: mux,
	}
	return httpServer.ListenAndServe()
}

type usersHandler struct {
	users *Users
}

// TODO: update endpoints to read and write json
func (uh *usersHandler) GetAll(w http.ResponseWriter, r *http.Request) {
	var users string
	var i = 0
	for _, user := range uh.users.users {
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

	if uh.users.exists(username) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("users already exists"))
		return
	}

	uh.users.add(username)
	log.Printf("user added: %s, %#v", username, uh.users)
	w.WriteHeader(http.StatusNoContent)
}

type roomHandler struct {
	rooms *Rooms
	users *Users
}

type User struct {
	Name string
}

type Users struct {
	users map[string]*User
}

func (u *Users) exists(username string) bool {
	_, exists := u.users[username]
	return exists
}

func (u *Users) add(username string) {
	u.users[username] = &User{
		Name: username,
	}
}

// func (u *Users) get(username string) *User {
// 	return u.users[username] // returns nil if not found
// }

type Room struct {
	name     string
	users    *Users
	messages []string
}

func NewRoom(name string) *Room {
	return &Room{
		name: name,
		users: &Users{
			users: make(map[string]*User),
		},
		messages: make([]string, 0),
	}
}

type Rooms struct {
	mu    *sync.Mutex
	rooms map[string]*Room
}

func NewRooms() *Rooms {
	return &Rooms{
		mu:    &sync.Mutex{},
		rooms: make(map[string]*Room),
	}
}

func (r *Rooms) add(name string) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.rooms[name]
	if exists {
		return errors.New("room already exists")
	}

	r.rooms[name] = NewRoom(name)
	return nil
}

func (r *Rooms) get(name string) *Room {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.rooms[name] // returns nil if not found
}

func (r *Rooms) exists(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	_, exists := r.rooms[name]
	return exists
}

func (rh *roomHandler) Add(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error reading body: %s", err.Error())))
		return
	}
	roomName := string(data)

	if rh.rooms.exists(roomName) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("room already exists"))
		return
	}

	rh.rooms.add(roomName)

	log.Printf("room added: %s, total rooms: %d", roomName, len(rh.rooms.rooms))
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
		w.Write([]byte("expected 2 args comma separate (roomName, username)"))
		return
	}
	roomName, username := args[0], args[1]

	if !rh.users.exists(username) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("user does not exist %s", username)))
		return
	}

	if !rh.rooms.exists(roomName) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("room does not exist: %s", roomName)))
		return
	}

	room := rh.rooms.get(roomName)
	if room.users.exists(username) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("user already in room"))
		return
	}
	room.users.add(username)

	log.Printf("user %q added to room %q, %d", username, roomName, len(rh.rooms.rooms))
	w.WriteHeader(http.StatusNoContent)
}

func (rh *roomHandler) GetMessages(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error reading body: %s", err.Error())))
		return
	}

	args := strings.Split(string(data), ",")
	if len(args) != 1 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected 1 arg (roomName)"))
		return
	}

	roomName := args[0]

	if !rh.rooms.exists(roomName) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("room does not exist: %s", roomName)))
		return
	}

	room := rh.rooms.get(roomName)
	messages := strings.Join(room.messages, "\n")
	respBody := fmt.Sprintf("room %q has %d messages:\n%s", roomName, len(room.messages), messages)
	w.Write([]byte(respBody))
}

func (rh *roomHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("error reading body: %s", err.Error())))
		return
	}

	args := strings.Split(string(data), ",")
	if len(args) != 3 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte("expected 3 args comma separate (roomName, username, message)"))
		return
	}
	roomName, username, message := args[0], args[1], args[2]

	// TODO: check if users exists... for _, user := range users {}

	if !rh.rooms.exists(roomName) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("room does not exist: %s", roomName)))
		return
	}

	// TODO: keep track of time, username, and messageStr in a struct
	room := rh.rooms.get(roomName)
	clockTime := time.Now().Format(time.Kitchen)
	message = fmt.Sprintf("%s %20s: %s", clockTime, username, message)
	room.messages = append(room.messages, message)

	log.Printf("message %q added to room %q, total message: %d", message, roomName, len(room.messages))
	w.WriteHeader(http.StatusNoContent)
}

func (rh *roomHandler) GetAllNames(w http.ResponseWriter, r *http.Request) {
	var rooms string
	var i = 0
	for _, room := range rh.rooms.rooms {
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
	if !rh.rooms.exists(roomName) {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(fmt.Sprintf("room does not exist: %s", roomName)))
		return
	}

	var users string
	room := rh.rooms.get(roomName)
	var i = 0
	for _, user := range room.users.users {
		comma := ""
		if i > 0 {
			comma = ","
		}
		users = users + comma + user.Name
	}
	w.Write([]byte(users))
}

// TODO: replace room message strings with timeMessages
type timeMessage struct {
	time    time.Time
	message string
}
