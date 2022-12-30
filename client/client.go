package client

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	serverAddress string
	httpClient    *http.Client
}

func New(serverAddress string) *Client {
	return &Client{
		serverAddress: serverAddress,
		httpClient:    &http.Client{},
	}
}

func (c *Client) GetRooms() ([]string, error) {
	url := fmt.Sprintf("%s/room/names", c.serverAddress)
	req, err := http.NewRequest(http.MethodGet, url, http.NoBody)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	rooms := strings.Split(string(data), ",")
	return rooms, nil
}

func (c *Client) GetRoomUsers(roomName string) ([]string, error) {
	url := fmt.Sprintf("%s/room/users", c.serverAddress)
	buf := bytes.NewBufferString(roomName)
	req, err := http.NewRequest(http.MethodGet, url, buf)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	users := strings.Split(string(data), ",")
	return users, nil
}
