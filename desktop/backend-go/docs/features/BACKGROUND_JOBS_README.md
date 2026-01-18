# Background Jobs System - Implementation Complete ✅

## 🎉 Status: READY FOR PRODUCTION

O sistema de Background Jobs está **100% implementado e integrado** no BusinessOS!

## 📦 O Que Foi Implementado

### 1. Database (Migration 036)
- ✅ Tabela `background_jobs` - jobs individuais
- ✅ Tabela `scheduled_jobs` - jobs recorrentes (cron)
- ✅ Funções SQL: `acquire_background_job`, `calculate_retry_time`, `release_stuck_jobs`
- ✅ Índices otimizados para queries rápidas
- ✅ **APLICADA NO BANCO DE DADOS SUPABASE**

### 2. Service Layer (3 arquivos)
- ✅ `background_jobs_service.go` - CRUD de jobs
- ✅ `background_jobs_worker.go` - Worker pool
- ✅ `background_jobs_scheduler.go` - Scheduler com cron

### 3. API Handlers
- ✅ 12 endpoints REST completos
- ✅ Validação de input
- ✅ Error handling adequado
- ✅ **INTEGRADO NO MAIN.GO**

### 4. Integração Main.go
- ✅ 3 workers iniciados automaticamente
- ✅ Scheduler iniciado
- ✅ Routes registradas
- ✅ Graceful shutdown implementado

## 🚀 Como Usar

### Iniciar o Servidor

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
Server starting on port 8080
```

### Testar o Sistema

#### Opção 1: Script Automático

```bash
cd desktop/backend-go/scripts/tests
chmod +x test_background_jobs_api.sh
./test_background_jobs_api.sh
```

Este script testa todos os 12 endpoints automaticamente!

#### Opção 2: Testes Manuais

**Criar um job:**
```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "email_send",
    "payload": {
      "to": "test@example.com",
      "subject": "Test",
      "body": "Hello!"
    },
    "priority": 1
  }'
```

**Listar jobs:**
```bash
curl http://localhost:8080/api/background-jobs
```

**Ver status de um job:**
```bash
curl http://localhost:8080/api/background-jobs/{job_id}
```

**Criar job agendado (todo dia às 9h):**
```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -H 'Content-Type: application/json' \
  -d '{
    "job_type": "daily_report",
    "payload": {"report_type": "sales"},
    "cron_expression": "0 9 * * *",
    "timezone": "America/Sao_Paulo",
    "name": "Daily Sales Report"
  }'
```

## 📊 Endpoints Disponíveis

### Background Jobs

| Método | Endpoint | Descrição |
|--------|----------|-----------|
| POST | `/api/background-jobs` | Criar job |
| GET | `/api/background-jobs` | Listar jobs (com filtros) |
| GET | `/api/background-jobs/:id` | Ver status |
| POST | `/api/background-jobs/:id/retry` | Retentar job |
| POST | `/api/background-jobs/:id/cancel` | Cancelar job |

**Filtros para listagem:**
- `?status=pending` - Apenas pendentes
- `?status=running` - Apenas rodando
- `?status=completed` - Apenas completados
- `?status=failed` - Apenas falhados
- `?job_type=email_send` - Por tipo
- `?limit=50&offset=0` - Paginação

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

## 🔧 Arquitetura

```
User Request (API)
    ↓
BackgroundJobsHandler
    ↓
BackgroundJobsService.EnqueueJob()
    ↓
Database (background_jobs table)
    ↓
JobWorker.AcquireJob() [polling a cada 5s]
    ↓
ProcessJob()
    ↓
JobHandler (email_send, report_generate, etc)
    ↓
CompleteJob() ou FailJob()
    ↓
Database (status updated)
```

### Retry Logic (Exponential Backoff)

```
Attempt 1: Imediato
Attempt 2: +1 minuto
Attempt 3: +5 minutos
Attempt 4+: +15 minutos

Max attempts: 3 (default)
```

Se falhar após max_attempts → status = `failed` (permanente)

## 🎨 Handlers Disponíveis

Atualmente registrados no sistema:

1. **`email_send`** - Envio de emails
   - Payload: `{to, subject, body}`

2. **`report_generate`** - Geração de relatórios
   - Payload: `{report_type, start_date, end_date}`

3. **`sync_calendar`** - Sincronização de calendário
   - Payload: `{user_id, calendar_id}`

### Adicionar Handler Customizado

**1. Criar o handler:**

```go
// Em qualquer arquivo .go do projeto
func CustomNotificationHandler(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
    userID, _ := payload["user_id"].(string)
    message, _ := payload["message"].(string)

    slog.InfoContext(ctx, "Sending notification", "user_id", userID)

    // Sua lógica aqui
    // ...

    return map[string]interface{}{
        "sent_at": time.Now(),
        "status": "delivered",
    }, nil
}
```

**2. Registrar no main.go (linha ~700):**

```go
worker.RegisterHandler("send_notification", CustomNotificationHandler)
```

**3. Usar:**

```bash
curl -X POST http://localhost:8080/api/background-jobs \
  -d '{"job_type": "send_notification", "payload": {"user_id": "123", "message": "Hello!"}}'
```

## 📝 Expressões Cron

Formato: `minuto hora dia mês dia_da_semana`

Exemplos:
- `*/5 * * * *` - A cada 5 minutos
- `0 9 * * *` - Todo dia às 9h
- `0 9 * * 1-5` - Dias úteis às 9h
- `0 0 1 * *` - Primeiro dia do mês à meia-noite
- `*/15 9-17 * * 1-5` - A cada 15min, 9h-17h, dias úteis

## 🧹 Limpeza de Jobs Antigos

Para evitar acúmulo no banco, você pode:

**1. Criar job agendado para cleanup:**

```bash
curl -X POST http://localhost:8080/api/scheduled-jobs \
  -d '{
    "job_type": "cleanup_old_jobs",
    "payload": {"older_than_days": 7},
    "cron_expression": "0 2 * * *",
    "name": "Daily Job Cleanup"
  }'
```

**2. Implementar handler de cleanup:**

```go
func CleanupOldJobsHandler(ctx context.Context, payload map[string]interface{}) (interface{}, error) {
    days, _ := payload["older_than_days"].(float64)
    duration := time.Duration(days) * 24 * time.Hour

    // Assumindo que jobsService está disponível
    count, err := jobsService.CleanupOldJobs(ctx, duration)
    if err != nil {
        return nil, err
    }

    return map[string]interface{}{
        "deleted_count": count,
        "older_than": duration.String(),
    }, nil
}
```

## 🐛 Debugging

### Ver Jobs Pendentes

```bash
curl "http://localhost:8080/api/background-jobs?status=pending"
```

### Ver Jobs Falhados

```bash
curl "http://localhost:8080/api/background-jobs?status=failed"
```

### Ver Jobs Rodando

```bash
curl "http://localhost:8080/api/background-jobs?status=running"
```

### Logs do Servidor

Procure por:
- `Worker started` - Workers iniciados
- `Job acquired` - Worker pegou um job
- `Processing job` - Job sendo processado
- `Job completed successfully` - Job completado
- `Job execution failed` - Job falhou
- `Job scheduled for retry` - Job será retentado

## 📈 Próximas Melhorias (Futuras)

Ideias para expandir o sistema:

1. **Dashboard Web**
   - Visualizar jobs em tempo real
   - Métricas (taxa de sucesso, tempo médio, etc)
   - Pausar/resumir workers

2. **Notificações**
   - Webhook quando job completa/falha
   - Email/Slack notifications

3. **Prioridade Dinâmica**
   - Ajustar prioridade baseado em carga
   - Queue prioritária para jobs urgentes

4. **Job Dependencies**
   - Job B só roda após Job A completar
   - DAG de jobs

5. **Métricas e Monitoring**
   - Prometheus metrics
   - Grafana dashboards

## ✅ Checklist de Verificação

- [x] Migration 036 aplicada
- [x] Service layer implementado
- [x] Worker pool implementado
- [x] Scheduler implementado
- [x] Handlers implementados
- [x] Routes registradas
- [x] Integrado no main.go
- [x] Graceful shutdown
- [x] Script de testes criado
- [x] Documentação completa

## 🎉 Conclusão

O sistema de Background Jobs está **pronto para produção**!

✅ Workers processam jobs automaticamente
✅ Retry logic com exponential backoff
✅ Jobs agendados com cron
✅ API REST completa
✅ Graceful shutdown
✅ Totalmente documentado

**Aproveite!** 🚀

---

**Documentação completa:** `BACKGROUND_JOBS_INTEGRATION_GUIDE.md`
**Script de testes:** `scripts/tests/test_background_jobs_api.sh`
**Arquivos fonte:** `internal/services/background_jobs_*.go`, `internal/handlers/background_jobs_handler.go`
