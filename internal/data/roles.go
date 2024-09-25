package data

import (
	"context"
	"database/sql"
	"time"
)

// TODO: реализовать role-based access control
const (
	AdminRole = "admin"
	UserRole  = "user"
	GuestRole = "guest"
)

type RoleModel struct {
	DB *sql.DB
}

type Roles []string

// TODO: getUserRole

func (r RoleModel) GetUserRoles(userId int64) (Roles, error) {

	q := `select role from roles
		inner join users_roles on roles.id = users_roles.role_id
		inner join users on users_roles.user_id = users.id
		where users.id = $1`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	rows, err := r.DB.QueryContext(ctx, q, userId)
	if err != nil {
		return nil, err
	}

	var roles Roles

	for rows.Next() {
		var role string
		err := rows.Scan(&role)
		if err != nil {
			return nil, err
		}
		roles = append(roles, role)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return roles, nil
}

// TODO: setUserRole

func (r RoleModel) AddPermissionToRole(roleId, permissionId int64) error {
	q := `insert into roles_permissions (role_id, permission_id)
        values ($1, $2)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.DB.ExecContext(ctx, q, roleId, permissionId)
	return err
}

// TODO: remove permissions from role

func (r RoleModel) AddRole(role string) error {
	q := `insert into roles (role) values ($1)`

	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()

	_, err := r.DB.ExecContext(ctx, q, role)
	return err
}

func (r RoleModel) hasPermission(role, permission string) (bool, error) {
	return true, nil
}
