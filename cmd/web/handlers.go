package main

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/lib/pq"
	"golang.org/x/crypto/bcrypt"
)

var (
	signingKey = []byte("cyber")
)

type Claims struct {
	Phone   string `json:"phone"`
	User_id int    `json:"user_id"`
	Dostup  int    `json:"dostup"`
	jwt.RegisteredClaims
}

func (app *application) home(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		app.notFound(w)
		return
	}

	s, err := app.quests.Latest()
	if err != nil {
		app.serverError(w, err)
		return
	}

	status := make([]int, 4)
	phone, user_id, _, err := homeHandler(w, r)
	var soul int
	if err == nil {
		for i := 0; i < 3; i++ {

			stat, err := app.quests.Checking(s[i].Quest_id, user_id)
			if err != nil {
				status[i] = 0
			}
			status[i] = stat
		}
		status[3], _ = app.quests.Checking(10, user_id)
		soul = app.quests.SoulCheck(user_id)

	}

	app.render(w, r, "home.page.tmpl", &templateData{
		Quests: s,
		Phone:  phone,
		Status: status,
		Souls:  soul,
	})
}

func (app *application) showSnippet(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || id < 0 {
		http.Error(w, "Invalid quest ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		return
	}

	_, user_id, _, err := homeHandler(w, r)

	if err != nil {
		return
	}

	err = app.quests.Take(id, user_id, 0)
	if err != nil {
		return
	}

	response := map[string]interface{}{
		"success": true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (app *application) newUser(w http.ResponseWriter, r *http.Request) {

	_, _, _, err := homeHandler(w, r)
	if err == nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.Method != http.MethodPost {
		app.render(w, r, "reg.page.tmpl", &templateData{})
		return
	}

	phone := r.FormValue("phone")
	pass := r.FormValue("password")

	_, err = app.quests.Verif(phone)
	if err == nil {
		app.render(w, r, "reg.page.tmpl", &templateData{
			Error: "Пользователь с таким телефоном уже зарегистрирован",
		})
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		app.serverError(w, err)

		return
	}
	err = app.quests.Regist(phone, string(passHash))
	if err != nil {
		app.serverError(w, err)
		return
	}
	app.render(w, r, "reg.page.tmpl", &templateData{})
}

func (app *application) login(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		_, _, _, err := homeHandler(w, r)
		if err == nil {
			http.Redirect(w, r, "/", http.StatusFound)
			return
		}
		app.render(w, r, "login.page.tmpl", &templateData{})
		return
	}

	phone := r.FormValue("phone")
	pass := r.FormValue("password")

	user, err := app.quests.Verif(phone)
	if err != nil {
		app.render(w, r, "login.page.tmpl", &templateData{
			Error: "Пользователь с таким телефоном не зарегистрирован",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(pass))
	if err != nil {
		app.render(w, r, "login.page.tmpl", &templateData{
			Error: "Неверный  пароль",
		})
		return
	}

	//создаем JWT
	expirationTime := time.Now().Add(55 * time.Minute)
	claims := &Claims{
		Phone:   user.Phone,
		User_id: user.User_id,
		Dostup:  user.Dostup,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString(signingKey)
	if err != nil {
		app.serverError(w, err)
		return
	}

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   tokenString,
		Expires: expirationTime,
		Path:    "/",
	})

	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *application) shop(w http.ResponseWriter, r *http.Request) {

	shops, err := app.quests.Products()
	if err != nil {
		return
	}

	phone, user_id, _, err := homeHandler(w, r)
	var soul int
	if err == nil {
		soul = app.quests.SoulCheck(user_id)
	}

	his, err := app.quests.HistoryB(user_id)
	if err != nil {
		return
	}

	app.render(w, r, "shop.page.tmpl", &templateData{
		Shop:    shops,
		Phone:   phone,
		Souls:   soul,
		History: his,
	})

}

func (app *application) admin(w http.ResponseWriter, r *http.Request) {

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	adm, _ := app.quests.ListCl()

	app.render(w, r, "show.page.tmpl", &templateData{
		Admin: adm,
	})
}

func (app *application) unlog(w http.ResponseWriter, r *http.Request) {

	http.SetCookie(w, &http.Cookie{
		Name:    "token",
		Value:   "",
		Expires: time.Now().Add(-time.Hour), // Устанавливаем срок действия в прошлом
		Path:    "/",                        // Путь должен совпадать с путем, по которому была установлена кука
	})

	http.Redirect(w, r, "/", http.StatusFound)

}

func (app *application) createquest(w http.ResponseWriter, r *http.Request) {

	if r.Method != http.MethodPost {
		app.render(w, r, "createquest.page.tmpl", &templateData{})
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	price, _ := strconv.Atoi(r.FormValue("price"))
	exp, err := strconv.Atoi(r.FormValue("exp"))

	if err != nil {
		app.render(w, r, "createquest.page.tmpl", &templateData{
			Error: "В опыте не должно быть ничего кроме цифр",
		})
		return
	}

	err = app.quests.CreateQ(title, description, price, exp)
	if err != nil {
		app.render(w, r, "createquest.page.tmpl", &templateData{
			Error: "Задание не создалось",
		})
		return
	}

	app.render(w, r, "createquest.page.tmpl", &templateData{
		Error: "Задание успешно создалось",
	})

}

func (app *application) complete(w http.ResponseWriter, r *http.Request) {
	quest_id, err := strconv.Atoi(r.URL.Query().Get("id"))
	if err != nil || quest_id < 0 {
		http.Error(w, "Invalid quest ID", http.StatusBadRequest)
		return
	}

	if r.Method != http.MethodPost {
		return
	}

	_, user_id, _, err := homeHandler(w, r)

	if err != nil {
		return
	}

	err = app.quests.Take(quest_id, user_id, 1)
	if err != nil {
		return
	}

	response := map[string]interface{}{
		"success": true,
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

func (app *application) profile(w http.ResponseWriter, r *http.Request) {

	phone, user_id, _, err := homeHandler(w, r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	user, err := app.quests.Info(user_id)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	app.render(w, r, "profile.page.tmpl", &templateData{
		Phone: phone,
		User:  user,
		Souls: user.Balance,
	})

}

func (app *application) purchase(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	_, user_id, _, err := homeHandler(w, r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	var request struct {
		Item_id int `json:"item_id"`
		Club_id int `json:"club_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	stat, err := app.quests.Buying(user_id, request.Item_id, request.Club_id)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
		return
	}
	if !stat {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]string{
			"error": "Purchase failed",
		})
		return
	}
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Purchase successful",
	})
}

func (app *application) repass(w http.ResponseWriter, r *http.Request) {
	phone, user_id, _, err := homeHandler(w, r)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}
	soul := app.quests.SoulCheck(user_id)

	if r.Method != "POST" {

		app.render(w, r, "repass.page.tmpl", &templateData{
			Phone: phone,
			Souls: soul,
		})
		return
	}

	old := r.FormValue("old_password")

	user, err := app.quests.Verif(phone)
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Pass), []byte(old))
	if err != nil {
		app.render(w, r, "repass.page.tmpl", &templateData{
			Phone: phone,
			Souls: soul,
			Error: "Неверный  пароль",
		})
		return
	}

	new := r.FormValue("new_password")
	confirm := r.FormValue("confirm_password")

	if new != confirm {
		app.render(w, r, "repass.page.tmpl", &templateData{
			Phone: phone,
			Souls: soul,
			Error: "Пароли не совпадают",
		})
		return
	}

	passHash, err := bcrypt.GenerateFromPassword([]byte(new), bcrypt.DefaultCost)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.quests.UpdatePass(user_id, string(passHash))
	if err != nil {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	http.Redirect(w, r, "/", http.StatusFound)
}

func (app *application) ready(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	var request struct {
		TakeID   int `json:"take_id"`
		UserId   int `json:"user_id"`
		Price    int `json:"price"`
		Exp      int `json:"exp"`
		Quest_id int `json:"quest_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}
	err = app.quests.Compl(request.TakeID, request.UserId, request.Price, request.Exp, request.Quest_id)
	if err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

}

func (app *application) checkin(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Method not allowed",
		})
		return
	}

	_, client_id, _, err := homeHandler(w, r)
	if err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Authorization required",
		})
		return
	}

	err = app.quests.Check(client_id)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	soul := app.quests.SoulCheck(client_id)
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Ежедневная награда получена! +5 душ",
		"souls":   soul,
	})
}

func (app *application) log(w http.ResponseWriter, r *http.Request) {

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	s, err := app.quests.Listlog()

	if err != nil {
		return
	}

	app.render(w, r, "logs.page.tmpl", &templateData{
		Log: s,
	})

}

func (app *application) createitem(w http.ResponseWriter, r *http.Request) {

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.Method != "POST" {
		app.render(w, r, "createitem.page.tmpl", &templateData{})
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	price, err := strconv.Atoi(r.FormValue("price"))

	if err != nil {
		app.render(w, r, "createitem.page.tmpl", &templateData{
			Error: "В цене не должно быть ничего кроме цифр",
		})
		return
	}
	count, err := strconv.Atoi(r.FormValue("count"))

	if err != nil {
		app.render(w, r, "createitem.page.tmpl", &templateData{
			Error: "В количестве не должно быть ничего кроме цифр",
		})
		return
	}

	err = app.quests.CreateI(title, description, price, count)
	if err != nil {
		app.render(w, r, "createitem.page.tmpl", &templateData{
			Error: "Предмет не добавлен",
		})
		return
	}

	app.render(w, r, "createitem.page.tmpl", &templateData{
		Error: "Предмет успешно добавлен",
	})
}

func (app *application) helpclient(w http.ResponseWriter, r *http.Request) {

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	if r.Method != "POST" {
		app.render(w, r, "helpclient.page.tmpl", &templateData{})
		return
	}

	phone := r.FormValue("phone")
	pass := r.FormValue("password")

	passHash, err := bcrypt.GenerateFromPassword([]byte(pass), bcrypt.DefaultCost)
	if err != nil {
		app.serverError(w, err)
		return
	}

	err = app.quests.AdminPass(phone, string(passHash))
	if err != nil {
		app.render(w, r, "helpclient.page.tmpl", &templateData{
			Error: "Пользователь не найден",
		})
		return
	}

	app.render(w, r, "helpclient.page.tmpl", &templateData{
		Error: "Пароль изменен",
	})

}

func UpdateRandomQuestsStatus(db *sql.DB) error {
	// Начинаем транзакцию
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %v", err)
	}
	defer tx.Rollback() // В случае ошибки откатываем

	// Устанавливаем всем заданиям active = false
	_, err = tx.Exec("UPDATE quests SET active = false")
	if err != nil {
		return fmt.Errorf("failed to deactivate all quests: %v", err)
	}

	// Выбираем 3 случайных задания
	rows, err := tx.Query(`
        SELECT quest_id FROM quests 
		WHERE quest_id != 10 AND delet = false
        ORDER BY random() 
        LIMIT 3
        FOR UPDATE`)
	if err != nil {
		return fmt.Errorf("failed to select random quests: %v", err)
	}
	defer rows.Close()

	var ids []int
	for rows.Next() {
		var id int
		if err := rows.Scan(&id); err != nil {
			return fmt.Errorf("failed to scan quest id: %v", err)
		}
		ids = append(ids, id)
	}

	if len(ids) == 0 {
		return nil // Нет заданий для обновления
	}

	// Обновляем выбранные задания
	_, err = tx.Exec(`
        UPDATE quests 
        SET active = true 
        WHERE quest_id = ANY($1)`, pq.Array(ids))
	if err != nil {
		return fmt.Errorf("failed to activate random quests: %v", err)
	}

	// Фиксируем транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %v", err)
	}

	return nil
}

func (app *application) takeItems(w http.ResponseWriter, r *http.Request) {
	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	history, err := app.quests.ListPuch()
	if err != nil {
		return
	}

	app.render(w, r, "takeitems.page.tmpl", &templateData{
		History: history,
	})

}

func (app *application) completeitem(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Method not allowed",
		})
		return
	}

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Authorization required",
		})
		return
	}

	var request struct {
		Purc_id int `json:"purc_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = app.quests.ComplItem(request.Purc_id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Награда выдана",
	})

}

func (app *application) logitem(w http.ResponseWriter, r *http.Request) {

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	history, err := app.quests.Listlogitem()
	if err != nil {
		return
	}

	app.render(w, r, "logsitem.page.tmpl", &templateData{
		History: history,
	})

}

func (app *application) deleteitem(w http.ResponseWriter, r *http.Request) {

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	shops, err := app.quests.Products()
	if err != nil {
		return
	}

	app.render(w, r, "deleteitem.page.tmpl", &templateData{
		Shop: shops,
	})
}

func (app *application) delitem(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Method not allowed",
		})
		return
	}

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Authorization required",
		})
		return
	}

	var request struct {
		Shop_id int `json:"shop_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = app.quests.DeleteItem(request.Shop_id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Предмет удален",
	})

}

func (app *application) deletequest(w http.ResponseWriter, r *http.Request) {

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	s, err := app.quests.AllQuests()
	if err != nil {
		app.serverError(w, err)
		return
	}

	app.render(w, r, "deletequest.page.tmpl", &templateData{
		Quests: s,
	})

}

func (app *application) delquest(w http.ResponseWriter, r *http.Request) {

	if r.Method != "POST" {
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Method not allowed",
		})
		return
	}

	_, _, dost, err := homeHandler(w, r)
	if err != nil || dost < 1 {
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": "Authorization required",
		})
		return
	}

	var request struct {
		Quest_id int `json:"quest_id"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, "Invalid request", http.StatusBadRequest)
		return
	}

	err = app.quests.DeleteQuest(request.Quest_id)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]string{
			"error": err.Error(),
		})
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Квест удален",
	})

}
