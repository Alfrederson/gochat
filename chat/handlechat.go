package chat

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// Faz o chat tudo aqui.
func (c *Chat) handleChat() GinHandler {
	// retorna uma closure que tem acesso às propriedades do
	// strut que contem o estado do chat.
	return func(ctx *gin.Context) {
		// coloca essa sessão nos nossos ouvintes.
		conn, err := c.Upgrader.Upgrade(
			ctx.Writer, ctx.Request, nil,
		)
		if err != nil {
			log.Println(err)
			ctx.String(http.StatusInternalServerError, "ups")
			return
		}

		// eu sei que tem um user aqui porque:
		// 1- ou o middleware botou ele aqui.
		// 2- ou o middleware não encontrou o cookie e não
		//    chegou nessa função.
		user := ctx.MustGet("id").(*User)
		user.Channel = make(chan Message)
		user.lastMessage = time.Now()
		c.addUser(user)
		defer conn.Close()

		// pega todas as mensagens que tiver no canal do usuário e
		// manda elas pra ele.
		go func() {
			for {
				msg, ok := <-user.Channel
				if !ok {
					return
				}
				conn.WriteJSON(msg)
			}
		}()

		// envia as mensagens velhas...
		for _, message := range c.history {
			conn.WriteJSON(message)
		}

		conn.WriteJSON(Message{
			From:    "😀",
			Content: fmt.Sprintf("Bem vindx, %s", user.Id),
		})

		for {
			// lê todas as mensagens que esse cara enviou
			_, message, err := conn.ReadMessage()
			// se não tem mais nada, it's over.
			if err != nil {
				c.broadcast(Message{
					From:    "😔",
					Content: user.Id + " saiu",
				})
				user.leave()
				log.Println(user.Id + " saiu ")
				break
			}

			now := time.Now()
			if time.Since(user.lastMessage) <= time.Second {
				if !user.muted {
					c.broadcast(Message{
						From:    "😲",
						Content: user.Id + " está surtando. Ativando SEDAÇÃO 💉",
					})
				}
				user.muted = true
			} else {
				user.muted = false
				c.broadcast(Message{
					From:    user.Id,
					Content: string(message),
				})
			}
			user.lastMessage = now
		}
	}
}
