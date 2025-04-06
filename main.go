package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"

	"com.github.cork89/connections/templates"
)

//go:embed utils/badwords.json
var badwordsjson string
var badwords []string

var tmpl = make(map[string]*template.Template)

type VerifyFailureReason string

const (
	NoChange        string              = "nochange"
	MissingCategory VerifyFailureReason = "Missing following categories: [%s]"
	NotEnoughWords  VerifyFailureReason = "Not enough words in category: [%s]"
	WordsNotAllowed VerifyFailureReason = "Some words not allowed in category: [%s]"
	DuplicateWords  VerifyFailureReason = "Duplicate words found: [%s]"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	homeHead := templates.HomeHead()
	homeBody := templates.HomeBody()
	component := templates.Base(homeHead, homeBody)

	err := component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

// not used keeping commented code to review later
func longWordHandler(word string) bool {
	// words := make([]string, 0)
	startWords := strings.Split(word, " ")

	longWord := false

	for _, wd := range startWords {
		if len(wd) < 9 {
			// words = append(words, wd)
		} else {
			longWord = true
			// 	var partial string
			// 	for i, ch := range word {
			// 		partial = fmt.Sprintf("%s%c", partial, ch)
			// 		if i%7 == 0 && i > 0 {
			// 			if string(ch) != " " {
			// 				partial += "-"
			// 			}
			// 			words = append(words, partial)
			// 			partial = ""
			// 		}
			// 	}
			// 	if partial != "" {
			// 		words = append(words, partial)
			// 	}
			// 	lastWordIdx := len(words) - 1
			// 	words[lastWordIdx] = strings.Trim(words[lastWordIdx], "-")
			// }
			// words = append(words, " ")
		}
	}
	// var partial string
	// for i, ch := range word {
	// 	partial = fmt.Sprintf("%s%c", partial, ch)
	// 	if i%14 == 0 && i > 0 {
	// 		if string(ch) != " " {
	// 			partial += "- "
	// 		}
	// 		words = append(words, partial)
	// 		partial = ""
	// 	}
	// }
	// if partial != "" {
	// 	words = append(words, partial)
	// }
	// lastWordIdx := len(words) - 1
	// words[lastWordIdx] = strings.Trim(words[lastWordIdx], "-")
	return longWord
}

func main() {
	err := initDataaccess()
	if err != nil {
		log.Fatalf("failed to connect to db, err: %v", err)
	}

	err = json.Unmarshal([]byte(badwordsjson), &badwords)

	if err != nil {
		log.Fatalf("failed to unmarshal bad words list")
	}

	// tmpl["headsup"] = template.Must(template.ParseFiles("static/headsup.html", "static/base.html"))

	router := http.NewServeMux()

	router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("GET /create/", createHandler)
	router.HandleFunc("POST /create/", func(w http.ResponseWriter, r *http.Request) { createPostHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /create/verify/", verifyHandler)

	router.HandleFunc("GET /mygames/", mygamesHandler)

	router.HandleFunc("POST /game/{gameId}/check/", func(w http.ResponseWriter, r *http.Request) { checkHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /game/{gameId}/shuffle/", func(w http.ResponseWriter, r *http.Request) { shuffleHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /game/{gameId}/deselectAll/", func(w http.ResponseWriter, r *http.Request) { deselectHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /game/{gameId}/reset/", func(w http.ResponseWriter, r *http.Request) { resetHandler(w, r, realDataAccess) })
	router.HandleFunc("GET /game/{gameId}/", func(w http.ResponseWriter, r *http.Request) { gameHandler(w, r, realDataAccess) })

	// router.HandleFunc("GET /headsup/", headsupHandler)

	router.HandleFunc("GET /random/", func(w http.ResponseWriter, r *http.Request) { randomHandler(w, r, realDataAccess) })
	router.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})

	router.HandleFunc("GET /", homeHandler)

	stack := CreateStack(
		Logging,
		Session,
		StaticCompression,
	)

	server := http.Server{
		Addr:    ":8000",
		Handler: stack(router),
	}
	server.ListenAndServe()
}
