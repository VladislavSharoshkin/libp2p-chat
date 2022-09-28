package internal

import (
	"context"
	"encoding/json"
	"github.com/aead/ecdh"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"libp2p-chat/misk"
	"net/http"
	"time"
)

type Message struct {
	Text      string
	From      peer.ID
	To        peer.ID
	CreatedAt time.Time
}

func NewMessage(Text string, From peer.ID, To peer.ID, CreatedAt time.Time) Message {
	return Message{Text: Text, From: From, To: To, CreatedAt: CreatedAt}
}

func (mes *Message) Hashing() {

}

func EncryptMessage(mes Message, pr crypto.PrivKey) (Message, error) {
	pub, err := mes.To.ExtractPublicKey()
	if err != nil {
		return Message{}, err
	}

	c25519 := ecdh.X25519()
	secret := c25519.ComputeSecret(pr, pub)

}

func MessageSendHttp(client *http.Client, from peer.ID, to peer.ID, text string) (mes Message, err error) {
	mes = NewMessage(text, from, to, time.Now())
	err = misk.PostJson(client, mes.To.String()+"/message/new", mes)
	return mes, err
}

func MessageSend(topic *pubsub.Topic, from peer.ID, to peer.ID, text string) (mes Message, err error) {
	mes = NewMessage(text, from, to, time.Now())
	postBody, err := json.Marshal(mes)
	if err != nil {
		return mes, err
	}
	err = topic.Publish(context.Background(), postBody)
	if err != nil {
		return mes, err
	}

	return mes, err
}

func ProcessingMessage() {

}
