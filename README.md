# GoChat

Esse é um chatzinho com websocket em go.

Quando ele é executado, ele serve uns 3 endpoints.

```
GET /
```
Um frontend simples em HTML.


```
GET /id
```
Retorna uma mensagem e um Cookie com um token identificador.

O token tem 3 partes separadas por . , e a primeira é o id do usuário encodado em Base64.

(o ID do usuário é gerado)

```
GET /chat
```
Endpoint websocket. Espera receber um cookie id=(token retornado pelo /id)

Mensagens são enviadas em texto plano e recebidas em JSON no formato

```
{
    "user_id" : "quem enviou",
    "msg" : "o conteúdo da mensagem"
}
```

### Pra rodar:

go run .

## Deploy ao vivo e a cores

https://chat.r718.org/

1- Provavelmente não vai ter ninguém dentro dele.

2- Não estou mais rodando no fly.io, mas em um tv box, então talvez você não consiga ver o chat em ação, mas confia que ele funciona