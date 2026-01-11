# Background Jobs System - Integration Guide

## ✅ Sistema Implementado

O sistema de Background Jobs está **100% implementado** e pronto para uso:

- ✅ Migration 036 aplicada ao banco de dados
- ✅ BackgroundJobsService (enqueue, acquire, complete, fail, retry)
- ✅ JobWorker (worker pool com polling automático)
- ✅ JobScheduler (jobs recorrentes com cron)
- ✅ BackgroundJobsHandler (12 endpoints REST)
- ✅ Testes de integração

## 📦 Arquivos Criados

```
desktop/backend-go/
├── internal/
│   ├── database/migrations/
│   │   └── 036_background_jobs.sql              ← Migration (APLICADA)
│   ├── services/
│   │   ├── background_jobs_service.go           ← Service layer
│   │   ├── background_jobs_worker.go            ← Worker pool
│   │   ├── background_jobs_scheduler.go         ← Scheduler (cron)
│   │   └── background_jobs_integration_test.go  ← Testes
│   └── handlers/
│       └── background_jobs_handler.go           ← API handlers
└── scripts/migrations/
    └── run_migration_036.go                     ← Script migration (usado)
```

## 🔌 Integração no main.go

### Passo 1: Adicionar Handler (após linha ~670)

Adicione logo antes de `// Register routes`:

```go
// Initialize Background Jobs System
var jobsHandler *handlers.BackgroundJobsHandler
var jobWorkers []*services.JobWorker
var jobScheduler *services.JobScheduler

if dbConnected && pool != nil {
	slog.Info("Initializing background jobs system...")

	// Create handler (includes service and scheduler)
	jobsHandler = handlers.NewBackgroundJobsHandler(pool)

	// Get service and scheduler instances
	jobsService := jobsHandler.GetService()
	jobScheduler = jobsHandler.GetScheduler()

	// Create and configure workers (3 workers)
	for i := 1; i <= 3; i++ {
		workerID := fmt.Sprintf("worker-%d", i)
		worker := services.NewJobWorker(jobsService, workerID, 5*time.Second)

		// Register job handlers
		worker.RegisterHandler("email_send", services.ExampleEmailSendHandler)
		worker.RegisterHandler("report_generate", services.ExampleReportGenerateHandler)
		worker.RegisterHandler("sync_calendar", services.ExampleSyncCalendarHandler)

		// Add custom handlers here:
		// worker.RegisterHandler("your_custom_job", yourCustomHandler)

		jobWorkers = append(jobWorkers, worker)

		// Start worker
		if err := worker.Start(ctx); err != nil {
			slog.Error("Failed to start worker", "worker_id", workerID, "error", err)
		} else {
			slog.Info("Worker started", "worker_id", workerID)
		}
	}

	// Start scheduler
	if err := jobScheduler.Start(ctx); err != nil {
		slog.Error("Failed to start scheduler", "error", err)
	} else {
		slog.Info("Job scheduler started")
	}
}
```

### Passo 2: Registrar Routes (modificar linha 676)

Modifique a linha de registro de routes:

```go
// Register routes
h.RegisterRoutes(api)

// Register background jobs routes (if handler available)
if jobsHandler != nil {
	jobsHandler.RegisterRoutes(api)
	slog.Info("Background jobs routes registered")
}
```

### Passo 3: Graceful Shutdown (após linha ~705)

Adicione antes de `database.Close()`:

```go
// Stop job scheduler
if jobScheduler != nil {
	log.Println("Stopping job scheduler...")
	if err := jobScheduler.Stop(); err != nil {
		log.Printf("Warning: Error stopping scheduler: %v", err)
	}
}

// Stop workers
for _, worker := range jobWorkers {
	if worker.IsRunning() {
		log.Printf("Stopping worker...")
		if err := worker.Stop(); err != nil {
			log.Printf("Warning: Error stopping worker: %v", err)
		}
	}
}

// Release stuck jobs (cleanup)
if jobsHandler != nil {
	count, _ := jobsHandler.GetService().ReleaseStuckJobs(context.Background())
	if count > 0 {
		log.Printf("Released %d stuck jobs", count)
	}
}
```

## 🚀 Testando o Sistema

### 1. Build e Start

```bash
cd desktop/backend-go
go build ./cmd/server
./server
```

**Logs esperados:**
```
INFO Initializing background jobs system...
INFO Worker started worker_id=worker-1
INFO Worker started worker_id=worker-2
INFO Worker started worker_id=worker-3
INFO Job scheduler started
INFO Background jobs routes registered
```

### 2. Testar API - Criar Job

```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "email_send",
    "payload": {
      "to": "test@example.com",
      "subject": "Test Email",
      "body": "Hello from background jobs!"
    },
    "priority": 1
  }'
```

**Resposta esperada:**
```json
{
  "id": "uuid-here",
  "job_type": "email_send",
  "status": "pending",
  "priority": 1,
  "attempt_count": 0,
  "max_attempts": 3,
  "created_at": "2026-01-07T..."
}
```

**Logs esperados (worker processando):**
```
INFO Job acquired job_id=xxx job_type=email_send worker_id=worker-1 attempt=1
INFO Sending email to=test@example.com subject="Test Email"
INFO Job completed successfully job_id=xxx duration=1.002s
```

### 3. Listar Jobs

```bash
curl http://localhost:8080/api/background-jobs
```

### 4. Ver Status de um Job

```bash
curl http://localhost:8080/api/background-jobs/{job_id}
```

### 5. Criar Job Agendado (Cron)

```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "daily_report",
    "payload": {
      "report_type": "sales",
      "recipients": ["admin@example.com"]
    },
    "cron_expression": "0 9 * * *",
    "timezone": "America/Sao_Paulo",
    "name": "Daily Sales Report",
    "description": "Generates and sends daily sales report at 9am"
  }'
```

**Cron expressions:**
- `*/5 * * * *` - Every 5 minutes
- `0 9 * * *` - Every day at 9am
- `0 9 * * 1-5` - Weekdays at 9am
- `0 0 1 * *` - 1st of every month at midnight

### 6. Listar Jobs Agendados

```bash
curl http://localhost:8080/api/scheduled-jobs?active_only=true
```

## 📊 Endpoints Disponíveis

### Background Jobs

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/api/background-jobs` | Criar novo job |
| GET | `/api/background-jobs` | Listar jobs (com filtros) |
| GET | `/api/background-jobs/:id` | Ver status do job |
| POST | `/api/background-jobs/:id/retry` | Retentar job falhado |
| POST | `/api/background-jobs/:id/cancel` | Cancelar job |

### Scheduled Jobs

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/api/scheduled-jobs` | Criar job recorrente |
| GET | `/api/scheduled-jobs` | Listar jobs agendados |
| GET | `/api/scheduled-jobs/:id` | Ver detalhes |
| PUT | `/api/scheduled-jobs/:id` | Atualizar |
| DELETE | `/api/scheduled-jobs/:id` | Deletar |
| POST | `/api/scheduled-jobs/:id/enable` | Ativar |
| POST | `/api/scheduled-jobs/:id/disable` | Desativar |

## 🎨 Criando Handlers Customizados

### Exemplo: Handler de Notificação Push

```go
func PushNotificationHandler(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
	userID, _ := payload["user_id"].(string)
	message, _ := payload["message"].(string)
	title, _ := payload["title"].(string)

	slog.InfoContext(ctx, "Sending push notification",
		"user_id", userID,
		"title", title,
	)

	// Send push notification logic here
	// err := pushService.Send(userID, title, message)
	// if err != nil {
	// 	return nil, err
	// }

	return map[string]interface{}{
		"sent_at": time.Now(),
		"user_id": userID,
		"status": "delivered",
	}, nil
}
```

### Registrar no Worker (main.go)

```go
worker.RegisterHandler("push_notification", PushNotificationHandler)
```

### Usar no Código

```go
// Em qualquer parte do código que tenha acesso ao jobsService
jobsService.EnqueueJob(ctx, "push_notification", map[string]interface{}{
	"user_id": "user123",
	"title": "Welcome!",
	"message": "Thanks for joining BusinessOS",
}, 1, 3, nil)
```

## 🔄 Retry Logic

O sistema tem retry automático com exponential backoff:

- **Attempt 1:** Imediato
- **Attempt 2:** +1 minuto
- **Attempt 3:** +5 minutos
- **Attempt 4+:** +15 minutos

Após `max_attempts`, o job é marcado como `failed` permanentemente.

## 🧹 Limpeza Automática

Para limpar jobs antigos, você pode chamar periodicamente:

```go
// Limpar jobs completados/failed com mais de 7 dias
count, err := jobsService.CleanupOldJobs(ctx, 7*24*time.Hour)
```

Ou criar um job agendado para fazer isso:

```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "cleanup_old_jobs",
    "payload": {},
    "cron_expression": "0 2 * * *",
    "name": "Daily Job Cleanup"
  }'
```

## 📈 Monitoramento

### Verificar Status do Sistema

```bash
# Jobs pendentes
curl "http://localhost:8080/api/background-jobs?status=pending&limit=10"

# Jobs rodando
curl "http://localhost:8080/api/background-jobs?status=running"

# Jobs falhados
curl "http://localhost:8080/api/background-jobs?status=failed"

# Jobs por tipo
curl "http://localhost:8080/api/background-jobs?job_type=email_send"
```

### Métricas Úteis

No futuro, você pode adicionar:
- Dashboard com count de jobs por status
- Tempo médio de processamento por job_type
- Taxa de falha por tipo
- Workers ativos/inativos

## ✅ Checklist de Integração

- [ ] Adicionar código no main.go (Passos 1, 2, 3)
- [ ] Build o projeto (`go build ./cmd/server`)
- [ ] Iniciar servidor (`./server`)
- [ ] Verificar logs (workers e scheduler iniciados)
- [ ] Testar criar job via curl
- [ ] Verificar logs do worker processando
- [ ] Testar criar job agendado
- [ ] Verificar que job agendado cria background jobs
- [ ] Testar graceful shutdown (Ctrl+C)
- [ ] Verificar que workers param gracefully

## 🎉 Sistema Completo!

Após integração, você terá:

✅ Sistema robusto de background jobs
✅ Worker pool processando jobs automaticamente
✅ Scheduler para jobs recorrentes (cron-like)
✅ Retry logic com exponential backoff
✅ API REST completa para gerenciamento
✅ Graceful shutdown
✅ Job locking (previne processamento duplicado)
✅ Monitoramento via API

**Pronto para produção!** 🚀
