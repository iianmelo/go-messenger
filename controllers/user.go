package controllers

import (
	"github.com/gofiber/fiber/v2"
	"go.messenger/database"
	"go.messenger/models"
	"go.messenger/services"
)

func GetUser(c *fiber.Ctx) error {
	firebaseId := c.Locals("firebaseId")

	var user models.User

	result := database.DB.Db.First(&user, "fire_token = ?", firebaseId)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found"})
	}

	return c.Status(fiber.StatusOK).JSON(services.PublicUser(user))
}

type UserUpdates struct {
	Name  string `json:"name"`
	Photo string `json:"photo"`
}

func UpdateUser(c *fiber.Ctx) error {

	var updates UserUpdates

	if err := c.BodyParser(&updates); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Cannot parse JSON"})
	}

	if updates.Name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{"error": "Name cannot be empty"})
	}

	firebaseId := c.Locals("firebaseId")

	var user models.User

	result := database.DB.Db.First(&user, "fire_token = ?", firebaseId)

	if result.Error != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "User not found"})
	}

	user.Name = updates.Name
	user.Avatar = updates.Photo

	if err := database.DB.Db.Save(&user).Error; err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": "Failed to update user"})
	}

	return c.Status(fiber.StatusOK).JSON(services.PublicUser(user))
}

func GetUsers(c *fiber.Ctx) error {
	current, err := currentUser(c)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "User not found"})
	}
	contacts, err := services.GetUsers(current.ID)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(fiber.Map{"error": "Contacts not found"})
	}

	return c.Status(fiber.StatusOK).JSON(contacts)
}

func currentUser(c *fiber.Ctx) (models.User, error) {
	var user models.User
	err := database.DB.Db.First(&user, "fire_token = ?", c.Locals("firebaseId")).Error
	return user, err
}
