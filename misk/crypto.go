package misk

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	Ed "crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"crypto/sha512"
	"encoding/base64"
	"fmt"
	"github.com/libp2p/go-libp2p/core/crypto"
	"io"
	"math/big"
)

func ToBase64(data []byte) string {
	return base64.StdEncoding.EncodeToString(data)
}

func FromBase64(data string) []byte {
	decodeString, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		return nil
	}
	return decodeString
}

func RandomBytes(len int) []byte {
	data := make([]byte, len)
	rand.Read(data)
	return data
}

func RandomBytes256() []byte {
	return RandomBytes(32)
}

func AesEncrypt(plaintext string, key []byte) string {

	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		fmt.Println(err)
	}

	enc := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return ToBase64(enc)
}

func AesDecrypt(plaintext2 string, key []byte) string {

	ciphertext := FromBase64(plaintext2)

	c, err := aes.NewCipher(key)
	if err != nil {
		fmt.Println(err)
	}

	gcm, err := cipher.NewGCM(c)
	if err != nil {
		fmt.Println(err)
	}

	nonceSize := gcm.NonceSize()
	if len(ciphertext) < nonceSize {
		fmt.Println(err)
	}

	nonce, ciphertext := ciphertext[:nonceSize], ciphertext[nonceSize:]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		fmt.Println(err)
	}
	return string(plaintext)
}

func PKCS5Padding(ciphertext []byte, blockSize int, after int) []byte {
	padding := (blockSize - len(ciphertext)%blockSize)
	padtext := bytes.Repeat([]byte{byte(padding)}, padding)
	return append(ciphertext, padtext...)
}

func clamp(k []byte) []byte {
	k[0] &= 248
	k[31] &= 127
	k[31] |= 64
	return k
}

func ToCurve25519SK(sk crypto.PrivKey) []byte {
	skB, err := sk.Raw()
	if err != nil {
		return nil
	}

	var ek [64]byte

	h := sha512.New()
	h.Write(skB[:32])
	h.Sum(ek[:0])

	return clamp(ek[:32])
}

var curve25519P, _ = new(big.Int).SetString("57896044618658097711785492504343953926634992332820282019728792003956564819949", 10)

func ToCurve25519PK(pk crypto.PubKey) []byte {
	pkB, err := pk.Raw()
	if err != nil {
		return nil
	}

	// ed25519.PublicKey is a little endian representation of the y-coordinate,
	// with the most significant bit set based on the sign of the x-ccordinate.
	bigEndianY := make([]byte, Ed.PublicKeySize)
	for i, b := range pkB {
		bigEndianY[Ed.PublicKeySize-i-1] = b
	}
	bigEndianY[0] &= 0b0111_1111

	// The Montgomery u-coordinate is derived through the bilinear map
	//
	//     u = (1 + y) / (1 - y)
	//
	// See https://blog.filippo.io/using-ed25519-keys-for-encryption.
	y := new(big.Int).SetBytes(bigEndianY)
	denom := big.NewInt(1)
	denom.ModInverse(denom.Sub(denom, y), curve25519P) // 1 / (1 - y)
	u := y.Mul(y.Add(y, big.NewInt(1)), denom)
	u.Mod(u, curve25519P)

	out := make([]byte, 32)
	uBytes := u.Bytes()
	n := len(uBytes)
	for i, b := range uBytes {
		out[n-i-1] = b
	}

	return out
}

func Sha(data ...[]byte) string {
	h := sha256.New()
	for _, d := range data {
		h.Write(d)
	}
	return ToBase64(h.Sum(nil))
}
