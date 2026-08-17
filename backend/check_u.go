package main
import (
	"context"
	"fmt"
	"github.com/finance-os/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)
func main() {
	cfg, _ := config.Load()
	dbPool, _ := pgxpool.New(context.Background(), cfg.DatabaseURL)
	rows, _ := dbPool.Query(context.Background(), "SELECT id, email FROM users")
	for rows.Next() {
		var id, email string
		rows.Scan(&id, &email)
		fmt.Printf("%s = %s\n", id, email)
	}
}
