package service

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Invoice representação de uma fatura de cartão de crédito.
type Invoice struct {
	AccountID         string        `json:"account_id"`
	DueDate           time.Time     `json:"due_date"`
	ClosedAmount      float64       `json:"closed_amount"`
	OpenAmount        float64       `json:"open_amount"`
	InstallmentsTotal float64       `json:"installments_total"`
	ProjectedTotal    float64       `json:"projected_total"`
	Breakdown         []InvoiceItem `json:"breakdown"`
}

type InvoiceItem struct {
	Description string    `json:"description"`
	Amount      float64   `json:"amount"`
	Type        string    `json:"type"` // transaction, installment
	Date        time.Time `json:"date"`
}

// ActiveInstallment representa um parcelamento em curso.
type ActiveInstallment struct {
	ID                 string    `json:"id"`
	MerchantName       string    `json:"merchant_name"`
	Amount             float64   `json:"amount"` // Valor da parcela individual
	TotalAmount        float64   `json:"total_amount"` // Valor total da compra
	InstallmentCurrent int       `json:"installment_current"`
	InstallmentsTotal  int       `json:"installments_total"`
	NextDueDate        time.Time `json:"next_due_date"`
	RemainingAmount    float64   `json:"remaining_amount"`
}

// InstallmentInfo contém dados extraídos de uma descrição de transação.
type InstallmentInfo struct {
	Current int
	Total   int
}

// InstallmentsService gerencia a detecção e persistência de parcelas.
type InstallmentsService struct {
	db *pgxpool.Pool
}

func NewInstallmentsService(db *pgxpool.Pool) *InstallmentsService {
	return &InstallmentsService{db: db}
}

// DetectInstallment tenta extrair X/Y de uma descrição.
func (s *InstallmentsService) DetectInstallment(description string) *InstallmentInfo {
	// Padrões comuns: (01/12), 1/10, [02/06]
	re := regexp.MustCompile(`\(?(\d{1,2})\s*/\s*(\d{1,2})\)?`)
	matches := re.FindStringSubmatch(description)

	if len(matches) == 3 {
		current, _ := strconv.Atoi(matches[1])
		total, _ := strconv.Atoi(matches[2])
		if current > 0 && total > 0 && current <= total {
			return &InstallmentInfo{Current: current, Total: total}
		}
	}
	return nil
}

// ProcessTransactions verifica e salva parcelamentos para um lote de transações.
func (s *InstallmentsService) ProcessTransactions(ctx context.Context, accountID string, transactions []map[string]interface{}) error {
	for _, tx := range transactions {
		var info *InstallmentInfo

		// 1. Tentar dados nativos da Pluggy
		if n, ok := tx["installment_number"].(int); ok {
			if t, ok := tx["installments_count"].(int); ok && t > 0 {
				info = &InstallmentInfo{Current: n, Total: t}
			}
		}

		// 2. Se não tiver, tentar Regex na descrição
		if info == nil {
			desc := tx["description"].(string)
			info = s.DetectInstallment(desc)
		}

		if info != nil {
			desc := tx["description"].(string)
			// Limpa o nome do estabelecimento removendo o (X/Y)
			merchant := strings.TrimSpace(regexp.MustCompile(`\(?\d{1,2}\s*/\s*\d{1,2}\)?`).ReplaceAllString(desc, ""))
			
			// No caso de dados nativos, o amount da transação é o valor da PARCELA.
			// No nosso banco salvamos o total_amount.
			totalAmount := tx["amount"].(float64) * float64(info.Total)

			_, err := s.db.Exec(ctx, `
				INSERT INTO installments (id, transaction_id, account_id, total_amount, installments_total, installment_current, merchant_name, start_date)
				VALUES (gen_random_uuid(), $1, $2, $3, $4, $5, $6, $7)
				ON CONFLICT DO NOTHING`,
				tx["id"], accountID, totalAmount, info.Total, info.Current, merchant, tx["date"])
			if err != nil {
				return err
			}
		}
	}
	return nil
}

// GetProjectedInvoice calcula a fatura projetada para uma conta de cartão.
func (s *InstallmentsService) GetProjectedInvoice(ctx context.Context, accountID string, referenceMonth time.Time) (*Invoice, error) {
	invoice := &Invoice{
		AccountID: accountID,
		Breakdown: []InvoiceItem{},
	}

	// 1. Buscar transações do ciclo atual (simplificado: transações do mês de referência)
	startOfMonth := time.Date(referenceMonth.Year(), referenceMonth.Month(), 1, 0, 0, 0, 0, referenceMonth.Location())
	endOfMonth := startOfMonth.AddDate(0, 1, -1)

	rows, err := s.db.Query(ctx, `
		SELECT description, amount, date FROM transactions 
		WHERE account_id = $1 AND date BETWEEN $2 AND $3 AND direction = 'debit'`,
		accountID, startOfMonth, endOfMonth)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item InvoiceItem
		item.Type = "transaction"
		if err := rows.Scan(&item.Description, &item.Amount, &item.Date); err != nil {
			continue
		}
		invoice.OpenAmount += item.Amount
		invoice.Breakdown = append(invoice.Breakdown, item)
	}

	// 2. Buscar parcelas que incidem neste mês
	instRows, err := s.db.Query(ctx, `
		SELECT merchant_name, total_amount, installments_total, installment_current, start_date
		FROM installments WHERE account_id = $1`, accountID)
	if err != nil {
		return nil, err
	}
	defer instRows.Close()

	for instRows.Next() {
		var inst struct {
			Merchant    string
			TotalAmount float64
			TotalParts  int
			CurrentPart int
			StartDate   time.Time
		}
		if err := instRows.Scan(&inst.Merchant, &inst.TotalAmount, &inst.TotalParts, &inst.CurrentPart, &inst.StartDate); err != nil {
			continue
		}

		partAmount := inst.TotalAmount / float64(inst.TotalParts)

		for i := 1; i <= inst.TotalParts; i++ {
			installmentMonth := inst.StartDate.AddDate(0, i-inst.CurrentPart, 0)
			if installmentMonth.Year() == referenceMonth.Year() && installmentMonth.Month() == referenceMonth.Month() {
				invoice.InstallmentsTotal += partAmount
				invoice.Breakdown = append(invoice.Breakdown, InvoiceItem{
					Description: fmt.Sprintf("%s (%d/%d)", inst.Merchant, i, inst.TotalParts),
					Amount:      partAmount,
					Type:        "installment",
					Date:        installmentMonth,
				})
			}
		}
	}

	invoice.ProjectedTotal = invoice.OpenAmount + invoice.InstallmentsTotal
	return invoice, nil
}

// GetActiveInstallments lista todos os parcelamentos ativos do usuário.
func (s *InstallmentsService) GetActiveInstallments(ctx context.Context, userID string) ([]ActiveInstallment, error) {
	query := `
		SELECT 
			i.id, i.merchant_name, i.total_amount, i.installments_total, i.installment_current, i.start_date
		FROM installments i
		JOIN connected_accounts acc ON i.account_id = acc.id
		WHERE acc.user_id = $1 AND i.installment_current < i.installments_total
		ORDER BY i.start_date ASC
	`
	rows, err := s.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var installments []ActiveInstallment
	for rows.Next() {
		var inst struct {
			ID           string
			Merchant     string
			TotalAmount  float64
			TotalParts   int
			CurrentPart  int
			StartDate    time.Time
		}
		if err := rows.Scan(&inst.ID, &inst.Merchant, &inst.TotalAmount, &inst.TotalParts, &inst.CurrentPart, &inst.StartDate); err != nil {
			continue
		}

		partAmount := inst.TotalAmount / float64(inst.TotalParts)
		remainingParts := inst.TotalParts - inst.CurrentPart
		// Próximo vencimento é no mês seguinte à transação original (simplificado)
		nextDue := inst.StartDate.AddDate(0, 1, 0)

		installments = append(installments, ActiveInstallment{
			ID:                 inst.ID,
			MerchantName:       inst.Merchant,
			Amount:             partAmount,
			TotalAmount:        inst.TotalAmount,
			InstallmentCurrent: inst.CurrentPart,
			InstallmentsTotal:  inst.TotalParts,
			NextDueDate:        nextDue,
			RemainingAmount:    partAmount * float64(remainingParts),
		})
	}

	return installments, nil
}
