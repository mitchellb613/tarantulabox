package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type TarantulaModelInterface interface {
	Insert(species string, name string, feed_interval_days int, notify bool, img_url string, owner_id int) (int, error)
	Get(id int) (*Tarantula, error)
}

type Tarantula struct {
	ID                 int
	Species            string
	Name               string
	Feed_Interval_Days int
	Notify             bool
	Img_URL            string
	Created            time.Time
	Owner_ID           int
}

type TarantulaModel struct {
	DB *pgxpool.Pool
}

func (m *TarantulaModel) Insert(species string, name string, feed_interval_days int, notify bool, img_url string, owner_id int) (int, error) {

	stmt := `
			INSERT INTO tarantulas (species, name, feed_interval_days, notify, img_url, owner_id)
			VALUES($1, $2, $3, $4, $5, $6)
			RETURNING id
			`

	var id int
	err := m.DB.QueryRow(context.Background(), stmt, species, name, feed_interval_days, notify, img_url, owner_id).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *TarantulaModel) Get(id int) (*Tarantula, error) {
	stmt := `SELECT id, species, name, feed_interval_days, notify, img_url, owner_id FROM tarantulas
	WHERE id = $1`

	row := m.DB.QueryRow(context.Background(), stmt, id)

	s := &Tarantula{}

	err := row.Scan(&s.ID, &s.Species, &s.Name, &s.Feed_Interval_Days, &s.Notify, &s.Img_URL, &s.Owner_ID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}
