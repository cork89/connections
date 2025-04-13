package models

type Color string
type GuessResult string
type Context string

const (
	Yellow      Color       = "yellow"
	Green       Color       = "green"
	Blue        Color       = "blue"
	Purple      Color       = "purple"
	Undefined   Color       = ""
	Three       GuessResult = "three"
	Four        GuessResult = "four"
	Other       GuessResult = "other"
	I18Nctx     Context     = "i18ncontext"
	Settingsctx Context     = "settingscontext"
)

type I18N struct {
	Settings          string
	Language          string
	CreateSuggestions string
	SaveChanges       string
	Home              string
	Create            string
	MyGames           string
	PlayRandom        string
	HeroTag           string
}

func (s *I18N) English() *I18N {
	s.Settings = "Settings"
	s.Language = "Language"
	s.CreateSuggestions = "Suggestions while creating"
	s.SaveChanges = "Save Changes"
	s.Home = "Home"
	s.Create = "Create"
	s.MyGames = "My Games"
	s.PlayRandom = "Play Random"
	s.HeroTag = "Make your own connections game!"
	return s
}

func (s *I18N) French() *I18N {
	s.Settings = "Paramètres"
	s.Language = "Langue"
	s.CreateSuggestions = "Créer des suggestions"
	s.SaveChanges = "Enregistrer les modifications"
	s.Home = "Accueil"
	s.Create = "Créer"
	s.MyGames = "Mes Jeux"
	s.PlayRandom = "Jeu Aléatoire"
	s.HeroTag = "Créez votre propre jeu de connexions!"
	return s
}

func (s *I18N) Spanish() *I18N {
	s.Settings = "Configuración"
	s.Language = "Lenguaje"
	s.CreateSuggestions = "Generar sugerencias"
	s.SaveChanges = "Guardar Cambios"
	s.Home = "Inicio"
	s.Create = "Crear"
	s.MyGames = "Mis juegos"
	s.PlayRandom = "Jugar al azar"
	s.HeroTag = "¡Crea tu propio juego de conexiones!"
	return s
}
