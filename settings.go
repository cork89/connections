package main

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"com.github.cork89/connections/models"
	"com.github.cork89/connections/templates"
)

func settingsHandler(w http.ResponseWriter, r *http.Request) {
	i18n := r.Context().Value(models.I18Nctx).(models.I18N)
	bitPackedSettings := r.Context().Value(models.Settingsctx).(models.BitPackedSettings)

	settingsHead := templates.SettingsHead()
	settingsBody := templates.SettingsBody(bitPackedSettings, i18n)
	component := templates.Base(settingsHead, settingsBody, i18n)

	err := component.Render(context.Background(), w)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}

func settingsPostHandler(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()

	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	lang := r.FormValue("lang")
	suggestions := r.FormValue("suggestions")

	bitPackedSettings := models.BitPackedSettings{Lang: models.English, Suggestions: suggestions == "on"}
	if lang == "fr" {
		bitPackedSettings.Lang = models.French
	} else if lang == "es" {
		bitPackedSettings.Lang = models.Spanish
	}

	cookie := http.Cookie{
		Name:     models.SettingsCookie,
		Value:    strconv.Itoa(bitPackedSettings.ToBitPacked()),
		Path:     "/",
		MaxAge:   int(time.Duration(2160 * time.Hour).Seconds()),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
	}

	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	http.SetCookie(w, &cookie)
	http.Redirect(w, r, "/settings/", http.StatusFound)
}
