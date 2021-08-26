package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	serv     *Server
	conn     *websocket.Conn
	sessID   string
	handleID string
	wmux     sync.Mutex
}

func (c *Client) Connect(wsURL string) {
	websocket.DefaultDialer.Subprotocols = []string{"janus-protocol"}
	var resp *http.Response
	var err error
	c.conn, resp, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println(err)
		log.Fatal("dial:", err, resp.StatusCode)
	}
	go func() {
		defer c.conn.Close()
		for {
			msg := ""
			_, message, err := c.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				continue
			}
			Debug("recv c:" + string(message))
			req := JanusMessage{}
			err = json.Unmarshal(message, &req)
			if err != nil {
				log.Printf("msg err: %s", err)
				continue
			}
			if req.Transaction == "1" { // Подключаемся к videoroom плагину
				c.sessID = fmt.Sprintf("%.f", req.Data["id"])
				msg = `{"janus":"attach","plugin":"janus.plugin.videoroom","opaque_id":"client1","transaction":"2","session_id":` + c.sessID + `}`
				Debug("send: " + msg + "\n")
				err = c.send(msg)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			if req.Transaction == "2" { // Входим в комнату и создаем приемника потока с сервера
				c.handleID = fmt.Sprintf("%.f", req.Data["id"])
				msg = `{"janus":"message","body":{"request":"join","room":1234, "ptype":"publisher", "display":"client1"},"transaction":"3","session_id":` + c.sessID + `,"handle_id":` + c.handleID + `}`
				Debug("send: " + msg + "\n")
				err = c.send(msg)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			if req.Jsep != nil {
				c.serv.SetAnswer(req.Jsep["sdp"].(string))
				/*{
					err := c.send(`{"janus":"trickle","candidate":{"completed":true},"transaction":"6","session_id":`+c.sessID+`,"handle_id":`+c.handleID+`}`)
					if err != nil {
						log.Println("write:", err)
						return
					}
				}*/

			}
			if req.Plugindata != nil { // Конфигурим соединение (куда подключаться?)
				if req.Plugindata["data"].(map[string]interface{})["videoroom"] == "joined" {
					//time.Sleep(time.Second)
					sdp, _ := json.Marshal(c.serv.GetOffer())
					msg = `{"janus":"message","body":{"request":"configure","audio":true, "video":true},"jsep":{"type":"offer","sdp":` + string(sdp) + `, "trickle":false, "ice-restart":false},"transaction":"4","session_id":` + c.sessID + `,"handle_id":` + c.handleID + `}`
					Debug("send: " + msg + "\n")
					err = c.send(msg)
					if err != nil {
						log.Println("write:", err)
						return
					}
				}
			}
		}
	}()
	go func() {
		// поддерживаем соединение
		for {
			c.send(`{"janus":"keepalive","transaction":"0","session_id":` + c.sessID + `,"handle_id":` + c.handleID + `}`)
			time.Sleep(time.Second * 30)
		}
	}()
	msg := `{"janus":"create","transaction":"1"}`
	Debug("send: " + msg + "\n")
	err = c.send(msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (c *Client) SetServer(serv *Server) {
	c.serv = serv
}

func (c *Client) send(msg string) error {
	c.wmux.Lock()
	defer c.wmux.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}
