// A super simple web server for use in testing.

package server

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"

	"github.com/google/uuid"
)

func errAuthorNotFound(authorID string) error {
	return fmt.Errorf("No such author: %s", authorID)
}

func errPublisherNotFound(publisherID string) error {
	return fmt.Errorf("No such publisher: %s", publisherID)
}

type server struct {
	sync.Mutex
	logger     *log.Logger
	authors    map[string]*Author
	publishers map[string]*Publisher
	books      map[string]*Book
}

func NewController(logger *log.Logger) *server {
	return &server{
		logger:     logger,
		authors:    map[string]*Author{},
		publishers: map[string]*Publisher{},
		books:      map[string]*Book{},
	}
}

func NewControllerWithBooks(logger *log.Logger, data []*Book) *server {
	authors := make(map[string]*Author, 0)
	publishers := make(map[string]*Publisher, 0)
	books := make(map[string]*Book, len(data))
	for _, book := range data {
		if book.Author != nil {
			if book.Author.ID != "" {
				if _, found := authors[book.Author.ID]; !found {
					authors[book.Author.ID] = book.Author
				}
			}
		}
		if book.Publisher != nil {
			if book.Publisher.ID != "" {
				if _, found := publishers[book.Publisher.ID]; !found {
					publishers[book.Publisher.ID] = book.Publisher
				}
			}
		}
		books[book.ID] = book
	}
	return &server{
		logger:     logger,
		authors:    authors,
		publishers: publishers,
		books:      books,
	}
}

func (s *server) Router() http.Handler {
	router := http.NewServeMux()
	router.Handle("/books/", handleBook(s))
	router.Handle("/books", handleBooks(s))
	return router
}

func handleBooks(s *server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "POST":
			postBooks(s, w, r)
			return
		case "PUT":
			putBooks(s, w, r)
			return
		case "GET":
			listBooks(s, w, r)
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	})
}

func handleBook(s *server) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case "GET":
			bookID := strings.Replace(r.URL.Path, "/books/", "", 1)
			getBook(s, w, r, string(bookID))
			return
		}
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	})
}

func getBook(
	s *server,
	w http.ResponseWriter,
	r *http.Request,
	bookID string,
) {
	book := s.getBook(bookID)
	if book == nil {
		http.Error(w, http.StatusText(http.StatusNotFound), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(book)
}

func listBooks(s *server, w http.ResponseWriter, r *http.Request) {
	// Our GET /books endpoint only supports a "sort" parameter
	params := r.URL.Query()
	if len(params) > 0 {
		if _, found := params["sort"]; !found {
			var msg string
			for key := range params {
				msg = fmt.Sprintf("invalid parameter: %s", key)
				break
			}
			http.Error(w, msg, http.StatusBadRequest)
			return
		}
	}
	var lbr ListBooksResponse
	lbr.Books = s.listBooks()
	w.Header().Set("Content-Type", "application/json; charset=UTF-8")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(&lbr)
}

func postBooks(s *server, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var cbr CreateBookRequest
	err := decoder.Decode(&cbr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	createdID, err := s.createBook(&cbr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(400)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	locURL := fmt.Sprintf("/books/%s", createdID)
	w.Header().Set("Location", locURL)
	w.WriteHeader(http.StatusCreated)
}

func (s *server) createBook(cbr *CreateBookRequest) (string, error) {
	s.Lock()
	defer s.Unlock()

	author, found := s.authors[cbr.AuthorID]
	if !found {
		return "", errAuthorNotFound(cbr.AuthorID)
	}

	publisher, found := s.publishers[cbr.PublisherID]
	if !found {
		return "", errPublisherNotFound(cbr.PublisherID)
	}
	createdID, err := uuid.NewRandom()
	if err != nil {
		panic(err)
	}

	s.books[cbr.Title] = &Book{
		Title:       cbr.Title,
		PublishedOn: cbr.PublishedOn,
		Pages:       cbr.Pages,
		ID:          createdID.String(),
		Author:      author,
		Publisher:   publisher,
	}
	return createdID.String(), nil
}

// putBooks accepts an array of Book entries and creates/replaces the Book
// entries in the API server. Not a great REST API design, but it allows us to
// test the PUT method and array pre-processing for the HTTP test hander
func putBooks(s *server, w http.ResponseWriter, r *http.Request) {
	decoder := json.NewDecoder(r.Body)
	var rbr ReplaceBooksRequest
	err := decoder.Decode(&rbr)
	if err != nil {
		w.Header().Set("Content-Type", "application/json; charset=UTF-8")
		w.WriteHeader(422)
		if err := json.NewEncoder(w).Encode(err); err != nil {
			panic(err)
		}
		return
	}

	for _, entry := range rbr {
		// Was ID field set? If so, replace the existing entry, otherwise
		// create a new book
		if entry.ID != "" {
			fmt.Printf("replacing book with ID %s\n", entry.ID)
		} else {
			cbr := CreateBookRequest{
				Title:       entry.Title,
				AuthorID:    entry.AuthorID,
				PublisherID: entry.PublisherID,
				Pages:       entry.Pages,
			}
			_, err := s.createBook(&cbr)
			if err != nil {
				fmt.Printf("XXXXXX: %s", err)
				w.Header().Set("Content-Type", "application/json; charset=UTF-8")
				w.WriteHeader(400)
				if err := json.NewEncoder(w).Encode(err); err != nil {
					panic(err)
				}
				return
			}
		}
	}

	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
}

func (s *server) listBooks() []*Book {
	res := make([]*Book, 0, len(s.books))
	for _, book := range s.books {
		res = append(res, book)
	}
	return res
}

func (s *server) getBook(bookID string) *Book {
	for _, book := range s.books {
		if book.ID == bookID {
			return book
		}
	}
	return nil
}

func (s *server) Log(args ...interface{}) {
	s.logger.Println(args...)
}

func (s *server) Panic(fs string, args ...interface{}) {
	s.logger.Fatalf(fs, args...)
}
