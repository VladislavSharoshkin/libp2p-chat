package controllers

import (
	"encoding/json"
	"libp2p-chat/internal"
	"libp2p-chat/misk"
	"log"
	"net/http"
)

func NewMessage(w http.ResponseWriter, r *http.Request) {
	message := internal.Message{}
	err := json.NewDecoder(r.Body).Decode(&message)
	if err != nil {
		log.Println(err)
	}

	misk.PrintBlock(
		misk.ValueInfo("New message from", message.From),
		misk.ValueInfo("Text", message.Text),
		misk.ValueInfo("Date", message.CreatedAt.Format("02.01.06 15:04")),
	)
}
