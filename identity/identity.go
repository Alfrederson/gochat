package identity

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"math/rand"
	"strings"
	"time"
)

type Id struct {
	Signature string
}

var randSource = rand.New(rand.NewSource(time.Now().UnixNano()))

func algum_de(lista []string) string {
	return lista[randSource.Intn(len(lista))]
}

// gera uma string de letras aleatórias de até length caracteres
func salt(length int) string {
	const letras = "pao de queijo e um trem bao dimais da conta so"

	letrasLen := len(letras)
	result := make([]byte, length)
	for i := 0; i < length; i++ {
		result[i] = letras[randSource.Intn(letrasLen)]
	}

	return string(result)
}

// assinar uma string.
func (i *Id) Sign(what string) string {
	key := []byte(i.Signature)
	h := hmac.New(sha256.New, key)
	h.Write([]byte(what))
	signature := hex.EncodeToString(h.Sum(nil))
	return signature
}

// checar a assinatura de uma string.
// uma string assinada = b64(payload).b64(salt).b64(signature)
func (i *Id) Check(what string) (string, error) {
	parts := strings.Split(what, ".")
	if len(parts) != 3 {
		return "", errors.New("token must have 3 parts")
	}
	pay, err := i.Decode(parts[0])
	if err != nil {
		return "", errors.New("error decoding payload from base64: " + err.Error())
	}
	salt, err := i.Decode(parts[1])
	if err != nil {
		return "", errors.New("error decoding salt from base64: " + err.Error())
	}
	signature, err := i.Decode(parts[2])
	if err != nil {
		return "", errors.New("error decoding signature from base64: " + err.Error())
	}
	signed := i.Sign(pay + salt)
	if signed == signature {
		return pay, nil
	}
	return "", errors.New("signature doesn't match")
}

// encoda uma string em base64.
func (i *Id) Encode(whatId string) string {
	return base64.StdEncoding.EncodeToString([]byte(whatId))
}

// decoda uma string de base64.
func (i *Id) Decode(whatId string) (string, error) {
	if whatId == "" {
		return "", errors.New("no string")
	}
	decoded, err := base64.StdEncoding.DecodeString(whatId)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func (i *Id) Generate() string {
	id := "u-" + algum_de(animais) + "-" + algum_de(adjetivos)
	salt := salt(16)
	signature := i.Sign(id + salt)
	return i.Encode(id) + "." + i.Encode(salt) + "." + i.Encode(signature)
}
