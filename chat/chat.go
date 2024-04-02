package chat

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type IID interface {
	Generate() string
	Check(string) (string, error)
}

type Message struct {
	From    string `json:"user_id"`
	Content string `json:"msg"`
}

func (m *Message) FromJSON(jsonStr string) error {
	return json.Unmarshal([]byte(jsonStr), m)
}

type User struct {
	Id      string
	Channel chan *Message
	Dead    bool

	lastMessage time.Time
	muted       bool
}

func (u *User) leave() {
	u.Dead = true
}

type Chat struct {
	Upgrader    websocket.Upgrader
	Engine      *gin.Engine
	Identity    IID
	active      []*User
	activeCount int

	//history     []*Message
	history     *History[*Message]
	HistorySize int
	mutex       sync.Mutex

	LogMessages bool
}

type GinHandler = gin.HandlerFunc

// A gente só adiciona um user, o user sendo filtrado da lista de users
// quando uma mensagem é enviada.
func (c *Chat) addUser(user *User) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.active = append(c.active, user)
}

// Envia um cookie de identidade.
func (c *Chat) handleIdentity() GinHandler {
	return func(ctx *gin.Context) {
		// antes daqui, a requisição passa pelo middleware.
		var id string
		const maxAge = 1000 * 60 * 60 * 24
		id, err := c.extractId(ctx)
		if err != nil {
			id = c.Identity.Generate()
			log.Println("gerando nova identidade...")
			ctx.SetCookie("id", id, maxAge, "*", ctx.Request.URL.Host, false, true)
		}
		ctx.JSON(http.StatusOK, gin.H{
			"msg":     "ok!",
			"id":      id,
			"expires": maxAge,
		})
	}
}

func (c *Chat) SaveHistory() {
	log.Println("salvando histórico...")
	history := c.history.Unroll()
	bytes, err := json.Marshal(history)
	if err != nil {
		log.Println("erro marshalizando histórico: ", err)
		return
	}
	err = os.WriteFile("history.json", bytes, 0644)
	if err != nil {
		log.Println("erro persistindo histórico: ", err)
	}
}

func (c *Chat) LoadHistory() {
	log.Println("carregando histórico...")
	bytes, err := os.ReadFile("history.json")
	if err != nil {
		log.Println("erro lendo histórico: ", err)
		return
	}
	messages := make([]Message, 0, c.history.size)
	err = json.Unmarshal(bytes, &messages)
	if err != nil {
		log.Println("erro desmarshalizando histórico: ", err)
		return
	}
	for i := range messages {
		c.history.Add(&messages[i])
	}
}

func (c *Chat) Setup() {
	if c.Identity == nil {
		panic("chat criado sem coisinha de identidade")
	}
	c.active = make([]*User, 0, 10)
	c.history = NewHistory[*Message](c.HistorySize)
	c.Engine.GET("/chat", c.midIdentity(), c.handleChat())
	c.Engine.GET("/id", c.handleIdentity())
}
