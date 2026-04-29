package app

import (
	"log"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
)

type Client struct {
	Conn *websocket.Conn
	URL  string
	Done chan struct{}
}

// Khởi tạo client
func NewClient(rawURL string) *Client {
	return &Client{
		URL:  rawURL,
		Done: make(chan struct{}),
	}
}

// Connect tới server
func (c *Client) Connect() error {
	u, err := url.Parse(c.URL)
	if err != nil {
		return err
	}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		return err
	}

	c.Conn = conn
	log.Println("connected:", u.String())

	return nil
}

// Start đọc message
func (c *Client) Start() {
	go c.readLoop()
}

// Gửi message
func (c *Client) Send(msg []byte) error {
	return c.Conn.WriteMessage(websocket.TextMessage, msg)
}

// Loop đọc
func (c *Client) readLoop() {
	defer close(c.Done)

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			log.Println("read error:", err)
			return
		}

		log.Printf("recv: %s\n", message)
	}
}

// Đóng connection
func (c *Client) Close() {
	if c.Conn != nil {
		c.Conn.WriteMessage(
			websocket.CloseMessage,
			websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""),
		)
		c.Conn.Close()
	}
}

// Reconnect đơn giản
func (c *Client) Reconnect() {
	for {
		log.Println("reconnecting...")
		err := c.Connect()
		if err == nil {
			c.Start()
			return
		}
		time.Sleep(2 * time.Second)
	}
}