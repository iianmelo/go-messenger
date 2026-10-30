package api_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"go.messenger/database"
	"go.messenger/routes"
)

func resetDB(t *testing.T) *fiber.App {
	t.Helper()
	t.Setenv("AUTH_MODE", "local")
	t.Setenv("JWT_SECRET", "test-secret")
	t.Setenv("SEED", "true")
	t.Setenv("DB_HOST", "127.0.0.1")
	t.Setenv("DB_USER", "messenger")
	t.Setenv("DB_PASSWORD", "senha")
	t.Setenv("DB_NAME", "messenger_test")
	t.Setenv("DB_PORT", "5432")

	database.ConnectDb()
	err := database.DB.Db.Exec(`TRUNCATE TABLE chats, group_members, "groups", users RESTART IDENTITY CASCADE`).Error
	if err != nil {
		t.Fatalf("truncate: %v", err)
	}
	database.SeedIfEmpty()

	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,HEAD,PUT,PATCH,DELETE,OPTIONS",
	}))
	routes.SetupRoutes(app, nil)
	return app
}

func do(t *testing.T, app *fiber.App, method, path, token string, body any) (*http.Response, []byte) {
	t.Helper()
	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}
	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Origin", "http://127.0.0.1:8080")
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatal(err)
	}
	payload, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Fatal(err)
	}
	return resp, payload
}

func login(t *testing.T, app *fiber.App, email, password string) string {
	t.Helper()
	resp, payload := do(t, app, http.MethodPost, "/auth", "", map[string]string{
		"email":    email,
		"password": password,
	})
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("login %s: %d %s", email, resp.StatusCode, payload)
	}
	var parsed struct {
		Token string `json:"token"`
		User  struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		} `json:"user"`
	}
	if err := json.Unmarshal(payload, &parsed); err != nil {
		t.Fatal(err)
	}
	if parsed.Token == "" || parsed.User.Email != email {
		t.Fatalf("login payload: %s", payload)
	}
	return parsed.Token
}

type chatDTO struct {
	ID       uint `json:"id"`
	Receiver struct {
		ID   uint   `json:"id"`
		Name string `json:"name"`
	} `json:"receiver"`
	Messages []struct {
		Text     string `json:"text"`
		SenderID uint   `json:"senderId"`
		Time     string `json:"time"`
	} `json:"messages"`
}

func chatsOf(t *testing.T, app *fiber.App, token string) []chatDTO {
	t.Helper()
	resp, payload := do(t, app, http.MethodGet, "/chats/user", token, nil)
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("chats: %d %s", resp.StatusCode, payload)
	}
	var chats []chatDTO
	if err := json.Unmarshal(payload, &chats); err != nil {
		t.Fatalf("chats json: %v %s", err, payload)
	}
	return chats
}

func TestAuthAndChats(t *testing.T) {
	app := resetDB(t)

	t.Run("rejects empty login", func(t *testing.T) {
		resp, _ := do(t, app, http.MethodPost, "/auth", "", map[string]string{"email": "", "password": ""})
		if resp.StatusCode != http.StatusBadRequest {
			t.Fatalf("status %d", resp.StatusCode)
		}
	})

	t.Run("rejects bad password", func(t *testing.T) {
		resp, _ := do(t, app, http.MethodPost, "/auth", "", map[string]string{
			"email":    "ana.bezerra@cin.ufpe.br",
			"password": "errada",
		})
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status %d", resp.StatusCode)
		}
	})

	token := login(t, app, "ana.bezerra@cin.ufpe.br", "senha123")

	t.Run("profile", func(t *testing.T) {
		resp, payload := do(t, app, http.MethodGet, "/users", token, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d %s", resp.StatusCode, payload)
		}
		if resp.Header.Get("Access-Control-Allow-Origin") != "*" {
			t.Fatalf("cors header %q", resp.Header.Get("Access-Control-Allow-Origin"))
		}
		var user struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(payload, &user); err != nil {
			t.Fatal(err)
		}
		if user.Name != "Ana Bezerra" || user.Email != "ana.bezerra@cin.ufpe.br" {
			t.Fatalf("user %+v", user)
		}
	})

	t.Run("protected route without token", func(t *testing.T) {
		resp, _ := do(t, app, http.MethodGet, "/users", "", nil)
		if resp.StatusCode != http.StatusUnauthorized {
			t.Fatalf("status %d", resp.StatusCode)
		}
	})

	t.Run("contacts exclude the signed-in user", func(t *testing.T) {
		resp, payload := do(t, app, http.MethodGet, "/users/all", token, nil)
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("status %d %s", resp.StatusCode, payload)
		}
		var contacts []struct {
			Name  string `json:"name"`
			Email string `json:"email"`
		}
		if err := json.Unmarshal(payload, &contacts); err != nil {
			t.Fatal(err)
		}
		if len(contacts) != 5 {
			t.Fatalf("contacts %d: %+v", len(contacts), contacts)
		}
		for _, contact := range contacts {
			if contact.Email == "ana.bezerra@cin.ufpe.br" {
				t.Fatal("list includes the current user")
			}
		}
	})

	t.Run("chats follow the latest message", func(t *testing.T) {
		chats := chatsOf(t, app, token)
		if len(chats) != 4 {
			t.Fatalf("len %d", len(chats))
		}
		if chats[0].Receiver.Name != "João Victor Rocha" {
			names := make([]string, len(chats))
			for i, chat := range chats {
				names[i] = chat.Receiver.Name
			}
			t.Fatalf("order %v", names)
		}
		last := chats[0].Messages[len(chats[0].Messages)-1]
		if last.Text == "" || last.Time == "" {
			t.Fatalf("last message %+v", last)
		}
	})

	t.Run("signup then empty chat list", func(t *testing.T) {
		resp, payload := do(t, app, http.MethodPost, "/auth/sign-up", "", map[string]string{
			"name":      "Lia Costa",
			"email":     "lia.costa@cin.ufpe.br",
			"password":  "senha123",
			"photo_url": "",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("signup %d %s", resp.StatusCode, payload)
		}
		lia := login(t, app, "lia.costa@cin.ufpe.br", "senha123")
		chats := chatsOf(t, app, lia)
		if len(chats) != 0 {
			t.Fatalf("expected empty list, got %d", len(chats))
		}
	})

	t.Run("reset password accepts the email", func(t *testing.T) {
		resp, payload := do(t, app, http.MethodPost, "/resetpassword", "", map[string]string{
			"email": "ana.bezerra@cin.ufpe.br",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("reset %d %s", resp.StatusCode, payload)
		}
	})

	t.Run("patch profile", func(t *testing.T) {
		resp, payload := do(t, app, http.MethodPatch, "/users", token, map[string]string{
			"name":  "Ana Bezerra",
			"photo": "",
		})
		if resp.StatusCode != http.StatusOK {
			t.Fatalf("patch %d %s", resp.StatusCode, payload)
		}
	})
}

func TestWebSocketMessagePersists(t *testing.T) {
	app := resetDB(t)
	token := login(t, app, "ana.bezerra@cin.ufpe.br", "senha123")
	chats := chatsOf(t, app, token)
	joao := chats[0].Receiver.ID

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	addr := ln.Addr().String()
	ln.Close()

	go func() {
		_ = app.Listen(addr)
	}()
	t.Cleanup(func() { _ = app.Shutdown() })

	deadline := time.Now().Add(3 * time.Second)
	for {
		conn, err := net.DialTimeout("tcp", addr, 100*time.Millisecond)
		if err == nil {
			conn.Close()
			break
		}
		if time.Now().After(deadline) {
			t.Fatal("api did not listen")
		}
		time.Sleep(20 * time.Millisecond)
	}

	wsURL := "ws://" + addr + "/chats?token=" + token
	socket, _, err := websocket.DefaultDialer.Dial(wsURL, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer socket.Close()

	payload, _ := json.Marshal(map[string]any{
		"text":       "Mensagem de teste da API",
		"receiverId": joao,
	})
	if err := socket.WriteMessage(websocket.TextMessage, payload); err != nil {
		t.Fatal(err)
	}

	deadline = time.Now().Add(3 * time.Second)
	for {
		updated := chatsOf(t, app, token)
		for _, chat := range updated {
			if chat.Receiver.ID != joao {
				continue
			}
			last := chat.Messages[len(chat.Messages)-1]
			if last.Text == "Mensagem de teste da API" && last.SenderID != 0 {
				return
			}
		}
		if time.Now().After(deadline) {
			t.Fatal("message was not stored")
		}
		time.Sleep(50 * time.Millisecond)
	}
}
