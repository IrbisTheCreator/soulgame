package postgresql

import (
	"database/sql"
	"errors"
	"soulgame/pkg/models"
)

var ErrNoUser = errors.New("models: неверный пароль")

func (m *QuestModel) Verif(phone string) (*models.User, error) {

	s := &models.User{}

	err := m.DB.QueryRow("select user_id,phone,pass,dostup FROM users WHERE phone = $1", phone).Scan(&s.User_id, &s.Phone, &s.Pass, &s.Dostup)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoUser
		} else {
			return nil, err
		}
	}
	return s, nil

}
