package main

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound   = errors.New("not found")
	ErrConflict   = errors.New("conflict")          // нарушен UNIQUE (23505)
	ErrInvalidRef = errors.New("invalid reference") // нарушен внешний ключ (23503)
)

// mapErr переводит ошибки драйвера и Postgres в ошибки слоя хранения,
// чтобы хендлеры не зависели от pgx.
func mapErr(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505":
			return ErrConflict
		case "23503":
			return ErrInvalidRef
		}
	}
	return err
}

type Store struct {
	DB *pgxpool.Pool
}

func NewStore(db *pgxpool.Pool) *Store {
	return &Store{DB: db}
}

// users

func (s *Store) GetUsers(ctx context.Context) ([]User, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, email, created_at
		FROM users
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := []User{}

	for rows.Next() {
		var u User

		if err := rows.Scan(
			&u.ID,
			&u.Name,
			&u.Email,
			&u.CreatedAt,
		); err != nil {
			return nil, err
		}

		users = append(users, u)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return users, nil
}

func (s *Store) GetUser(ctx context.Context, id int64) (User, error) {
	var u User

	err := s.DB.QueryRow(ctx, `
		SELECT id, name, email, created_at
		FROM users
		WHERE id = $1
	`, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
	)

	return u, mapErr(err)
}

func (s *Store) CreateUser(ctx context.Context, u User) (User, error) {
	err := s.DB.QueryRow(ctx, `
		INSERT INTO users (name, email)
		VALUES ($1, $2)
		RETURNING id, name, email, created_at
	`, u.Name, u.Email).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
	)

	return u, mapErr(err)
}

func (s *Store) UpdateUser(ctx context.Context, u User) (User, error) {
	err := s.DB.QueryRow(ctx, `
		UPDATE users
		SET name = $1, email = $2
		WHERE id = $3
		RETURNING id, name, email, created_at
	`, u.Name, u.Email, u.ID).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.CreatedAt,
	)

	return u, mapErr(err)
}

func (s *Store) DeleteUser(ctx context.Context, id int64) error {
	result, err := s.DB.Exec(ctx, `
		DELETE FROM users
		WHERE id = $1
	`, id)

	if err != nil {
		return mapErr(err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// categories

func (s *Store) GetCategories(ctx context.Context) ([]Category, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, name, created_at
		FROM categories
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	categories := []Category{}

	for rows.Next() {
		var c Category

		if err := rows.Scan(
			&c.ID,
			&c.Name,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}

		categories = append(categories, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return categories, nil
}

func (s *Store) GetCategory(ctx context.Context, id int64) (Category, error) {
	var c Category

	err := s.DB.QueryRow(ctx, `
		SELECT id, name, created_at
		FROM categories
		WHERE id = $1
	`, id).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)

	return c, mapErr(err)
}

func (s *Store) CreateCategory(ctx context.Context, c Category) (Category, error) {
	err := s.DB.QueryRow(ctx, `
		INSERT INTO categories (name)
		VALUES ($1)
		RETURNING id, name, created_at
	`, c.Name).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)

	return c, mapErr(err)
}

func (s *Store) UpdateCategory(ctx context.Context, c Category) (Category, error) {
	err := s.DB.QueryRow(ctx, `
		UPDATE categories
		SET name = $1
		WHERE id = $2
		RETURNING id, name, created_at
	`, c.Name, c.ID).Scan(
		&c.ID,
		&c.Name,
		&c.CreatedAt,
	)

	return c, mapErr(err)
}

func (s *Store) DeleteCategory(ctx context.Context, id int64) error {
	result, err := s.DB.Exec(ctx, `
		DELETE FROM categories
		WHERE id = $1
	`, id)

	if err != nil {
		return mapErr(err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// tickets

func (s *Store) GetTickets(ctx context.Context) ([]Ticket, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, title, description, status, user_id, category_id, created_at
		FROM tickets
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	tickets := []Ticket{}

	for rows.Next() {
		var t Ticket

		if err := rows.Scan(
			&t.ID,
			&t.Title,
			&t.Description,
			&t.Status,
			&t.UserID,
			&t.CategoryID,
			&t.CreatedAt,
		); err != nil {
			return nil, err
		}

		tickets = append(tickets, t)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return tickets, nil
}

func (s *Store) GetTicket(ctx context.Context, id int64) (Ticket, error) {
	var t Ticket

	err := s.DB.QueryRow(ctx, `
		SELECT id, title, description, status, user_id, category_id, created_at
		FROM tickets
		WHERE id = $1
	`, id).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.UserID,
		&t.CategoryID,
		&t.CreatedAt,
	)

	return t, mapErr(err)
}

func (s *Store) CreateTicket(ctx context.Context, t Ticket) (Ticket, error) {
	err := s.DB.QueryRow(ctx, `
		INSERT INTO tickets (
			title,
			description,
			status,
			user_id,
			category_id
		)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING
			id,
			title,
			description,
			status,
			user_id,
			category_id,
			created_at
	`,
		t.Title,
		t.Description,
		t.Status,
		t.UserID,
		t.CategoryID,
	).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.UserID,
		&t.CategoryID,
		&t.CreatedAt,
	)

	return t, mapErr(err)
}

func (s *Store) UpdateTicket(ctx context.Context, t Ticket) (Ticket, error) {
	err := s.DB.QueryRow(ctx, `
		UPDATE tickets
		SET
			title = $1,
			description = $2,
			status = $3,
			user_id = $4,
			category_id = $5
		WHERE id = $6
		RETURNING
			id,
			title,
			description,
			status,
			user_id,
			category_id,
			created_at
	`,
		t.Title,
		t.Description,
		t.Status,
		t.UserID,
		t.CategoryID,
		t.ID,
	).Scan(
		&t.ID,
		&t.Title,
		&t.Description,
		&t.Status,
		&t.UserID,
		&t.CategoryID,
		&t.CreatedAt,
	)

	return t, mapErr(err)
}

func (s *Store) DeleteTicket(ctx context.Context, id int64) error {
	result, err := s.DB.Exec(ctx, `
		DELETE FROM tickets
		WHERE id = $1
	`, id)

	if err != nil {
		return mapErr(err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// comments

func (s *Store) GetComments(ctx context.Context) ([]Comment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, ticket_id, user_id, text, created_at
		FROM comments
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	comments := []Comment{}

	for rows.Next() {
		var c Comment

		if err := rows.Scan(
			&c.ID,
			&c.TicketID,
			&c.UserID,
			&c.Text,
			&c.CreatedAt,
		); err != nil {
			return nil, err
		}

		comments = append(comments, c)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return comments, nil
}

func (s *Store) GetComment(ctx context.Context, id int64) (Comment, error) {
	var c Comment

	err := s.DB.QueryRow(ctx, `
		SELECT id, ticket_id, user_id, text, created_at
		FROM comments
		WHERE id = $1
	`, id).Scan(
		&c.ID,
		&c.TicketID,
		&c.UserID,
		&c.Text,
		&c.CreatedAt,
	)

	return c, mapErr(err)
}

func (s *Store) CreateComment(ctx context.Context, c Comment) (Comment, error) {
	err := s.DB.QueryRow(ctx, `
		INSERT INTO comments (ticket_id, user_id, text)
		VALUES ($1, $2, $3)
		RETURNING id, ticket_id, user_id, text, created_at
	`,
		c.TicketID,
		c.UserID,
		c.Text,
	).Scan(
		&c.ID,
		&c.TicketID,
		&c.UserID,
		&c.Text,
		&c.CreatedAt,
	)

	return c, mapErr(err)
}

func (s *Store) UpdateComment(ctx context.Context, c Comment) (Comment, error) {
	err := s.DB.QueryRow(ctx, `
		UPDATE comments
		SET
			ticket_id = $1,
			user_id = $2,
			text = $3
		WHERE id = $4
		RETURNING id, ticket_id, user_id, text, created_at
	`,
		c.TicketID,
		c.UserID,
		c.Text,
		c.ID,
	).Scan(
		&c.ID,
		&c.TicketID,
		&c.UserID,
		&c.Text,
		&c.CreatedAt,
	)

	return c, mapErr(err)
}

func (s *Store) DeleteComment(ctx context.Context, id int64) error {
	result, err := s.DB.Exec(ctx, `
		DELETE FROM comments
		WHERE id = $1
	`, id)

	if err != nil {
		return mapErr(err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// attachments

func (s *Store) GetAttachments(ctx context.Context) ([]Attachment, error) {
	rows, err := s.DB.Query(ctx, `
		SELECT id, ticket_id, file_name, file_url, created_at
		FROM attachments
		ORDER BY id
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	attachments := []Attachment{}

	for rows.Next() {
		var a Attachment

		if err := rows.Scan(
			&a.ID,
			&a.TicketID,
			&a.FileName,
			&a.FileURL,
			&a.CreatedAt,
		); err != nil {
			return nil, err
		}

		attachments = append(attachments, a)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return attachments, nil
}

func (s *Store) GetAttachment(ctx context.Context, id int64) (Attachment, error) {
	var a Attachment

	err := s.DB.QueryRow(ctx, `
		SELECT id, ticket_id, file_name, file_url, created_at
		FROM attachments
		WHERE id = $1
	`, id).Scan(
		&a.ID,
		&a.TicketID,
		&a.FileName,
		&a.FileURL,
		&a.CreatedAt,
	)

	return a, mapErr(err)
}

func (s *Store) CreateAttachment(ctx context.Context, a Attachment) (Attachment, error) {
	err := s.DB.QueryRow(ctx, `
		INSERT INTO attachments (ticket_id, file_name, file_url)
		VALUES ($1, $2, $3)
		RETURNING id, ticket_id, file_name, file_url, created_at
	`,
		a.TicketID,
		a.FileName,
		a.FileURL,
	).Scan(
		&a.ID,
		&a.TicketID,
		&a.FileName,
		&a.FileURL,
		&a.CreatedAt,
	)

	return a, mapErr(err)
}

func (s *Store) UpdateAttachment(ctx context.Context, a Attachment) (Attachment, error) {
	err := s.DB.QueryRow(ctx, `
		UPDATE attachments
		SET
			ticket_id = $1,
			file_name = $2,
			file_url = $3
		WHERE id = $4
		RETURNING id, ticket_id, file_name, file_url, created_at
	`,
		a.TicketID,
		a.FileName,
		a.FileURL,
		a.ID,
	).Scan(
		&a.ID,
		&a.TicketID,
		&a.FileName,
		&a.FileURL,
		&a.CreatedAt,
	)

	return a, mapErr(err)
}

func (s *Store) DeleteAttachment(ctx context.Context, id int64) error {
	result, err := s.DB.Exec(ctx, `
		DELETE FROM attachments
		WHERE id = $1
	`, id)

	if err != nil {
		return mapErr(err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}
