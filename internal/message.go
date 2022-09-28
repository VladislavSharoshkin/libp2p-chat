package internal

import (
	"context"
	"github.com/aead/ecdh"
	pubsub "github.com/libp2p/go-libp2p-pubsub"
	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
	"libp2p-chat/misk"
	"log"
	"time"
)

type Message struct {
	ID            string
	Text          string
	From          peer.ID
	To            peer.ID
	CreatedAt     time.Time
	EncryptedData string
}

func NewMessage(Text string, From peer.ID, To peer.ID, CreatedAt time.Time) Message {
	return Message{Text: Text, From: From, To: To, CreatedAt: CreatedAt}
}

func (mes *Message) hash() {
	mes.ID = misk.Sha([]byte(mes.EncryptedData))
}

func (mes *Message) encrypt(pr crypto.PrivKey) error {
	pub, err := mes.To.ExtractPublicKey()
	if err != nil {
		return err
	}

	c25519 := ecdh.X25519()
	prB := misk.ToCurve25519SK(pr)
	pubB := misk.ToCurve25519PK(pub)
	secret := c25519.ComputeSecret(prB, pubB)
	mes.EncryptedData = misk.AesEncrypt(mes.Text, secret)

	return nil
}

func (mes *Message) decrypt(pr crypto.PrivKey) error {
	pub, err := mes.From.ExtractPublicKey()
	if err != nil {
		return err
	}

	c25519 := ecdh.X25519()
	prB := misk.ToCurve25519SK(pr)
	pubB := misk.ToCurve25519PK(pub)
	secret := c25519.ComputeSecret(prB, pubB)
	mes.Text = misk.AesDecrypt(mes.EncryptedData, secret)

	return nil
}

//func MessageSendHttp(client *http.Client, from peer.ID, to peer.ID, text string) (mes Message, err error) {
//	mes = NewMessage(text, from, to, time.Now())
//	err = misk.PostJson(client, mes.To.String()+"/message/new", mes)
//	return mes, err
//}

func MessageSend(topic *pubsub.Topic, from peer.ID, to peer.ID, text string, pr crypto.PrivKey) (mes Message, err error) {

	mes = NewMessage(text, from, to, time.Now())
	err = mes.encrypt(pr)
	if err != nil {
		return mes, err
	}

	mes.Text = ""

	err = topic.Publish(context.Background(), misk.Marsh(mes))
	if err != nil {
		return mes, err
	}

	return mes, err
}

func ProcessMessage(mes Message, pr crypto.PrivKey) {
	err := mes.decrypt(pr)
	if err != nil {
		log.Println(err)
	}

	misk.PrintBlock(
		misk.ValueInfo("New message from", mes.From),
		misk.ValueInfo("To", mes.To),
		misk.ValueInfo("Date", mes.CreatedAt.Format("02.01.06 15:04")),
		misk.ValueInfo("Encrypted Data", mes.EncryptedData),
		misk.ValueInfo("Text", mes.Text),
	)
}
