package handlers

import (
	"fmt"
	"log"
	"net/http"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)

var clients = make(map[WsConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WsConnection struct {
	*websocket.Conn
}

// represents JSON response for WebSocket communication.
type WsJsonResponse struct {
	Action string `json:"action"`
	Message string `json:"message"`
	MessageType string `json:"messageType"`
}

type WsPayload struct {
	Action string `json:"action"`
	Message string `json:"message"`
	Username string `json:"username"`
	CurentConn WsConnection `json:"-"`
}

func WsEndpoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w,r, nil)
	if err != nil {
		log.Println(err)
		return 
	}
	log.Println("client connected to endpoint")

	var response WsJsonResponse
	response.Message = `<em><small>Welcome to the WebSocket endpoint!</small></em>`

	conn := WsConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(response)
	if err != nil {
		log.Println(err)
	} 
	go ListenForWs(&conn)
}

func ListenForWs(conn *WsConnection) {
	defer func(){ 
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))   
		}
	}()
		
	var payload WsPayload
	for {
		err := conn.ReadJSON(&payload)
		if err != nil {
			log.Println(err)
		} else {	
			payload.CurentConn = *conn
			wsChan <- payload
		}
	}
}
func ListenToWsChannel() {
	var response WsJsonResponse
	
	for {
		e := <-wsChan
		switch e.Action {
		case "username":
			// get a list of all users and send it back through broadcast
			broadcast(response)
		}


		
		// response.Action = "got here"
		// response.Message = fmt.Sprintf("some message, and action is %s", e.Action)
	}
}

func broadcast(response WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(response)
		if err != nil {
			log.Println("websocket write error", err)
			_ = client.Close()
			delete(clients, client)
		}
		
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error{
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil

}
