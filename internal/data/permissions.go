package data

import (
	"context"
	"database/sql"
	"github.com/lib/pq"
	"time"
)

type PermissionsModel struct {
	DB *sql.DB
}

type Permissions []string

func (p Permissions) Include(permission string) bool {
	for _, perm := range p {
		if perm == permission {
			return true
		}
	}
	return false
}

// GetPermissionsForRole

func (p PermissionsModel) GetForUser(userId int64) (Permissions, error) {
	q := `select p.permission
            from permissions as p
            inner join users_permissions as up on p.id = up.permission_id
            inner join users as u on up.user_id = u.id
            where u.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := p.DB.QueryContext(ctx, q, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var perm Permissions

	for rows.Next() {
		var permission string
		err := rows.Scan(&permission)
		if err != nil {
			return nil, err
		}
		perm = append(perm, permission)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}

	return perm, nil
}

func (p PermissionsModel) Grant(userId int64, permissions ...string) error {
	q := `insert into users_permissions (user_id, permission_id)
		select $1, p.id from permissions as p where p.permission = any($2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := p.DB.ExecContext(ctx, q, userId, pq.Array(permissions))
	return err
}
