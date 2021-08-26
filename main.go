package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"
)

var debug = false
var testServerWsUrl = "ws://192.168.1.5:8188"      // WS клиента тестируемого сервера Janus
var clientWsUrl = "ws://192.168.1.100:8188"        // WS клиента сервера Janus с которого тестируют
var AdminHTTPURL = "http://192.168.1.5:7089/admin" // without '/' http api админки Janus тестируемого сервера
var adminSecret = "c9r9q8n439f8q3ufnq3"
var streamContaintsName = "1080"

func Debug(msg string) {
	if !debug {
		return
	}
	fmt.Println(msg)
}

type Message struct {
	C int                    `json:"c"`
	A string                 `json:"a"`
	D map[string]interface{} `json:"d"`
}

type JanusMessage struct {
	Transaction string                 `json:"transaction"`
	Data        map[string]interface{} `json:"data"`
	Jsep        map[string]interface{} `json:"jsep"`
	Janus       string                 `json:"janus"`
	Plugindata  map[string]interface{} `json:"plugindata"`
}

type Stream struct {
	Id          int
	Description string
	VideoAgeMs  int
}

func main() {
	fmt.Println("Start janus_stress_test")
	// Коннект к ws тестируемого сервера Janus
	server := Server{}
	server.Connect(testServerWsUrl)
	// TODO Ждем завершения коннекта, костыль, переделать на событие
	time.Sleep(time.Second)
	// Получение списка стримов с сервера
	list := server.GetList()
	useStreams := []Stream{}
	// Поиск стримов в названии которых есть 1080 и по которым идет видео трафик
	for _, stream := range list {
		if strings.Index(stream.Description, streamContaintsName) != -1 && stream.VideoAgeMs < 500 {
			useStreams = append(useStreams, stream)
		}
	}
	rand.Seed(42)
	// Запускаем 5 соединений на просмотр стримов
	for i := 0; i < 5; i++ {
		s := Server{}
		s.SetStreamId(useStreams[rand.Intn(len(useStreams))].Id)
		s.Connect(testServerWsUrl)
		c := Client{}
		c.SetServer(&s)
		c.Connect(clientWsUrl)
	}
	// Выводим статистику каждые 3 секунды
	go func() {
		for {
			time.Sleep(time.Second * 3)
			stats()
		}
	}()

	// Ждем минуту для завершения
	time.Sleep(time.Second * 600)
}
