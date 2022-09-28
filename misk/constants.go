package misk

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/fatih/color"
	"github.com/libp2p/go-libp2p/core/peer"
	"log"
	"net/http"
)

func PostJson(client *http.Client, url string, structData interface{}) error {
	postBody, err := json.Marshal(structData)
	if err != nil {
		return err
	}
	_, err = client.Post("libp2p://"+url, "application/json", bytes.NewBuffer(postBody))
	return err
}

func ColorMain(data string) string {
	return color.HiCyanString(data)
}

func ValueInfo(name string, value interface{}) string {
	return fmt.Sprint(ColorMain(name+": "), value)
}

func PrintBlock(data ...interface{}) {
	for _, value := range data {
		fmt.Println(value)
	}
	fmt.Println()
}

func PrintError(err error) {
	if err != nil {
		log.Println(err)
	}
}

func ClearConsole() {
	fmt.Print("\033[H\033[2J")
}

func LoadPeerID(data string) peer.ID {
	id, _ := peer.Decode(data)
	return id
}
