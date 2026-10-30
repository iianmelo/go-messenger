package main

import (
	"log"

	"go.messenger/config"
	"go.messenger/database"
	"go.messenger/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Print("No .env file found")
	}
}

const gopherDraw = `
       ,_---~~~~~----._
_,,_,*^____      _____''*g*\"*,          Welcome to go message!
/ __/ /'     ^.  /      \ ^@q   f     /
[  @f | @))    |  | @))   l  0 _/    /
\'/   \~____ / __ \_____/    \      /
|           _l__l_           I     /
}          [______]           I   /
]            | | |            |
]             ~ ~             |
|                            |
|                           |`

func main() {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization",
		AllowMethods: "GET,POST,HEAD,PUT,PATCH,DELETE,OPTIONS",
	}))

	var fireAuth = database.OptionalFirebaseAuth()
	if config.LocalMode() {
		log.Println("AUTH_MODE=local (JWT)")
	}
	database.ConnectDb()
	routes.SetupRoutes(app, fireAuth)

	log.Println(gopherDraw)
	log.Println("escutando :3000")

	if err := app.Listen(":3000"); err != nil {
		log.Fatal(err)
	}
}
