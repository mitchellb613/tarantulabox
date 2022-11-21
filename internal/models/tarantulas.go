package models

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/jackc/pgx/v4/pgxpool"
)

type TarantulaModelInterface interface {
	Insert(species string, name string, feed_interval_days int, notify bool, img_url string, owner_id int, next_feed_date time.Time) (int, error)
	Get(id int) (*Tarantula, error)
	GetNotifications(n int) ([]*Notification, error)
	UpdateNextFeedDate(n *Notification) error
}

type Tarantula struct {
	ID                 int
	Species            string
	Name               string
	Next_Feed_Date     time.Time
	Feed_Interval_Days int
	Notify             bool
	Img_URL            string
	Created            time.Time
	Owner_ID           int
}

type TarantulaModel struct {
	DB *pgxpool.Pool
}

func (m *TarantulaModel) Insert(species string, name string, feed_interval_days int, notify bool, img_url string, owner_id int, next_feed_date time.Time) (int, error) {

	stmt := `
			INSERT INTO tarantulas (species, name, feed_interval_days, notify, img_url, owner_id, next_feed_date)
			VALUES($1, $2, $3, $4, $5, $6, $7)
			RETURNING id
			`

	var id int
	err := m.DB.QueryRow(context.Background(), stmt, species, name, feed_interval_days, notify, img_url, owner_id, next_feed_date).Scan(&id)
	if err != nil {
		return 0, err
	}

	return id, nil
}

func (m *TarantulaModel) Get(id int) (*Tarantula, error) {
	stmt := `SELECT id, species, name, feed_interval_days, notify, img_url, owner_id, next_feed_date FROM tarantulas
	WHERE id = $1`

	row := m.DB.QueryRow(context.Background(), stmt, id)

	s := &Tarantula{}

	err := row.Scan(&s.ID, &s.Species, &s.Name, &s.Feed_Interval_Days, &s.Notify, &s.Img_URL, &s.Owner_ID, &s.Next_Feed_Date)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNoRecord
		} else {
			return nil, err
		}
	}

	return s, nil
}

type Notification struct {
	Tarantula_ID       int
	NotifyTime         time.Time
	Feed_Interval_Days int
}

func (m *TarantulaModel) GetNotifications(n int) ([]*Notification, error) {
	stmt := `select id, next_feed_date, feed_interval_days
	from tarantulas
	order by next_feed_date
	limit $1
	`
	rows, err := m.DB.Query(context.Background(), stmt, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	notifications := []*Notification{}

	for rows.Next() {
		n := &Notification{}

		err = rows.Scan(&n.Tarantula_ID, &n.NotifyTime, &n.Feed_Interval_Days)
		if err != nil {
			return nil, err
		}

		notifications = append(notifications, n)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return notifications, nil
}

func (m *TarantulaModel) UpdateNextFeedDate(n *Notification) error {
	stmt := `update tarantulas
	set next_feed_date = $1
	where id = $2
	`
	_, err := m.DB.Exec(context.Background(), stmt, n.NotifyTime.AddDate(0, 0, n.Feed_Interval_Days), n.Tarantula_ID)
	if err != nil {
		return err
	}
	return nil
}
