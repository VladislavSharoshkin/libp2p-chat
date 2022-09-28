package internal

import "libp2p-chat/misk"

type MessageKey struct {
	Key []byte
	Iv  []byte
}

func NewMessageKey(Key []byte, Iv []byte) MessageKey {
	return MessageKey{Key: Key, Iv: Iv}
}

func GenerateKey() MessageKey {
	key := misk.RandomBytes(32)
	iv := misk.RandomBytes(32)
	return NewMessageKey(key, iv)
}
