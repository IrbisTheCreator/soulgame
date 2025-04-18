package main

import (
	"database/sql"
	"html/template"
	"log"
	"path/filepath"
	"soulgame/pkg/models"
	"time"
)

type templateData struct {
	Quest   *models.Quests
	Quests  []*models.Quests
	Phone   string
	Status  []int
	Souls   int
	Error   string
	Shop    []*models.Shop
	Admin   []*models.Adm
	User    *models.User
	History []*models.HistoryBought
	Log     []*models.Log
}

func newTemplateCache(dir string) (map[string]*template.Template, error) {

	cache := map[string]*template.Template{}

	pages, err := filepath.Glob(filepath.Join(dir, "*.page.tmpl"))
	if err != nil {
		return nil, err
	}

	for _, page := range pages {

		name := filepath.Base(page)
		ts, err := template.ParseFiles(page)
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.layout.tmpl"))
		if err != nil {
			return nil, err
		}

		ts, err = ts.ParseGlob(filepath.Join(dir, "*.partial.tmpl"))
		if err != nil {
			return nil, err
		}

		cache[name] = ts
	}

	return cache, nil
}

func ScheduleDailyUpdate(db *sql.DB) {
	go func() {
		for {
			// Вычисляем время до следующей полночи
			now := time.Now()
			// Вычисляем время следующего обновления (сегодня или завтра в 21:00)
			next := time.Date(now.Year(), now.Month(), now.Day(), 21, 0, 0, 0, now.Location())
			if now.After(next) {
				// Если сейчас уже после 21:00, планируем на завтра
				next = next.Add(24 * time.Hour)
			}
			duration := next.Sub(now)

			// Ждем до полночи
			time.Sleep(duration)

			// Выполняем обновление
			if err := UpdateRandomQuestsStatus(db); err != nil {
				log.Printf("Failed to update quests status: %v", err)
			} else {
				log.Println("Successfully updated quests status at", time.Now())
			}
		}
	}()
}
