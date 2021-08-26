package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

type Server struct {
	conn      *websocket.Conn
	sessID    string
	handleID  string
	offer     string
	list      []Stream
	wmux      sync.Mutex
	offerChan chan bool
	listChan  chan bool
	streamId  int
}

func (s *Server) Connect(wsURL string) {
	s.offerChan = make(chan bool, 1)
	s.listChan = make(chan bool, 1)
	websocket.DefaultDialer.Subprotocols = []string{"janus-protocol"}
	var resp *http.Response
	var err error
	s.conn, resp, err = websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		fmt.Println(err)
		log.Fatal("dial:", err, resp.StatusCode)
	}
	go func() {
		defer s.conn.Close()
		for {
			_, message, err := s.conn.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				continue
			}
			Debug("recv s: " + string(message))
			req := JanusMessage{}
			err = json.Unmarshal(message, &req)
			if err != nil {
				log.Printf("msg err: %s", err)
				continue
			}
			if req.Transaction == "list" {
				streams := []Stream{}
				if req.Plugindata != nil {
					list := req.Plugindata["data"].(map[string]interface{})["list"].([]interface{})
					for _, stream := range list {
						streamInfo := stream.(map[string]interface{})
						if streamInfo["video_age_ms"] != nil {
							streams = append(streams, Stream{
								Id:          int(streamInfo["id"].(float64)),
								Description: streamInfo["description"].(string),
								VideoAgeMs:  int(streamInfo["video_age_ms"].(float64)),
							})
						}
					}
				}
				s.list = streams
				s.listChan <- true
			}
			if req.Transaction == "1" { // Подключаемся к плагину стриминг
				s.sessID = fmt.Sprintf("%.f", req.Data["id"])
				err = s.send(`{"janus":"attach","plugin":"janus.plugin.streaming","opaque_id":"test1","transaction":"2","session_id":` + s.sessID + `}`)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			if req.Transaction == "2" { // Запускаем просмотр стрима
				s.handleID = fmt.Sprintf("%.f", req.Data["id"])
				err = s.send(`{"janus":"message","body":{"request":"watch","id":` + strconv.Itoa(s.streamId) + `},"transaction":"3","session_id":` + s.sessID + `,"handle_id":` + s.handleID + `}`)
				if err != nil {
					log.Println("write:", err)
					return
				}
			}
			if req.Jsep != nil {
				s.offer = req.Jsep["sdp"].(string)
				s.offerChan <- true
			}
			/*if req.Plugindata != nil {
				if req.Plugindata["data"].(map[string]interface{})["result"].(map[string]interface{})["status"] == "starting" {

				}
			}*/
		}
	}()
	go func() {
		// поддерживаем соединение
		for {
			if s.sessID != "" {
				s.send(`{"janus":"keepalive","transaction":"0","session_id":` + s.sessID + `,"handle_id":` + s.handleID + `}`)
			}
			time.Sleep(time.Second * 30)
		}
	}()
	err = s.send(`{"janus":"create","transaction":"1"}`)
	if err != nil {
		log.Println("write:", err)
		return
	}
}

func (s *Server) GetOffer() string {
	<-s.offerChan
	return s.offer
}

func (s *Server) SetStreamId(streamId int) {
	s.streamId = streamId
}

func (s *Server) GetList() []Stream {
	err := s.send(`{"janus":"message","body":{"request":"list"},"transaction":"list","session_id":` + s.sessID + `,"handle_id":` + s.handleID + `}`)
	if err != nil {
		log.Println("write:", err)
		return nil
	}
	<-s.listChan
	return s.list
}

func (s *Server) send(msg string) error {
	s.wmux.Lock()
	defer s.wmux.Unlock()
	Debug("send s: " + string(msg))
	return s.conn.WriteMessage(websocket.TextMessage, []byte(msg))
}

func (s *Server) SetAnswer(sdp string) {
	sdpByte, _ := json.Marshal(sdp)
	msg := `{"janus":"message","body":{"request":"start"},"transaction":"5","jsep":{"type":"answer","sdp":` + string(sdpByte) + `, "trickle":false},"session_id":` + s.sessID + `,"handle_id":` + s.handleID + `}`
	Debug("send s: " + msg + "\n")
	err := s.send(msg)
	if err != nil {
		log.Println("write:", err)
		return
	}
	{
		err := s.send(`{"janus":"trickle","candidate":{"completed":true},"transaction":"6","session_id":` + s.sessID + `,"handle_id":` + s.handleID + `}`)
		if err != nil {
			log.Println("write:", err)
			return
		}
	}
}
