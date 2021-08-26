package main

import (
	"fmt"
	"log"
	"strconv"
)

func stats() {
	a := Admin{}
	sessions, err := a.SessionList()
	if err != nil {
		log.Println("admin err:", err)
		return
	}
	fmt.Println("current sess: " + strconv.Itoa(len(sessions)))
	streamCount := 0
	for _, sess := range sessions {
		handles, err := a.HandleList(sess)
		if err != nil {
			log.Println("admin err:", err)
			continue
		}
		for _, handle := range handles {
			handleInfo, err := a.HandleInfo(sess, handle)
			if err != nil {
				log.Println("admin err:", err)
				continue
			}
			inStatComponent := handleInfo.Get("streams.0.components.0.in_stats")
			value := inStatComponent.Get("video_packets")
			if value.Int() > 0 {
				streamCount++
			}
			valueNackIn := inStatComponent.Get("video_nacks")
			if valueNackIn.Int() > 0 {
				fmt.Println(valueNackIn.String())
			}
			outStatComponent := handleInfo.Get("streams.0.components.0.out_stats")
			valueNackOut := outStatComponent.Get("video_nacks")
			if valueNackOut.Int() > 0 {
				fmt.Println(valueNackOut.String())
			}
			value = outStatComponent.Get("video_packets")
			if value.Int() > 0 {
				streamCount++
			}
		}
	}
	fmt.Println("stream count: " + strconv.Itoa(streamCount))
}
