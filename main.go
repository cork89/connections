package main

import (
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"strings"
)

//go:embed utils/badwords.json
var badwordsjson string
var badwords []string

var tmpl = make(map[string]*template.Template)

type Color string
type GuessResult string
type VerifyFailureReason string

const (
	Yellow          Color               = "yellow"
	Green           Color               = "green"
	Blue            Color               = "blue"
	Purple          Color               = "purple"
	Undefined       Color               = ""
	NoChange        string              = "nochange"
	Three           GuessResult         = "three"
	Four            GuessResult         = "four"
	Other           GuessResult         = "other"
	MissingCategory VerifyFailureReason = "Missing following categories: [%s]"
	NotEnoughWords  VerifyFailureReason = "Not enough words in category: [%s]"
	WordsNotAllowed VerifyFailureReason = "Some words not allowed in category: [%s]"
	DuplicateWords  VerifyFailureReason = "Duplicate words found: [%s]"
)

func homeHandler(w http.ResponseWriter, r *http.Request) {
	log.Println("homeHandler, url: ", r.URL.Path)

	err := tmpl["home"].ExecuteTemplate(w, "base.html", nil)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

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

	// words := make([]Word, 0, 16)
	// group1 := []Word{{Id: 1, Word: "word1", Category: Category{CategoryId: 1}}, {Id: 2, Word: "word2", Category: Category{CategoryId: 1}}, {Id: 3, Word: "word3", Category: Category{CategoryId: 1}}, {Id: 4, Word: "word4", Category: Category{CategoryId: 1}}}
	// group2 := []Word{{Id: 5, Word: "word5", Category: Category{CategoryId: 2}}, {Id: 6, Word: "word6", Category: Category{CategoryId: 2}}, {Id: 7, Word: "word7", Category: Category{CategoryId: 2}}, {Id: 8, Word: "word8", Category: Category{CategoryId: 2}}}
	// group3 := []Word{{Id: 9, Word: "word9", Category: Category{CategoryId: 3}}, {Id: 10, Word: "word10", Category: Category{CategoryId: 3}}, {Id: 11, Word: "word11", Category: Category{CategoryId: 3}}, {Id: 12, Word: "word12", Category: Category{CategoryId: 3}}}
	// group4 := []Word{{Id: 13, Word: "word13", Category: Category{CategoryId: 4}}, {Id: 14, Word: "word14", Category: Category{CategoryId: 4}}, {Id: 15, Word: "word15", Category: Category{CategoryId: 4}}, {Id: 16, Word: "word16", Category: Category{CategoryId: 4}}}

	// words = append(words, group1...)
	// words = append(words, group2...)
	// words = append(words, group3...)
	// words = append(words, group4...)
	// data.Words = words

	funcTemp := template.New("funcs").Funcs(template.FuncMap{"mod": func(i, j int) bool { return i%j == 0 }, "getColor": func(i int) Color {
		if i == 1 {
			return Yellow
		} else if i == 2 {
			return Green
		} else if i == 3 {
			return Blue
		} else if i == 4 {
			return Purple
		} else {
			return Undefined
		}
	},
		"getGuesses": func(i int) string { return strings.Repeat("<span class=\"mistakes-bubble\"></span>", i) },
		"times": func(n int) []struct{} {
			return make([]struct{}, n)
		},
		"longWords": longWordHandler,
	})
	tmpl["home"] = template.Must(template.ParseFiles("static/home.html", "static/base.html"))
	tmpl["game"] = template.Must(funcTemp.ParseFiles("static/board.html", "static/game.html", "static/base.html"))
	tmpl["board"] = template.Must(funcTemp.ParseFiles("static/board.html"))
	tmpl["create"] = template.Must(template.ParseFiles("static/create.html", "static/base.html"))
	tmpl["404"] = template.Must(template.ParseFiles("static/404.html", "static/base.html"))

	router := http.NewServeMux()

	router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("GET /create/", createHandler)
	router.HandleFunc("POST /create/", createPostHandler)
	router.HandleFunc("POST /create/verify/", verifyHandler)

	router.HandleFunc("POST /game/{gameId}/check/", checkHandler)
	router.HandleFunc("POST /game/{gameId}/shuffle/", shuffleHandler)
	router.HandleFunc("POST /game/{gameId}/deselectAll/", deselectHandler)
	router.HandleFunc("POST /game/{gameId}/reset/", resetHandler)
	router.HandleFunc("GET /game/{gameId}/", gameHandler)

	router.HandleFunc("GET /random/", randomHandler)
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
