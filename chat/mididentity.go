package chat

import (
	"log"
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

// middleware que decodifica a identidade e põe dentro do contexto.
func (c *Chat) midIdentity() GinHandler {
	return func(ctx *gin.Context) {
		// cookie com a identidade
		// se não tem a identidade, cai fora pois está tentando entrar ilegalmente
		// usamos o status code 451 para representar que a ação é ILEGAL
		cookie, err := ctx.Request.Cookie("id")
		if err != nil {
			log.Println(err)
			ctx.AbortWithStatusJSON(http.StatusUnavailableForLegalReasons, gin.H{"err": "no cookie sir"})
			return
		}
		// decodifica o cookie
		cookieValue, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			log.Println(err)
			ctx.AbortWithStatusJSON(http.StatusUnavailableForLegalReasons, gin.H{"err": "no idea sir"})
			return
		}
		// checa a assinatura
		id, err := c.Identity.Check(cookieValue)
		if err != nil {
			log.Println(err)
			ctx.AbortWithStatusJSON(http.StatusUnavailableForLegalReasons, gin.H{"err": "signature is wrong sir"})
			return
		}
		// toma lá teu user
		ctx.Set("id", &User{
			Id: id,
		})
		ctx.Next()
	}
}
