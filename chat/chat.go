package chat

import (
	"fmt"
	"io"
	"log"

	"golang.org/x/net/context"
)

type Server struct {
	UnimplementedChatServiceServer
}

func (s *Server) MessageUnary(ctx context.Context, in *Message) (*Message, error) {
	log.Printf("Receive message body from client: %s", in.Body)
	return &Message{Body: "Hello From the Server!"}, nil
}

// 第一引数がリクエスト、第二引数に stream オブジェクト
func (s *Server) MessageServertream(in *Message, stream ChatService_MessageServertreamServer) error {
	log.Printf("ServerStream recv: %s", in.Body)
	// 例としてメッセージを3回送信
	for i := 0; i < 3; i++ {
		if err := stream.Send(&Message{
			Body: fmt.Sprintf("message %d", i),
		}); err != nil {
			return err
		}
	}
	return nil
}

// 引数は stream オブジェクトのみ
func (s *Server) MessageClientStream(stream ChatService_MessageClientStreamServer) error {
	// クライアントからの複数メッセージを受け取りつつ
	cnt := 0
	for {
		msg, err := stream.Recv()
		if err == io.EOF {
			// 受信完了 → 最後にまとめてレスポンス
			return stream.SendAndClose(&Message{
				Body: fmt.Sprintf("received %d messages", cnt),
			})
		}
		if err != nil {
			return err
		}
		log.Printf("ClientStream recv: %s", msg.Body)
		cnt++
	}
}

// 引数は stream オブジェクトのみ
func (s *Server) MessageBiStreams(stream ChatService_MessageBiStreamsServer) error {
	for {
		// クライアントから受信
		req, err := stream.Recv()
		if err == io.EOF {
			return nil // クライアント終了
		}
		if err != nil {
			return err
		}
		log.Printf("BiDi recv: %s", req.Body)

		// サーバ→クライアントも都度送信
		if err := stream.Send(&Message{
			Body: "echo: " + req.Body,
		}); err != nil {
			return err
		}
	}
}
