package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/hryt430/gRPC-catchup/chat"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func main() {
	// 1) サーバ接続
	conn, err := grpc.Dial("localhost:9000", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		log.Fatalf("failed to connect: %v", err)
	}
	defer conn.Close()

	client := chat.NewChatServiceClient(conn)
	scanner := bufio.NewScanner(os.Stdin)

	// 2) メニュー表示
	fmt.Println(`
===== ChatClient =====
1) Unary (MessageUnary)
2) Server Stream (MessageServertream)
3) Client Stream (MessageClientStream)
4) Bidirectional Stream (MessageBiStreams)
5) Exit
----------------------`)

	for {
		fmt.Print("Select> ")
		if !scanner.Scan() {
			break
		}
		choice := scanner.Text()

		switch choice {
		case "1":
			fmt.Print("Enter text for Unary> ")
			scanner.Scan()
			text := scanner.Text()
			doUnary(client, text)

		case "2":
			fmt.Print("Enter text for Server Stream> ")
			scanner.Scan()
			text := scanner.Text()
			doServerStream(client, text)

		case "3":
			fmt.Println("Enter texts for Client Stream (empty line to finish):")
			var msgs []string
			for {
				fmt.Print("> ")
				scanner.Scan()
				line := scanner.Text()
				if line == "" {
					break
				}
				msgs = append(msgs, line)
			}
			doClientStream(client, msgs)

		case "4":
			fmt.Println("Enter texts for BiDi Stream (empty line to finish send):")
			var msgs []string
			for {
				fmt.Print("> ")
				scanner.Scan()
				line := scanner.Text()
				if line == "" {
					break
				}
				msgs = append(msgs, line)
			}
			doBiDiStream(client, msgs)

		case "5":
			fmt.Println("Bye!")
			return
		default:
			fmt.Println("Invalid choice")
		}
	}
}

// 1. 単一リクエスト／単一レスポンス
func doUnary(client chat.ChatServiceClient, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	resp, err := client.MessageUnary(ctx, &chat.Message{Body: text})
	if err != nil {
		log.Printf("Unary error: %v\n", err)
		return
	}
	fmt.Printf("Unary Response: %s\n\n", resp.Body)
}

// 2. サーバーストリーミング
func doServerStream(client chat.ChatServiceClient, text string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.MessageServertream(ctx, &chat.Message{Body: text})
	if err != nil {
		log.Printf("ServerStream error: %v\n", err)
		return
	}
	fmt.Println("Server Stream Responses:")
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			log.Printf("Recv error: %v\n", err)
			return
		}
		fmt.Printf("  - %s\n", msg.Body)
	}
	fmt.Println()
}

// 3. クライアントストリーミング
func doClientStream(client chat.ChatServiceClient, texts []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	stream, err := client.MessageClientStream(ctx)
	if err != nil {
		log.Printf("ClientStream error: %v\n", err)
		return
	}
	for _, t := range texts {
		if err := stream.Send(&chat.Message{Body: t}); err != nil {
			log.Printf("Send error: %v\n", err)
			return
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		log.Printf("CloseAndRecv error: %v\n", err)
		return
	}
	fmt.Printf("Client Stream Response: %s\n\n", resp.Body)
}

// 4. 双方向ストリーミング
func doBiDiStream(client chat.ChatServiceClient, texts []string) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	stream, err := client.MessageBiStreams(ctx)
	if err != nil {
		log.Printf("BiDiStream error: %v\n", err)
		return
	}

	// 受信ゴルーチン
	done := make(chan struct{})
	go func() {
		for {
			msg, err := stream.Recv()
			if err == io.EOF {
				close(done)
				return
			}
			if err != nil {
				log.Printf("Recv error: %v\n", err)
				close(done)
				return
			}
			fmt.Printf("  << %s\n", msg.Body)
		}
	}()

	// 送信
	for _, t := range texts {
		fmt.Printf("  >> %s\n", t)
		if err := stream.Send(&chat.Message{Body: t}); err != nil {
			log.Printf("Send error: %v\n", err)
			break
		}
	}
	stream.CloseSend()
	<-done
	fmt.Println()
}
