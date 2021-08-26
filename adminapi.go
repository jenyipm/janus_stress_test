package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/tidwall/gjson"
)

type Admin struct {
	conn *websocket.Conn
}

func (a *Admin) Execute(action string, command string) ([]byte, error) {
	jsonStr := []byte(`{"janus": "` + command + `", "transaction": "1", "admin_secret": adminSecret}`)
	resp, err := http.Post(AdminHTTPURL+action, "application/json", bytes.NewBuffer(jsonStr))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := ioutil.ReadAll(resp.Body)
	return body, nil
}

func (a *Admin) SessionList() ([]int, error) {
	resp, err := a.Execute("", "list_sessions")
	if err != nil {
		return nil, err
	}
	req := struct {
		Janus    string `json:"janus"`
		Sessions []int  `json:"sessions"`
	}{}
	err = json.Unmarshal(resp, &req)
	if err != nil {
		return nil, err
	}
	return req.Sessions, nil
}

func (a *Admin) HandleList(session int) ([]int, error) {
	resp, err := a.Execute("/"+strconv.Itoa(session), "list_handles")
	if err != nil {
		return nil, err
	}
	req := struct {
		Janus   string `json:"janus"`
		Handles []int  `json:"handles"`
	}{}
	err = json.Unmarshal(resp, &req)
	if err != nil {
		return nil, err
	}
	return req.Handles, nil
}

func (a *Admin) HandleInfo(session int, handle int) (gjson.Result, error) {
	resp, err := a.Execute("/"+strconv.Itoa(session)+"/"+strconv.Itoa(handle), "handle_info")
	if err != nil {
		return gjson.Result{}, err
	}
	req := struct {
		Janus string          `json:"janus"`
		Info  json.RawMessage `json:"info"`
	}{}
	err = json.Unmarshal(resp, &req)
	if err != nil {
		return gjson.Result{}, err
	}
	return gjson.Parse(string(req.Info)), nil
}
