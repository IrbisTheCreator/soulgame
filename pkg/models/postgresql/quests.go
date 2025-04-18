package postgresql

import (
	"database/sql"
	"errors"
	"soulgame/pkg/models"
)

type QuestModel struct {
	DB *sql.DB
}

func (m *QuestModel) Insert(title, description string, price int) error {

	_, err := m.DB.Exec("INSERT INTO quests (title, description, price) VALUES ($1,$2,$3)", title, description, price)
	if err != nil {
		return err
	}

	return nil
}

func (m *QuestModel) Get(quest_id int) (*models.Quests, error) {

	s := &models.Quests{}
	err := m.DB.QueryRow("select * FROM quests WHERE quest_id = $1", quest_id).Scan(&s.Quest_id, &s.Title, &s.Description, &s.Price, &s.Active)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, models.ErrNoRecord
		} else {
			return nil, err
		}
	}
	return s, nil
}

func (m *QuestModel) Latest() ([]*models.Quests, error) {
	tru := true
	rows, err := m.DB.Query("SELECT * FROM quests WHERE active = $1", tru)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quests []*models.Quests
	for rows.Next() {
		s := &models.Quests{}
		err = rows.Scan(&s.Quest_id, &s.Title, &s.Description, &s.Price, &s.Active, &s.Exp, &s.Delet)
		if err != nil {
			return nil, err
		}
		quests = append(quests, s)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return quests, nil

}

func (m *QuestModel) Regist(phone string, pass string) error {

	_, err := m.DB.Exec("INSERT INTO users (phone, pass) VALUES ($1,$2)", phone, pass)
	if err != nil {
		return err
	}

	return nil
}

func (m *QuestModel) Take(quest_id, user_id, per int) error {

	if per == 1 {
		_, err := m.DB.Exec("UPDATE taken SET status = 2 WHERE quest_id = $1 AND user_id = $2 AND dat = CURRENT_DATE", quest_id, user_id)
		if err != nil {
			return err
		}
	} else {
		_, err := m.DB.Exec("INSERT INTO taken (user_id, quest_id) VALUES ($1,$2)", user_id, quest_id)
		if err != nil {
			return err
		}
	}

	return nil
}

func (m *QuestModel) Checking(quest_id, user_id int) (int, error) {
	var res int
	err := m.DB.QueryRow("select status FROM taken WHERE quest_id = $1 AND user_id = $2 AND dat = CURRENT_DATE", quest_id, user_id).Scan(&res)
	if err != nil {
		if err == sql.ErrNoRows {
			// Если запись не найдена, возвращаем статус 0 и nil ошибку
			return 0, nil
		}
		return 0, err
	}

	return res, nil
}

func (m *QuestModel) SoulCheck(user_id int) int {
	var res int
	err := m.DB.QueryRow("select balance FROM users WHERE user_id = $1", user_id).Scan(&res)
	if err != nil {
		return 0
	}
	return res
}

func (m *QuestModel) Products() ([]*models.Shop, error) {

	rows, err := m.DB.Query("SELECT id_shop, title, description, count, price FROM shop WHERE count != $1", 0)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quests []*models.Shop
	for rows.Next() {
		s := &models.Shop{}

		err = rows.Scan(&s.Shop_id, &s.Title, &s.Description, &s.Count, &s.Price)
		if err != nil {
			return nil, err
		}
		quests = append(quests, s)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return quests, nil
}

func (m *QuestModel) ListCl() ([]*models.Adm, error) {

	rows, err := m.DB.Query("SELECT t.take_id, u.phone, q.title, t.status, t.user_id, q.price, q.expiriance, t.quest_id FROM taken t JOIN users u ON t.user_id = u.user_id JOIN quests q ON t.quest_id = q.quest_id WHERE t.dat = CURRENT_DATE AND status = 2  ORDER BY t.take_id;")

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var adm []*models.Adm
	for rows.Next() {
		s := &models.Adm{}
		err = rows.Scan(&s.Take_id, &s.Phone, &s.Title, &s.Now, &s.User_id, &s.Price, &s.Exp, &s.Quest_id)
		if err != nil {
			return nil, err
		}
		adm = append(adm, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return adm, nil

}

func (m *QuestModel) CreateQ(title, description string, price, exp int) error {
	_, err := m.DB.Exec("INSERT INTO quests (title, description, price, expiriance) VALUES ($1,$2,$3,$4)", title, description, price, exp)
	if err != nil {
		return err
	}
	return nil
}

func (m *QuestModel) Info(user_id int) (*models.User, error) {
	s := &models.User{}

	err := m.DB.QueryRow("select level, exp, balance FROM users WHERE user_id = $1", user_id).Scan(&s.Level, &s.Exp, &s.Balance)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func (m *QuestModel) Buying(user_id, item_id, club_id int) (bool, error) {
	tx, err := m.DB.Begin()
	if err != nil {
		return false, err
	}
	defer tx.Rollback()
	var userSouls int
	err = tx.QueryRow("SELECT balance FROM users WHERE user_id = $1 FOR UPDATE", user_id).Scan(&userSouls)

	if err != nil {
		return false, err
	}

	var itemPrice int
	var itemCount int
	err = tx.QueryRow("SELECT price, count FROM shop WHERE id_shop = $1 FOR UPDATE", item_id).Scan(&itemPrice, &itemCount)

	if err != nil {
		return false, err
	}

	if userSouls < itemPrice {
		return false, nil
	}

	if itemCount != -1 && itemCount <= 0 {
		return false, nil
	}

	_, err = tx.Exec("UPDATE users SET balance = balance - $1 WHERE user_id = $2",
		itemPrice, user_id)

	if err != nil {
		return false, err
	}

	if itemCount != -1 {
		_, err = tx.Exec("UPDATE shop SET count = count - 1 WHERE id_shop = $1", item_id)
		if err != nil {
			return false, err
		}
	}

	_, err = tx.Exec(`INSERT INTO purchases (user_id, item_id, club_id, status) 
         VALUES ($1, $2, $3, FALSE)`,
		user_id, item_id, club_id)

	if err != nil {
		return false, err
	}

	if err := tx.Commit(); err != nil {
		return false, nil
	}

	return true, nil
}

func (m *QuestModel) HistoryB(user_id int) ([]*models.HistoryBought, error) {

	rows, err := m.DB.Query("SELECT p.time, s.title, p.status, p.club_id FROM purchases p JOIN shop s ON p.item_id = s.id_shop WHERE p.user_id = $1 ORDER BY p.time DESC LIMIT 10", user_id)

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var his []*models.HistoryBought
	for rows.Next() {
		s := &models.HistoryBought{}
		err = rows.Scan(&s.PurchaseDate, &s.ItemTitle, &s.IsIssued, &s.ClubName)
		if err != nil {
			return nil, err
		}
		his = append(his, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return his, nil

}

func (m *QuestModel) UpdatePass(user_id int, pass string) error {
	_, err := m.DB.Exec("UPDATE users SET pass = $1 WHERE user_id = $2",
		pass, user_id)

	if err != nil {
		return err
	}
	return nil
}

func (m *QuestModel) Compl(take_id, user_id, price, exp, quest_id int) error {

	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("UPDATE taken SET status = 3 WHERE take_id = $1",
		take_id)

	if err != nil {
		return err
	}

	var userExp int
	err = tx.QueryRow("SELECT exp FROM users WHERE user_id = $1 FOR UPDATE", user_id).Scan(&userExp)

	if err != nil {
		return err
	}

	userExp += exp
	level := 0
	if userExp >= 100 {
		level = 1
		userExp = userExp % 100
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1, level = level + $2, exp = $3  WHERE user_id = $4",
		price, level, userExp, user_id)

	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO log (client_id, quest_id) VALUES ($1,$2)", user_id, quest_id)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *QuestModel) Check(user_id int) error {
	tx, err := m.DB.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	_, err = tx.Exec("INSERT INTO taken (user_id, quest_id, status) VALUES ($1,$2,$3)", user_id, 10, 3)
	if err != nil {
		return err
	}

	var userExp int
	err = tx.QueryRow("SELECT exp FROM users WHERE user_id = $1 FOR UPDATE", user_id).Scan(&userExp)

	if err != nil {
		return err
	}

	userExp += 5
	level := 0
	if userExp >= 100 {
		level = 1
		userExp = userExp % 100
	}

	_, err = tx.Exec("UPDATE users SET balance = balance + $1, level = level + $2, exp = $3  WHERE user_id = $4",
		5, level, userExp, user_id)

	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT INTO log (client_id, quest_id) VALUES ($1,$2)", user_id, 10)
	if err != nil {
		return err
	}

	if err := tx.Commit(); err != nil {
		return err
	}

	return nil
}

func (m *QuestModel) Listlog() ([]*models.Log, error) {

	rows, err := m.DB.Query("SELECT l.log_id, u.phone, q.title, l.times FROM log l JOIN users u ON l.client_id = u.user_id JOIN quests q ON l.quest_id = q.quest_id ORDER BY log_id desc LIMIT 100")

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var his []*models.Log
	for rows.Next() {
		s := &models.Log{}
		err = rows.Scan(&s.Log_id, &s.Phone, &s.Title, &s.Time)
		if err != nil {
			return nil, err
		}
		his = append(his, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return his, nil

}

func (m *QuestModel) CreateI(title, description string, price, count int) error {
	_, err := m.DB.Exec("INSERT INTO shop (title, description, price, count) VALUES ($1,$2,$3,$4)", title, description, price, count)
	if err != nil {
		return err
	}
	return nil
}

func (m *QuestModel) AdminPass(phone, pass string) error {

	var res int
	err := m.DB.QueryRow("select user_id FROM users WHERE phone = $1", phone).Scan(&res)
	if err != nil {
		return err
	}

	_, err = m.DB.Exec("UPDATE users SET pass = $1 WHERE user_id = $2",
		pass, res)

	if err != nil {
		return err
	}
	return nil
}

func (m *QuestModel) ListPuch() ([]*models.HistoryBought, error) {
	rows, err := m.DB.Query("SELECT p.purc_id, p.time, s.title, p.club_id, u.phone  FROM purchases p JOIN shop s ON p.item_id = s.id_shop JOIN users u ON p.user_id = u.user_id WHERE status = false ORDER BY p.time DESC")

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var his []*models.HistoryBought
	for rows.Next() {
		s := &models.HistoryBought{}
		err = rows.Scan(&s.Purc_id, &s.PurchaseDate, &s.ItemTitle, &s.ClubName, &s.Phone)
		if err != nil {
			return nil, err
		}
		his = append(his, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return his, nil
}

func (m *QuestModel) ComplItem(purc_id int) error {

	_, err := m.DB.Exec("UPDATE purchases SET status = true WHERE purc_id = $1",
		purc_id)

	if err != nil {
		return err
	}

	return nil

}

func (m *QuestModel) Listlogitem() ([]*models.HistoryBought, error) {

	rows, err := m.DB.Query("SELECT p.purc_id, p.time, s.title, p.club_id, u.phone  FROM purchases p JOIN shop s ON p.item_id = s.id_shop JOIN users u ON p.user_id = u.user_id WHERE status = true ORDER BY p.time DESC LIMIT 100")

	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var his []*models.HistoryBought
	for rows.Next() {
		s := &models.HistoryBought{}
		err = rows.Scan(&s.Purc_id, &s.PurchaseDate, &s.ItemTitle, &s.ClubName, &s.Phone)
		if err != nil {
			return nil, err
		}
		his = append(his, s)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return his, nil

}

func (m *QuestModel) DeleteItem(shop_id int) error {

	_, err := m.DB.Exec("UPDATE shop SET count = 0 WHERE id_shop = $1",
		shop_id)

	if err != nil {
		return err
	}

	return nil

}

func (m *QuestModel) AllQuests() ([]*models.Quests, error) {
	rows, err := m.DB.Query("SELECT * FROM quests WHERE delet = false")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quests []*models.Quests
	for rows.Next() {
		s := &models.Quests{}
		err = rows.Scan(&s.Quest_id, &s.Title, &s.Description, &s.Price, &s.Active, &s.Exp, &s.Delet)
		if err != nil {
			return nil, err
		}
		quests = append(quests, s)

	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return quests, nil

}

func (m *QuestModel) DeleteQuest(quest_id int) error {

	_, err := m.DB.Exec("UPDATE quests SET delet = true WHERE quest_id = $1",
		quest_id)

	if err != nil {
		return err
	}

	return nil

}
