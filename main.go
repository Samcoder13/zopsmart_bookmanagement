package main

import (
	"database/sql"
	"fmt"
	"gofr.dev/pkg/gofr"
	"log"
	"net/http"
	"github.com/jinzhu/gorm"
	_"github.com/jinzhu/gorm/dialects/mysql"
	// _ "github.com/go-sql-driver/mysql"
)

// Book represents a book entity
type Book struct {
	ID    int    `json:"id"`
	Title string `json:"title"`
	Author string `json:"author"`
}

var db *gorm.DB

func init() {
	// Open a database connection
	var err error
	db, err = gorm.Open("mysql", "root:1234@tcp(localhost:3306)/booksdb?charset=utf8parseTime=True&loc=Local")
	if err != nil {
		log.Fatal(err)
	}

	// Create the "books" table if it doesn't exist
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS books (
		id INT AUTO_INCREMENT PRIMARY KEY,
		title VARCHAR(255) NOT NULL,
		author VARCHAR(255) NOT NULL
	)`)
	if err != nil {
		log.Fatal(err)
	}
	
}

func main() {
	// Initialise gofr object
	app := gofr.New()

	// Route to add a book
	app.POST("/books", addBookHandler)

	// Route to get all books
	app.GET("/books", getAllBooksHandler)

	// Route to get a single book by ID
	app.GET("/books/{id}", getBookByIDHandler)

	// Route to update a book by ID
	app.PUT("/books/{id}", updateBookHandler)

	// Route to delete a book by ID
	app.DELETE("/books/{id}", deleteBookHandler)

	// Start the server on the default port (8000)
	app.Start()
}

func addBookHandler(ctx *gofr.Context) (interface{}, error) {
	var newBook Book
	if err := ctx.Bind(&newBook); err != nil {
		return nil, err
	}

	// Insert the book into the database
	result, err := db.Exec("INSERT INTO books (title, author) VALUES (?, ?)", newBook.Title, newBook.Author)
	if err != nil {
		return nil, err
	}

	// Get the last inserted ID
	lastInsertedID, _ := result.LastInsertId()
	newBook.ID = int(lastInsertedID)

	return newBook, nil
}

func getAllBooksHandler(ctx *gofr.Context) (interface{}, error) {
	rows, err := db.Query("SELECT * FROM books")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookList []Book
	for rows.Next() {
		var book Book
		err := rows.Scan(&book.ID, &book.Title, &book.Author)
		if err != nil {
			return nil, err
		}
		bookList = append(bookList, book)
	}

	return bookList, nil
}

func getBookByIDHandler(ctx *gofr.Context) (interface{}, error) {
	id := ctx.PathParam("id")
	// bookID := gofr.ToInt(id)
	bookID := id

	var book Book
	err := db.QueryRow("SELECT * FROM books WHERE id=?", bookID).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil
			// err.Response{
			// 	StatusCode: 500,
			// 	Reason:     "Not found",
			// 	Detail:     err.New("database error"),
			// }
			// gofr.ErrNotFound(fmt.Sprintf("Book with ID %d not found", bookID))
		}
		return nil, err
	}

	return book, nil
}

func updateBookHandler(ctx *gofr.Context) (interface{}, error) {
	id := ctx.PathParam("id")
	// bookID := gofr.ToInt(id)
	bookID := id

	var updatedBook Book
	if err := ctx.Bind(&updatedBook); err != nil {
		return nil
		
	}

	// Check if the book exists
	if _, err := getBookByID(bookID); err != nil {
		return nil, err
	}

	// Update the book in the database
	_, err := db.Exec("UPDATE books SET title=?, author=? WHERE id=?", updatedBook.Title, updatedBook.Author, bookID)
	if err != nil {
		return nil, err
	}

	updatedBook.ID = bookID
	return updatedBook, nil
}

func deleteBookHandler(ctx *gofr.Context) (interface{}, error) {
	id := ctx.PathParam("id")
	
	bookID := id

	// Check if the book exists
	deletedBook, err := getBookByID(bookID)
	if err != nil {
		return nil, err
	}

	// Delete the book from the database
	_, err = db.Exec("DELETE FROM books WHERE id=?", bookID)
	if err != nil {
		return nil, err
	}

	return deletedBook, nil
}

func getBookByID(bookID int) (Book, error) {
	var book Book
	err := db.QueryRow("SELECT * FROM books WHERE id=?", bookID).Scan(&book.ID, &book.Title, &book.Author)
	if err != nil {
		if err == sql.ErrNoRows {
			return book
			
		}
		return book, err
	}
	return book, nil
}