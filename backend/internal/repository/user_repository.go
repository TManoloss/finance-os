package repository

import (
	"context"
	"time"

	"github.com/finance-os/backend/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

// UserRepository implementa a interface de persistência para usuários.
type UserRepository struct {
	db *pgxpool.Pool
}

// NewUserRepository cria uma nova instância de UserRepository.
func NewUserRepository(db *pgxpool.Pool) *UserRepository {
	return &UserRepository{db: db}
}

// Create insere um novo usuário no banco de dados.
func (r *UserRepository) Create(ctx context.Context, user *models.User) error {
	query := `INSERT INTO users (id, name, email, password_hash, created_at) 
	          VALUES (gen_random_uuid(), $1, $2, $3, NOW()) 
	          RETURNING id, created_at`
	
	err := r.db.QueryRow(ctx, query, user.Name, user.Email, user.PasswordHash).Scan(&user.ID, &user.CreatedAt)
	return err
}

// FindByEmail busca um usuário pelo email.
func (r *UserRepository) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at FROM users WHERE email = $1`
	
	var user models.User
	err := r.db.QueryRow(ctx, query, email).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// FindByID busca um usuário pelo ID.
func (r *UserRepository) FindByID(ctx context.Context, id string) (*models.User, error) {
	query := `SELECT id, name, email, password_hash, created_at FROM users WHERE id = $1`
	
	var user models.User
	err := r.db.QueryRow(ctx, query, id).Scan(
		&user.ID, &user.Name, &user.Email, &user.PasswordHash, &user.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// SaveRefreshToken salva um novo refresh token no banco de dados.
func (r *UserRepository) SaveRefreshToken(ctx context.Context, userID, token string, expiresAt time.Time) error {
	query := `INSERT INTO refresh_tokens (id, user_id, token, expires_at, created_at) 
	          VALUES (gen_random_uuid(), $1, $2, $3, NOW())`
	
	_, err := r.db.Exec(ctx, query, userID, token, expiresAt)
	return err
}

// ValidateRefreshToken verifica se um token é válido e retorna o user_id.
func (r *UserRepository) ValidateRefreshToken(ctx context.Context, token string) (string, error) {
	query := `SELECT user_id FROM refresh_tokens WHERE token = $1 AND expires_at > NOW()`
	
	var userID string
	err := r.db.QueryRow(ctx, query, token).Scan(&userID)
	if err != nil {
		return "", err
	}
	return userID, nil
}

// DeleteRefreshToken remove um refresh token do banco de dados (logout ou rotação).
func (r *UserRepository) DeleteRefreshToken(ctx context.Context, token string) error {
	query := `DELETE FROM refresh_tokens WHERE token = $1`
	
	_, err := r.db.Exec(ctx, query, token)
	return err
}
