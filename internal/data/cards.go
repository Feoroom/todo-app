package data

import (
	"context"
	"database/sql"
	"errors"
	"github.com/lib/pq"
	"library/internal/validation"
	"time"
)

type Card struct {
	ID        int64     `json:"id"`
	Title     string    `json:"title"`
	Events    Events    `json:"events"`
	CreatedAt time.Time `json:"-"`
}

type CardModel struct {
	DB *sql.DB
}

func ValidateCard(v *validation.Validator, card *Card) {

	v.Check(card.Title != "", "title", "must be provided")
	v.Check(len(card.Title) <= 200, "title", "must be less than 200 bytes long")

}

func (c CardModel) Get(id int64) (*Card, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}
	//TODO Добавить пагинацию для events
	q := `
		select cards.id,
			   cards.title,
			   cards.created_at,
			   coalesce(
							   array_agg(row_to_json(events.*))
							   filter ( where events.id is not null ),
							   '{}'
			   ) as events
		
		from cards
				 LEFT JOIN events
						   ON cards.id = events.card_id
		WHERE cards.id = $1
		GROUP BY cards.id, cards.title, cards.created_at;`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	var card Card

	err := c.DB.QueryRowContext(ctx, q, id).Scan(
		&card.ID,
		&card.Title,
		&card.CreatedAt,
		pq.Array(&card.Events),
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &card, nil
}

func (c CardModel) Insert(card *Card) error {

	q := `insert into cards (title) 
		values ($1)
		returning id, created_at`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return c.DB.QueryRowContext(ctx, q, card.Title).Scan(&card.ID, &card.CreatedAt)
}

func (c CardModel) Update (card *Card) error {
	q := `update cards
		set title=$1
		where id=$2`

	ctx, cancel := context.WithTimeout(context.Background(), 3 * time.Second)
	defer cancel()

	_, err := c.DB.ExecContext(ctx, q, card.Title, card.ID)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return ErrEditConflict
		default:
			return err
		}
	}

	return nil
}
