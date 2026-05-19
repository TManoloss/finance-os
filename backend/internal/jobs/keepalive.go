package jobs

import (
	"context"
	"log"
	"net/http"
	"time"
)

// KeepAliveJob mantém o serviço acordado no Render free tier.
type KeepAliveJob struct {
	selfURL    string
	httpClient *http.Client
}

// NewKeepAliveJob cria um novo job de keep-alive.
func NewKeepAliveJob(selfURL string) *KeepAliveJob {
	return &KeepAliveJob{
		selfURL: selfURL,
		httpClient: &http.Client{
			Timeout: 15 * time.Second,
		},
	}
}

// Start inicia o loop de keep-alive.
func (j *KeepAliveJob) Start(ctx context.Context) {
	if j.selfURL == "" {
		log.Println("[KeepAlive] SELF_URL não definida, ignorando keep-alive")
		return
	}

	ticker := time.NewTicker(10 * time.Minute)
	defer ticker.Stop()

	log.Printf("[KeepAlive] Iniciando pings para %s/health a cada 10 minutos", j.selfURL)

	consecutiveFailures := 0

	for {
		select {
		case <-ctx.Done():
			log.Println("[KeepAlive] Parando job de keep-alive...")
			return
		case <-ticker.C:
			resp, err := j.httpClient.Get(j.selfURL + "/health")
			if err != nil {
				consecutiveFailures++
				log.Printf("[KeepAlive] Erro ao realizar ping (%d falhas consecutivas): %v", consecutiveFailures, err)
			} else {
				resp.Body.Close()
				if resp.StatusCode != http.StatusOK {
					consecutiveFailures++
					log.Printf("[KeepAlive] Resposta inesperada (%d falhas consecutivas): Status %d", consecutiveFailures, resp.StatusCode)
				} else {
					consecutiveFailures = 0
				}
			}

			if consecutiveFailures >= 3 {
				log.Printf("[KeepAlive] WARNING: O serviço pode estar com problemas de conectividade ou indisponível.")
			}
		}
	}
}
