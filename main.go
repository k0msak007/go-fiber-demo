package main

import (
	"log"
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
	"github.com/k0msak007/go-fiber-postgres/models"
	"github.com/k0msak007/go-fiber-postgres/storage"
	"gorm.io/gorm"
)

type Book struct {
	Author    string `json:"author"`
	Title     string `json:"title"`
	Publisher string `json:"publisher"`
}

type Repository struct {
	DB *gorm.DB
}

func (r *Repository) CreateBook(context *fiber.Ctx) error {
	book := Book{}
	err := context.BodyParser(&book)
	if err != nil {
		context.Status(fiber.StatusUnprocessableEntity).JSON(
			&fiber.Map{
				"message": "request failed",
			},
		)
		return err
	}

	err = r.DB.Create(&book).Error
	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(
			&fiber.Map{
				"message": "could not create book",
			},
		)
		return err
	}

	context.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"message": "book has been added",
			"data":    &book,
		},
	)
	return nil
}

func (r *Repository) GetBooks(context *fiber.Ctx) error {
	bookModels := &[]models.Books{}

	err := r.DB.Find(&bookModels).Error
	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(
			&fiber.Map{
				"message": "could not get books",
			},
		)
		return err
	}

	context.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"message": "books fetched successfully",
			"data":    bookModels,
		},
	)
	return nil
}

func (r *Repository) DeleteBook(context *fiber.Ctx) error {
	bookModel := models.Books{}
	id := context.Params("id")
	if id == "" {
		context.Status(fiber.StatusInternalServerError).JSON(
			&fiber.Map{
				"message": "id cannot be empty",
			},
		)
		return nil
	}

	err := r.DB.Delete(bookModel, id).Error
	if err != nil {
		context.Status(fiber.StatusBadRequest).JSON(
			&fiber.Map{
				"message": "could not delete books",
			},
		)
		return err
	}

	context.Status(fiber.StatusOK).JSON(
		&fiber.Map{
			"message": "books deleted successfully",
		},
	)
	return nil
}

func (r *Repository) SetupRoutes(app *fiber.App) {
	api := app.Group("/api")
	api.Post("/create_books", r.CreateBook)
	api.Delete("/delete_book/:id", r.DeleteBook)
	// api.Get("/get_books/:id", r.GetBookID)
	api.Get("/books", r.GetBooks)
}

func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal(err)
	}

	config := &storage.Config{
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASS"),
		DBName:   os.Getenv("DB_NAME"),
		SSLMode:  os.Getenv("DB_SSLMODE"),
	}

	db, err := storage.NewConnection(config)
	if err != nil {
		log.Fatal("could not load the database")
	}

	err = models.MigrateBooks(db)
	if err != nil {
		log.Fatal("could not migrate db")
	}

	r := Repository{
		DB: db,
	}

	app := fiber.New()
	r.SetupRoutes(app)
	app.Listen(":8000")
}
