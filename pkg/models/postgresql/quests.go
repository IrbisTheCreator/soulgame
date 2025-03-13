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
	err := m.DB.QueryRow("select * FROM quests WHERE quest_id = $1", quest_id).Scan(&s.Quest_id, &s.Title, &s.Description, &s.Price)
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
	rows, err := m.DB.Query("SELECT * FROM quests ORDER  BY quest_id DESC LIMIT 3")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var quests []*models.Quests
	for rows.Next() {
		s := &models.Quests{}

		err = rows.Scan(&s.Quest_id, &s.Title, &s.Description, &s.Price)
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
