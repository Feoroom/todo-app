package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lib/pq"
	"library/internal/validation"
	"time"
)

type Event struct {
	ID          int64     `json:"id" `
	CreatedAt   time.Time `json:"-"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	TextBlocks  []string  `json:"text_blocks,omitempty"`
	Date        Date      `json:"date,omitempty"`
	Version     int64     `json:"version,omitempty"`
	CardId      int64     `json:"card_id"`
}

type EventModel struct {
	DB *sql.DB
}

type Events []Event

func ValidateEvent(v *validation.Validator, event *Event) {
	//Title validation
	v.Check(event.Title != "", "title", "must be provided")
	v.Check(len(event.Title) <= 200, "title", "must be less than 200 bytes long")

	//Description validation
	v.Check(event.Description != "", "description", "must be provided")
	v.Check(len(event.Description) <= 1000, "description", "must not be less than 200 bytes long")

	//TextBlocks validation
	v.Check(event.TextBlocks != nil, "text_blocks", "must be provided")
	v.Check(len(event.TextBlocks) >= 1, "text_blocks", "must contain at least 1 element")
	v.Check(len(event.TextBlocks) <= 10, "text_blocks", "must not contain more than 10 elements")
	v.Check(!validation.In("", event.TextBlocks...), "text_blocks", "must not contain empty elements")

	//Date validation
	v.Check(!event.Date.Time.IsZero(), "date", "must be provided")
	v.Check(event.Date.CheckYear(), "date", "incorrect year")
	v.Check(event.Date.CheckMonth(), "date", "incorrect month")
	v.Check(event.Date.CheckDay(), "date", "incorrect day")

	//CardId validation
	v.Check(event.CardId != 0, "card_id", "must be provided")
	// Если можно, добавить проверку на то, что такое card_id есть в базе
}

func (e EventModel) Insert(event *Event) error {
	q := `insert into events (title, description, text_blocks, date, card_id)
			values ($1, $2, $3, $4, $5)
			returning id, created_at, version`
	args := []interface{}{event.Title, event.Description, pq.Array(event.TextBlocks), event.Date.Time, event.CardId}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	return e.DB.QueryRowContext(ctx, q, args...).Scan(&event.ID, &event.CreatedAt, &event.Version)
}

func (e EventModel) Get(id int64) (*Event, error) {

	if id < 1 {
		return nil, ErrRecordNotFound
	}

	q := `select id, created_at, title, description, text_blocks, date, version, card_id 
			from events
			where id=$1`

	var event Event

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, q, id).Scan(
		&event.ID,
		&event.CreatedAt,
		&event.Title,
		&event.Description,
		pq.Array(&event.TextBlocks),
		&event.Date.Time,
		&event.Version,
		&event.CardId,
	)

	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, ErrRecordNotFound
		default:
			return nil, err
		}
	}

	return &event, nil
}

func (e EventModel) GetAll(title string, date Date, filters Filters) ([]*Event, Metadata, error) {

	//TODO: нужно будет пофиксить баг связанный с фильтрацией по дате
	// если выбрана такая дата date = current_date, то выводятся все элементы
	// возможно просто нужно задать другое значение по умолчанию
	q := fmt.Sprintf(`
		select count(*) over(), id, created_at, title, description, text_blocks, date, version, card_id
		from events
		where (to_tsvector('simple', title) @@ plainto_tsquery('simple', $1) or $1 = '')
		and (date = $2 or $2 = current_date)
		order by %s %s, id ASC
		limit $3 offset $4 `, filters.sortColumn(), filters.sortDirection())

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	args := []interface{}{title, date.Time, filters.limit(), filters.offset()}

	rows, err := e.DB.QueryContext(ctx, q, args...)
	if err != nil {
		return nil, Metadata{}, err
	}
	defer rows.Close()

	totalRecords := 0
	var events []*Event

	for rows.Next() {
		var event Event

		err := rows.Scan(
			&totalRecords,
			&event.ID,
			&event.CreatedAt,
			&event.Title,
			&event.Description,
			pq.Array(&event.TextBlocks),
			&event.Date.Time,
			&event.Version,
			&event.CardId,
		)
		if err != nil {
			return nil, Metadata{}, err
		}

		events = append(events, &event)
	}

	metadata := calculateMetadata(totalRecords, filters.Page, filters.PageSize)

	return events, metadata, nil
}

func (e EventModel) Update(event *Event) error {
	q := `update events
		set title=$1, description=$2, date=$3, text_blocks=$4, version = version + 1
		where id=$5 and version=$6
		returning version`

	args := []interface{}{
		event.Title,
		event.Description,
		event.Date.Time,
		pq.Array(event.TextBlocks),
		event.ID,
		event.Version,
	}

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	err := e.DB.QueryRowContext(ctx, q, args...).Scan(&event.Version)
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

func (e EventModel) Delete(id int64) error {
	if id < 1 {
		return ErrRecordNotFound
	}

	q := `delete from events
			where id=$1`
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	res, err := e.DB.ExecContext(ctx, q, id)
	if err != nil {
		return err
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return err
	}

	if rowsAffected == 0 {
		return ErrRecordNotFound
	}

	return nil
}

func (e *Event) Scan(src any) error {
	var data []byte
	switch v := src.(type) {
	case string:
		data = []byte(v)
	case []byte:
		data = v
	}

	return json.Unmarshal(data, e)
}

func (e *Events) Scan(src any) error {
	return pq.GenericArray{A: e}.Scan(src)
}
