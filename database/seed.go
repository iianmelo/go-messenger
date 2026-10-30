package database

import (
	"encoding/json"
	"log"
	"os"

	"go.messenger/models"
	"golang.org/x/crypto/bcrypt"
)

const seedPassword = "senha123"

type seedMessage struct {
	SenderID   uint   `json:"senderId"`
	ReceiverID uint   `json:"receiverId"`
	Name       string `json:"name"`
	Text       string `json:"text"`
	Time       string `json:"time"`
}

// SeedIfEmpty inserts the Portuguese demo users and chats the front-end
// already shows in demo mode. It does nothing when the users table has rows
// or when SEED=false.
func SeedIfEmpty() {
	if os.Getenv("SEED") == "false" {
		return
	}
	if DB.Db == nil {
		return
	}

	var count int64
	if err := DB.Db.Model(&models.User{}).Count(&count).Error; err != nil {
		log.Printf("seed: count users: %v", err)
		return
	}
	if count > 0 {
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(seedPassword), bcrypt.DefaultCost)
	if err != nil {
		log.Printf("seed: hash password: %v", err)
		return
	}

	people := []struct {
		name  string
		email string
	}{
		{"Ana Bezerra", "ana.bezerra@cin.ufpe.br"},
		{"Carlos Lima", "carlos.lima@cin.ufpe.br"},
		{"Mariana Souza", "mariana.souza@cin.ufpe.br"},
		{"João Victor Rocha", "joao.rocha@cin.ufpe.br"},
		{"Pedro Alves", "pedro.alves@cin.ufpe.br"},
		{"Beatriz Nunes", "beatriz.nunes@cin.ufpe.br"},
	}

	users := make([]models.User, 0, len(people))
	for _, person := range people {
		user := models.User{
			Name:     person.name,
			Email:    person.email,
			Password: string(hash),
		}
		if err := DB.Db.Create(&user).Error; err != nil {
			log.Printf("seed: create %s: %v", person.email, err)
			return
		}
		user.FireToken = "local-" + itoa(user.ID)
		if err := DB.Db.Model(&user).Update("fire_token", user.FireToken).Error; err != nil {
			log.Printf("seed: token %s: %v", person.email, err)
			return
		}
		users = append(users, user)
	}

	byEmail := map[string]models.User{}
	for _, user := range users {
		byEmail[user.Email] = user
	}
	ana := byEmail["ana.bezerra@cin.ufpe.br"]
	joao := byEmail["joao.rocha@cin.ufpe.br"]
	carlos := byEmail["carlos.lima@cin.ufpe.br"]
	mariana := byEmail["mariana.souza@cin.ufpe.br"]
	pedro := byEmail["pedro.alves@cin.ufpe.br"]

	chats := []struct {
		other    models.User
		messages []seedMessage
	}{
		{
			other: joao,
			messages: []seedMessage{
				{joao.ID, ana.ID, joao.Name, "Ana, a apresentação de ESS ficou para sexta, no CIn.", "2024-12-06T15:12:00Z"},
				{ana.ID, joao.ID, ana.Name, "Pode ser. Eu mostro o front e o fluxo de login.", "2024-12-06T15:18:00Z"},
				{joao.ID, ana.ID, joao.Name, "Eu levo o roteiro da API em Go e o modelo do Postgres.", "2024-12-06T15:40:00Z"},
			},
		},
		{
			other: carlos,
			messages: []seedMessage{
				{carlos.ID, ana.ID, carlos.Name, "O serviço da API no Render está suspenso.", "2024-12-05T18:40:00Z"},
				{ana.ID, carlos.ID, ana.Name, "Sem o backend, o front abre com os dados de demonstração.", "2024-12-05T18:52:00Z"},
				{carlos.ID, ana.ID, carlos.Name, "Assim a banca vê as conversas mesmo sem o Firebase.", "2024-12-05T19:05:00Z"},
			},
		},
		{
			other: mariana,
			messages: []seedMessage{
				{mariana.ID, ana.ID, mariana.Name, "Conseguiu autenticar no Firebase?", "2024-12-04T11:02:00Z"},
				{ana.ID, mariana.ID, ana.Name, "O login chama POST /auth e guarda o token.", "2024-12-04T11:11:00Z"},
				{mariana.ID, ana.ID, mariana.Name, "Beleza. O Postgres fica só no serviço em Go.", "2024-12-04T11:20:00Z"},
			},
		},
		{
			other: pedro,
			messages: []seedMessage{
				{pedro.ID, ana.ID, pedro.Name, "Vou revisar o texto da arquitetura antes da entrega.", "2024-12-03T09:04:00Z"},
				{ana.ID, pedro.ID, ana.Name, "Deixei o diagrama com Next, Fiber, Firebase e Postgres.", "2024-12-03T09:15:00Z"},
			},
		},
	}

	for _, item := range chats {
		raw, err := json.Marshal(item.messages)
		if err != nil {
			log.Printf("seed: messages: %v", err)
			return
		}
		chat := models.Chat{
			UserID:     ana.ID,
			ReceiverID: item.other.ID,
			Messages:   raw,
		}
		if err := DB.Db.Create(&chat).Error; err != nil {
			log.Printf("seed: chat %s: %v", item.other.Name, err)
			return
		}
	}

	log.Printf("seed: %d usuários, senha de demonstração configurada, %d conversas", len(users), len(chats))
}

func itoa(id uint) string {
	if id == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for id > 0 {
		i--
		buf[i] = byte('0' + id%10)
		id /= 10
	}
	return string(buf[i:])
}
