package chat

import (
	"encoding/json"
	"net/http"
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
	Channel chan Message
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

	history     []*Message
	HistorySize int
	mutex       sync.Mutex
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
		newId := c.Identity.Generate()
		ctx.SetCookie("id", newId, int(time.Duration.Hours(24*30)), "*", ctx.Request.URL.Host, true, true)
		ctx.JSON(http.StatusOK, gin.H{
			"msg": "ok!",
			"id":  newId,
		})
	}
}

func (c *Chat) Setup() {
	if c.Identity == nil {
		panic("chat criado sem coisinha de identidade")
	}
	c.active = make([]*User, 0, 10)
	c.history = make([]*Message, 0, 10)

	c.Engine.GET("/chat", c.midIdentity(), c.handleChat())
	c.Engine.GET("/id", c.handleIdentity())
}
