package main

import (
	"context"
	_ "embed"
	"encoding/json"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	"com.github.cork89/connections/models"
	"com.github.cork89/connections/templates"
	"github.com/joho/godotenv"
	"github.com/revrost/go-openrouter"
)

//go:embed utils/badwords.json
var badwordsjson string
var badwords []string

var tmpl = make(map[string]*template.Template)

var model *openrouter.Client

type VerifyFailureReason string

const (
	NoChange        string              = "nochange"
	MissingCategory VerifyFailureReason = "Missing following categories: [%s]"
	NotEnoughWords  VerifyFailureReason = "Not enough words in category: [%s]"
	WordsNotAllowed VerifyFailureReason = "Some words not allowed in category: [%s]"
	DuplicateWords  VerifyFailureReason = "Duplicate words found: [%s]"
)

func homeHtmxHandler(w http.ResponseWriter, r *http.Request) {
	head := templates.HomeHead()
	body := templates.HomeBody(models.Desktop, models.I18N{})
	component := templates.BaseHtmx(head, body)
	w.Header().Set("Content-Type", "text/html")
	err := component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func determineUserAgentType(ua string) models.UserAgentType {
	if strings.Contains(strings.ToLower(ua), "mobile") {
		return models.Mobile
	} else {
		return models.Desktop
	}
}

func homeHandler(w http.ResponseWriter, r *http.Request) {
	userAgentType := determineUserAgentType(r.UserAgent())
	i18n := r.Context().Value(models.I18Nctx).(models.I18N)

	homeHead := templates.HomeHead()
	homeBody := templates.HomeBody(userAgentType, i18n)
	component := templates.Base(homeHead, homeBody, i18n)

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
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatalln("failed to load .env")
	}

	err = initDataaccess()
	if err != nil {
		log.Fatalf("failed to connect to db, err: %v", err)
	}

	err = json.Unmarshal([]byte(badwordsjson), &badwords)

	if err != nil {
		log.Fatalf("failed to unmarshal bad words list")
	}

	model = openrouter.NewClient(os.Getenv("OPENROUTER_API_KEY"),
		openrouter.WithXTitle("Connections"),
		openrouter.WithHTTPReferer("https://hearteyesemoji.dev"))

	// tmpl["headsup"] = template.Must(template.ParseFiles("static/headsup.html", "static/base.html"))

	router := http.NewServeMux()
	rateLimitRouter := http.NewServeMux()

	router.Handle("GET /static/", http.StripPrefix("/static/", http.FileServer(http.Dir("static"))))

	router.HandleFunc("GET /create/", createHandler)
	router.HandleFunc("POST /create/", func(w http.ResponseWriter, r *http.Request) { createPostHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /create/verify/", verifyHandler)
	rateLimitRouter.HandleFunc("POST /create/suggestions/", suggestionsHandler)
	router.Handle("POST /create/suggestions/", RateLimiter(rateLimitRouter))
	router.HandleFunc("GET /mygames/", mygamesHandler)

	router.HandleFunc("POST /game/{gameId}/check/", func(w http.ResponseWriter, r *http.Request) { checkHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /game/{gameId}/shuffle/", func(w http.ResponseWriter, r *http.Request) { shuffleHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /game/{gameId}/deselectAll/", func(w http.ResponseWriter, r *http.Request) { deselectHandler(w, r, realDataAccess) })
	router.HandleFunc("POST /game/{gameId}/reset/", func(w http.ResponseWriter, r *http.Request) { resetHandler(w, r, realDataAccess) })
	router.HandleFunc("GET /game/{gameId}/", func(w http.ResponseWriter, r *http.Request) { gameHandler(w, r, realDataAccess) })

	// router.HandleFunc("GET /headsup/", headsupHandler)

	router.HandleFunc("GET /random/", func(w http.ResponseWriter, r *http.Request) { randomHandler(w, realDataAccess) })
	router.HandleFunc("GET /robots.txt", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "static/robots.txt")
	})

	router.HandleFunc("GET /settings/", settingsHandler)
	router.HandleFunc("POST /settings/", settingsPostHandler)

	router.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) { randomHandler(w, realDataAccess) })

	// HTMX exploration
	// router.HandleFunc("GET /randomHtmx/", func(w http.ResponseWriter, r *http.Request) { randomHtmxHandler(w, r, realDataAccess) })
	// router.HandleFunc("GET /mygamesHtmx/", mygamesHtmxHandler)
	// router.HandleFunc("GET /createHtmx/", createHtmxHandler)
	// router.HandleFunc("GET /homeHtmx/", homeHtmxHandler)

	stack := CreateStack(
		Logging,
		Session,
		StaticCompression,
		// CacheControl,
		Settings,
	)

	server := http.Server{
		Addr:    ":8000",
		Handler: stack(router),
	}
	server.ListenAndServe()
}
