package chat

import (
	"errors"
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

func (c *Chat) extractId(ctx *gin.Context) (string, error) {
	cookie, err := ctx.Request.Cookie("id")
	if err != nil {
		log.Println("erro lendo o cookie da requisição:", err)
		return "", errors.New("no cookie")
	}
	cookieValue, err := url.QueryUnescape(cookie.Value)
	if err != nil {
		log.Println("erro decodificando o cookie:", err)
		return "", errors.New("no idea sir")
	}
	id, err := c.Identity.Check(cookieValue)
	if err != nil {
		log.Println("erro checando assinatura:", err)
		return "", errors.New("signature is wrong sir")
	}
	return id, nil
}

// middleware que decodifica a identidade e põe dentro do contexto.
func (c *Chat) midIdentity() GinHandler {
	return func(ctx *gin.Context) {
		// cookie com a identidade
		// se não tem a identidade, cai fora pois está tentando entrar ilegalmente
		// usamos o status code 451 para representar que a ação é ILEGAL
		id, err := c.extractId(ctx)
		if err != nil {
			ctx.AbortWithStatusJSON(http.StatusUnavailableForLegalReasons, gin.H{"err": err.Error()})
			return
		}
		// toma lá teu user
		ctx.Set("id", &User{
			Id: id,
		})
		ctx.Next()
	}
}
