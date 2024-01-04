package main

import (
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"testing"
	"time"

	"github.com/Alfrederson/gochat/chat"
	"github.com/Alfrederson/gochat/identity"
	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type TestRig struct {
	engine   *gin.Engine
	server   *httptest.Server
	recorder *httptest.ResponseRecorder
	chat     *chat.Chat
}

func (t *TestRig) shutdown() {
	t.server.Close()
}

func (r *TestRig) req(method string, path string, t *testing.T, cookies ...*http.Cookie) *http.Response {
	req, err := http.NewRequest(method, path, nil)
	if err != nil {
		t.Errorf("erro em %s %s : %v", method, path, err)
	}

	for _, cookie := range cookies {
		req.AddCookie(cookie)
	}

	r.engine.ServeHTTP(r.recorder, req)
	return r.recorder.Result()
}

func (r *TestRig) ws(path string, t *testing.T, cookies ...*http.Cookie) *websocket.Conn {
	wsUrl := "ws" + r.server.URL[4:] + path // isso dá tipo ws://127.0.0.1:xxxx/(path)
	dialer := websocket.DefaultDialer

	t.Logf("ws: %s", wsUrl)

	dialer.Jar, _ = cookiejar.New(nil)
	u, _ := url.Parse(r.server.URL)
	dialer.Jar.SetCookies(u, cookies)

	cookieString := ""
	cookieCount := len(cookies)
	for i, cookie := range cookies {
		cookieString += cookie.String()
		if i < cookieCount-1 {
			cookieString += " "
		}
	}
	header := http.Header{}
	header.Add("Cookie", cookieString)
	ws, _, err := dialer.Dial(wsUrl, header)
	if err != nil {
		t.Fatalf("erro abrindo websocket %v", err)
	}
	return ws
}

func makeTestRig() *TestRig {
	rig := TestRig{}

	rig.engine = gin.Default()
	rig.chat = &chat.Chat{
		Upgrader: websocket.Upgrader{
			CheckOrigin: func(r *http.Request) bool {
				return true
			},
		},
		Engine: rig.engine,
		Identity: &identity.Id{
			Signature: "qualquer coisa aqui",
		},
		HistorySize: 5,
	}
	rig.chat.Setup()

	rig.server = httptest.NewServer(rig.engine)
	rig.recorder = httptest.NewRecorder()

	return &rig
}

func getId(rig *TestRig, t *testing.T) string {
	result := rig.req("GET", "/id", t)
	if result.StatusCode != http.StatusOK {
		t.Fatalf("o status deveria ser %d, mas foi %d ", http.StatusOK, result.StatusCode)
	}
	cookies := result.Cookies()
	if len(cookies) != 1 {
		t.Errorf("cade os cookies? esperava 2, achei %d", len(cookies))
		return ""
	}
	if cookies[0].Name != "id" {
		t.Errorf("cade o id no cookie? cookie = %s", cookies[0].String())
		return ""
	}
	return cookies[0].Value
}

// testa se requisição em /id dá um id.
func TestId(t *testing.T) {
	t.Logf("testando se pedir o id dá um cookie com o id...")
	rig := makeTestRig()
	defer rig.shutdown()

	_ = getId(rig, t) // ignora o valor.
}

// testa se não consegue entrar no chat sem o cookie.
func TestChatJoinNoCookie(t *testing.T) {
	t.Logf("testando se entrar sem o cookie de id dá ruim...")
	rig := makeTestRig()
	defer rig.shutdown()

	result := rig.req("GET", "/chat", t)

	if result.StatusCode != http.StatusUnavailableForLegalReasons {
		t.Errorf("entrar no chat sem cookie deveria dar um erro %d. recebi um %d", http.StatusUnavailableForLegalReasons, result.StatusCode)
	}
}

// testa se não consegue entrar no chat com um cookie inválido.
func TestChatJoinFakeCookie(t *testing.T) {
	t.Logf("testando se entrar com cookie falso dá ruim...")
	rig := makeTestRig()
	defer rig.shutdown()

	result := rig.req("GET", "/chat", t,
		&http.Cookie{Name: "id", Value: "COOKIE FORJADO"},
	)

	if result.StatusCode != http.StatusUnavailableForLegalReasons {
		t.Errorf("entrar no chat com cookie falso deveria dar um erro %d. recebi um %d", http.StatusUnavailableForLegalReasons, result.StatusCode)
	}
}

// testa se consegue entrar no chat com um cookie válido.
func TestChatJoin(t *testing.T) {
	t.Logf("testando se entrar no chat com cookie verdadeiro dá bom...")
	rig := makeTestRig()
	defer rig.shutdown()
	id := getId(rig, t)

	result := rig.req("GET", "/chat", t,
		&http.Cookie{Name: "id", Value: id},
	)
	if result.StatusCode != http.StatusOK {
		t.Errorf("entrar no chat com cookie bom deveria dar bom. recebi status %d ", result.StatusCode)
	}
}

func TestChatEcho(t *testing.T) {
	// 1- faz a requisição no id pra pedir o cookie
	// 2- manda uma mensagem
	// 3- espera receber ela de volta
	rig := makeTestRig()
	defer rig.shutdown()

	// pego o id
	id := getId(rig, t)

	t.Logf("id: %s", id)
	// crio um websocket pra entrar no chat.
	ws := rig.ws("/chat", t,
		&http.Cookie{Name: "id", Value: id},
	)
	defer ws.Close()

	msg := chat.Message{}
	var recebida []byte
	enviada := ""

	// a gente deve receber uma mensagem.
	_, recebida, err := ws.ReadMessage()
	if err != nil {
		t.Fatalf("erro lendo resposta: %v", err)
	}

	err = msg.FromJSON(string(recebida))
	if err != nil {
		t.Errorf("A mensagem de boas vindas deveria ter vindo como json. Ela foi: %s", string(recebida))
	}
	if msg.From != "😀" {
		t.Errorf("A mensagem de boas vindas deveria ser de 😀, mas era de %s", msg.From)
	}

	enviada = "eu sou a flor silvestre que perfuma os campos"

	// tenho que esperar 1 segundo antes de enviar, senão vou ser SEDADO.
	time.Sleep(time.Millisecond * 1100)

	// mando uma mensagem!
	if err := ws.WriteMessage(websocket.TextMessage, []byte(enviada)); err != nil {
		t.Fatalf("erro mandando mensagem: %v", err)
	}

	// espero receber ela de volta
	_, recebida, err = ws.ReadMessage()
	if err != nil {
		t.Fatalf("erro lendo resposta: %v", err)
	}
	err = msg.FromJSON(string(recebida))
	if err != nil {
		t.Fatalf("o eco deveria ser parseável como json. recebi: %s", string(recebida))
	}

	if msg.Content != enviada {
		t.Fatalf("o eco deveria ser igual ao que enviei. \nenviei: [%s]\n recebi: [%s]", enviada, msg.Content)
	}

}

// and that's basically it!
//
// isso é uma base.
// a natureza da comunicação em um chat torna tudo razoavelmente complicado de testar.
// estratégia pra testes:
// - fixar um baseline de funcionamento com os benditos testes
// - identificar bugs
// - reproduzir o bug como uma função de teste
// - consertar o bug
// - deixar o teste lá
