package main

import (
	"fmt"
	"net/http"
	"path/filepath"

	"github.com/Alfrederson/gochat/chat"
	"github.com/Alfrederson/gochat/identity"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

func main() {

	r := gin.Default()

	static := filepath.Join(".", "www")
	r.NoRoute(
		func(ctx *gin.Context) {
			http.ServeFile(ctx.Writer, ctx.Request, filepath.Join(static, ctx.Request.URL.Path))
		},
	)

	port := 8080

	c := chat.Chat{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Engine: r,
		Identity: &identity.Id{
			// NOTA MENTAL: se você estiver copiando e colando isso, certifique-se de não usar esse carimbo (tirar ele do environment ou do céu, não sei)
			Signature: "a soma de todos os medos",
		},
		HistorySize: 5,
	}

	c.Setup()

	fmt.Printf("Tentando escutar na porta %d ... \n", port)
	if err := r.Run(fmt.Sprintf("0.0.0.0:%d", port)); err != nil {
		fmt.Println("Deu ruim: ", err)
	}
}
