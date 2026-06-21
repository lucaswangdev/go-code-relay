package main

import (
	"bufio"
	"fmt"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"syscall"

	"github.com/gorilla/websocket"
)

var version = "dev"

type Message struct {
	Type     string   `json:"type"`
	Content  string   `json:"content,omitempty"`
	Error    string   `json:"error,omitempty"`
	Modified []string `json:"modified,omitempty"`
	Deleted  []string `json:"deleted,omitempty"`
}

type RelayClient struct {
	serverURL string
	conn     *websocket.Conn
	done     chan struct{}
	respReady chan struct{}
}

func NewRelayClient(serverURL string) *RelayClient {
	return &RelayClient{
		serverURL:  serverURL,
		done:      make(chan struct{}),
		respReady: make(chan struct{}),
	}
}

func (c *RelayClient) Connect() error {
	var err error
	c.conn, _, err = websocket.DefaultDialer.Dial(c.serverURL, nil)
	return err
}

func (c *RelayClient) Close() {
	close(c.done)
	if c.conn != nil {
		c.conn.Close()
	}
}

func (c *RelayClient) SendMessage(content string) error {
	msg := Message{Type: "message", Content: content}
	return c.conn.WriteJSON(msg)
}

func (c *RelayClient) Clear() error {
	return c.conn.WriteJSON(Message{Type: "clear"})
}

func (c *RelayClient) GetDiff() ([]string, []string, error) {
	if err := c.conn.WriteJSON(Message{Type: "diff"}); err != nil {
		return nil, nil, err
	}
	var msg Message
	if err := c.conn.ReadJSON(&msg); err != nil {
		return nil, nil, err
	}
	return msg.Modified, msg.Deleted, nil
}

func (c *RelayClient) WaitForDone() error {
	<-c.respReady
	return nil
}

func (c *RelayClient) ReceiveLoop() {
	for {
		select {
		case <-c.done:
			return
		default:
		}

		var msg Message
		if err := c.conn.ReadJSON(&msg); err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseNormalClosure) {
				fmt.Printf("\n错误: 读取消息失败: %v\n", err)
			}
			return
		}

		switch msg.Type {
		case "chunk":
			fmt.Print(msg.Content)
		case "done":
			select {
			case c.respReady <- struct{}{}:
			default:
			}
		case "cleared", "pong":
		case "error":
			fmt.Printf("\n错误: %s\n", msg.Error)
			select {
			case c.respReady <- struct{}{}:
			default:
			}
		}
	}
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "--version" {
		fmt.Println("go-code-relay " + version)
		return
	}

	serverURL := os.Getenv("RELAY_SERVER_URL")
	if serverURL == "" {
		if len(os.Args) > 1 {
			serverURL = os.Args[1]
		} else {
			fmt.Println("错误: 请提供服务器地址")
			fmt.Println("用法: go-code-relay ws://服务器地址:端口")
			fmt.Println("或设置环境变量: export RELAY_SERVER_URL=ws://...")
			os.Exit(1)
		}
	}

	if _, err := url.Parse(serverURL); err != nil {
		fmt.Printf("错误: 无效的服务器地址: %v\n", err)
		os.Exit(1)
	}

	client := NewRelayClient(serverURL)

	if err := client.Connect(); err != nil {
		fmt.Printf("错误: 连接服务器失败: %v\n", err)
		os.Exit(1)
	}
	defer client.Close()

	fmt.Println("=" + strings.Repeat("=", 49))
	fmt.Println("Claude Code Relay")
	fmt.Println("=" + strings.Repeat("=", 49))
	fmt.Println("服务器:", serverURL)
	fmt.Println("命令: /clear 清空历史, /diff 显示变更, /quit 退出")
	fmt.Println("=" + strings.Repeat("=", 49))
	fmt.Println()

	go client.ReceiveLoop()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	scanner := bufio.NewScanner(os.Stdin)
	for {
		fmt.Print("You> ")

		select {
		case <-sigChan:
			fmt.Println("\n正在退出...")
			return
		default:
		}

		if !scanner.Scan() {
			break
		}

		input := strings.TrimSpace(scanner.Text())
		if input == "" {
			continue
		}

		switch strings.ToLower(input) {
		case "/quit", "/exit", "quit", "exit":
			fmt.Println("再见!")
			return
		case "/clear", "clear":
			if err := client.Clear(); err != nil {
				fmt.Printf("错误: %v\n", err)
			} else {
				fmt.Println("[历史已清空]")
			}
			continue
		case "/diff":
			modified, deleted, err := client.GetDiff()
			if err != nil {
				fmt.Printf("错误: %v\n", err)
			} else if len(modified) == 0 && len(deleted) == 0 {
				fmt.Println("[本次会话无文件变更]")
			} else {
				for _, f := range modified {
					fmt.Printf("  M %s\n", f)
				}
				for _, f := range deleted {
					fmt.Printf("  D %s\n", f)
				}
			}
			continue
		case "/help", "help":
			printHelp()
			continue
		}

		fmt.Print("Claude> ")
		if err := client.SendMessage(input); err != nil {
			fmt.Printf("\n错误: 发送失败: %v\n", err)
			continue
		}

		<-client.respReady
		fmt.Println()
	}
}

func printHelp() {
	fmt.Println(`命令:
  /help     显示帮助
  /clear    清空对话历史
  /diff     显示本次会话修改的文件
  /quit     退出程序

直接输入内容与 Claude 对话`)
}
