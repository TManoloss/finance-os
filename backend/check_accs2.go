package main
import (
	"context"
	"fmt"
	"log"
	"github.com/finance-os/backend/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)
func main() {
	cfg, _ := config.Load()
	dbPool, _ := pgxpool.New(context.Background(), cfg.DatabaseURL)
	rows, _ := dbPool.Query(context.Background(), "SELECT id, user_id, COALESCE(pluggy_item_id, 'NULL'), COALESCE(pluggy_account_id, 'NULL') FROM connected_accounts")
	for rows.Next() {
		var id, uid, iid, aid string
		rows.Scan(&id, &uid, &iid, &aid)
		fmt.Printf("UID: %s | Item: %s | Acc: %s\n", uid, iid, aid)
	}
}
