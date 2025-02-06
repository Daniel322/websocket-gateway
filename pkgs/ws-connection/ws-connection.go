package WsConnection

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type WsConnection struct {
	ConnectionId     string                     `json:"connectionId"`
	CreateConnection func(c echo.Context) error `json:"createConnection"`
	Websocket        *websocket.Conn            `json:"websocket"`
}

var (
	upgrader = websocket.Upgrader{
		ReadBufferSize:  1024, // Размер буфера чтения
		WriteBufferSize: 1024, // Размер буфера записи
		// Позволяет определить, должен ли сервер сжимать сообщения
		EnableCompression: true,
	}
)

// setInterval имитирует поведение setInterval из JavaScript
func setInterval(callback func(), interval time.Duration) chan bool {
	ticker := time.NewTicker(interval) // Создаём тикер с заданным интервалом
	stop := make(chan bool)            // Канал для остановки интервала

	go func() {
		for {
			select {
			case <-ticker.C: // Срабатывает каждые `interval` времени
				callback() // Выполняем переданную функцию
			case <-stop: // Если получен сигнал остановки
				ticker.Stop() // Останавливаем тикер
				return        // Выходим из горутины
			}
		}
	}()

	return stop // Возвращаем канал для остановки
}

var connection WsConnection = WsConnection{
	CreateConnection: func(c echo.Context) error {
		fmt.Println("here start to make ws")
		keepalive, err := strconv.Atoi(os.Getenv("KEEPALIVE_TIME"))
		if err != nil {
			c.Logger().Error(err)
		}
		ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
		stop := setInterval(func() {
			fmt.Println("interval cb")
			ws.WriteMessage(websocket.TextMessage, []byte("keepalive"))
		}, time.Duration(keepalive)*time.Second)
		if err != nil {
			stop <- true
			c.Logger().Error(err)
		}
		closeHandler := ws.CloseHandler()
		ws.SetCloseHandler(func(code int, text string) error {
			stop <- true
			fmt.Printf("Connection closed with code %d and text: %s\n", code, text)
			err = closeHandler(code, text)
			return err
		})
		defer ws.Close()

		for {
			// Write
			err := ws.WriteMessage(websocket.TextMessage, []byte("Hello, Client!"))
			if err != nil {
				c.Logger().Error(err)
				break
			}

			// Read
			// var msgData SocketMessage
			// err = ws.ReadJSON(&msgData)
			_, msg, err := ws.ReadMessage()
			if err != nil {
				stop <- true
				c.Logger().Error(err)
				break
			}
			fmt.Printf("%s\n", msg)
		}

		return err
	},
}
