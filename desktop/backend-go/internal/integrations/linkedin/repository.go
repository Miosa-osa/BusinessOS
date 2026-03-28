package linkedin

import (
	"database/sql"
	"fmt"
	"log/slog"
	"time"
)

// Repository provides data access for LinkedIn integration.
// Uses database/sql with parameterized queries for safety.
type Repository struct {
	logger *slog.Logger
	db     *sql.DB
}

// NewRepository creates a new LinkedIn repository.
func NewRepository(logger *slog.Logger, db *sql.DB) *Repository {
	return &Repository{
		logger: logger,
		db:     db,
	}
}

// CreateContact inserts a new contact into the database.
func (r *Repository) CreateContact(contact *Contact) error {
	query := `
		INSERT INTO linkedin_contacts
		(linkedin_id, name, title, company, industry, connection_date, icp_score, icp_scored_at, raw_csv, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		contact.LinkedInID,
		contact.Name,
		contact.Title,
		contact.Company,
		contact.Industry,
		contact.ConnectionDate,
		contact.ICPScore,
		contact.ICPScoredAt,
		contact.RawCSV,
		contact.CreatedAt,
		contact.UpdatedAt,
	).Scan(&contact.ID)

	if err != nil {
		r.logger.Error("CreateContact failed", "error", err)
		return fmt.Errorf("create contact failed: %w", err)
	}

	r.logger.Debug("Contact created", "id", contact.ID, "name", contact.Name)
	return nil
}

// UpdateContact updates an existing contact.
func (r *Repository) UpdateContact(contact *Contact) error {
	query := `
		UPDATE linkedin_contacts
		SET name = $1, title = $2, company = $3, industry = $4, icp_score = $5,
		    icp_scored_at = $6, updated_at = $7
		WHERE id = $8
	`

	_, err := r.db.Exec(
		query,
		contact.Name,
		contact.Title,
		contact.Company,
		contact.Industry,
		contact.ICPScore,
		contact.ICPScoredAt,
		contact.UpdatedAt,
		contact.ID,
	)

	if err != nil {
		r.logger.Error("UpdateContact failed", "id", contact.ID, "error", err)
		return fmt.Errorf("update contact failed: %w", err)
	}

	r.logger.Debug("Contact updated", "id", contact.ID, "name", contact.Name)
	return nil
}

// GetContactByEmail retrieves a contact by email (derived from linkedin_id prefix).
// Returns nil if not found.
func (r *Repository) GetContactByEmail(email string) (*Contact, error) {
	query := `
		SELECT id, linkedin_id, name, title, company, industry, connection_date,
		       icp_score, icp_scored_at, raw_csv, created_at, updated_at
		FROM linkedin_contacts
		WHERE raw_csv LIKE $1
		LIMIT 1
	`

	contact := &Contact{}
	err := r.db.QueryRow(query, "%"+email+"%").Scan(
		&contact.ID,
		&contact.LinkedInID,
		&contact.Name,
		&contact.Title,
		&contact.Company,
		&contact.Industry,
		&contact.ConnectionDate,
		&contact.ICPScore,
		&contact.ICPScoredAt,
		&contact.RawCSV,
		&contact.CreatedAt,
		&contact.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("GetContactByEmail failed", "email", email, "error", err)
		return nil, fmt.Errorf("get contact by email failed: %w", err)
	}

	return contact, nil
}

// GetQualifiedContacts retrieves contacts with ICP score >= minScore.
func (r *Repository) GetQualifiedContacts(minScore float64, limit int) ([]*Contact, error) {
	query := `
		SELECT id, linkedin_id, name, title, company, industry, connection_date,
		       icp_score, icp_scored_at, raw_csv, created_at, updated_at
		FROM linkedin_contacts
		WHERE icp_score >= $1
		ORDER BY icp_score DESC, created_at DESC
		LIMIT $2
	`

	rows, err := r.db.Query(query, minScore, limit)
	if err != nil {
		r.logger.Error("GetQualifiedContacts failed", "error", err)
		return nil, fmt.Errorf("get qualified contacts failed: %w", err)
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		contact := &Contact{}
		if err := rows.Scan(
			&contact.ID,
			&contact.LinkedInID,
			&contact.Name,
			&contact.Title,
			&contact.Company,
			&contact.Industry,
			&contact.ConnectionDate,
			&contact.ICPScore,
			&contact.ICPScoredAt,
			&contact.RawCSV,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		); err != nil {
			r.logger.Error("Row scan failed", "error", err)
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		contacts = append(contacts, contact)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("GetQualifiedContacts iteration failed", "error", err)
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return contacts, nil
}

// GetQualifiedContactsPaginated retrieves contacts with ICP score >= minScore,
// with offset-based pagination support.
func (r *Repository) GetQualifiedContactsPaginated(minScore float64, limit, offset int) ([]*Contact, error) {
	query := `
		SELECT id, linkedin_id, name, title, company, industry, connection_date,
		       icp_score, icp_scored_at, raw_csv, created_at, updated_at
		FROM linkedin_contacts
		WHERE icp_score >= $1
		ORDER BY icp_score DESC, created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.Query(query, minScore, limit, offset)
	if err != nil {
		r.logger.Error("GetQualifiedContactsPaginated failed", "error", err)
		return nil, fmt.Errorf("get qualified contacts paginated failed: %w", err)
	}
	defer rows.Close()

	var contacts []*Contact
	for rows.Next() {
		contact := &Contact{}
		if err := rows.Scan(
			&contact.ID,
			&contact.LinkedInID,
			&contact.Name,
			&contact.Title,
			&contact.Company,
			&contact.Industry,
			&contact.ConnectionDate,
			&contact.ICPScore,
			&contact.ICPScoredAt,
			&contact.RawCSV,
			&contact.CreatedAt,
			&contact.UpdatedAt,
		); err != nil {
			r.logger.Error("Row scan failed", "error", err)
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		contacts = append(contacts, contact)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("GetQualifiedContactsPaginated iteration failed", "error", err)
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return contacts, nil
}

// CountQualifiedContacts returns the total count of contacts with ICP score >= minScore.
func (r *Repository) CountQualifiedContacts(minScore float64) (int64, error) {
	var count int64
	err := r.db.QueryRow(
		"SELECT COUNT(*) FROM linkedin_contacts WHERE icp_score >= $1",
		minScore,
	).Scan(&count)
	if err != nil {
		r.logger.Error("CountQualifiedContacts failed", "error", err)
		return 0, fmt.Errorf("count qualified contacts failed: %w", err)
	}
	return count, nil
}

// CreateMessage inserts a message into the queue.
func (r *Repository) CreateMessage(msg *LinkedInMessageQueue) error {
	query := `
		INSERT INTO linkedin_message_queue
		(contact_id, step_id, scheduled_at, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`

	err := r.db.QueryRow(
		query,
		msg.ContactID,
		msg.StepID,
		msg.ScheduledAt,
		msg.Status,
		msg.CreatedAt,
		msg.UpdatedAt,
	).Scan(&msg.ID)

	if err != nil {
		r.logger.Error("CreateMessage failed", "error", err)
		return fmt.Errorf("create message failed: %w", err)
	}

	r.logger.Debug("Message created", "id", msg.ID, "contact_id", msg.ContactID)
	return nil
}

// GetPendingMessages retrieves all messages with status = 'pending'.
func (r *Repository) GetPendingMessages() ([]*LinkedInMessageQueue, error) {
	query := `
		SELECT id, contact_id, step_id, scheduled_at, sent_at, status, created_at, updated_at
		FROM linkedin_message_queue
		WHERE status = 'pending'
		ORDER BY scheduled_at ASC
	`

	rows, err := r.db.Query(query)
	if err != nil {
		r.logger.Error("GetPendingMessages failed", "error", err)
		return nil, fmt.Errorf("get pending messages failed: %w", err)
	}
	defer rows.Close()

	var messages []*LinkedInMessageQueue
	for rows.Next() {
		msg := &LinkedInMessageQueue{}
		if err := rows.Scan(
			&msg.ID,
			&msg.ContactID,
			&msg.StepID,
			&msg.ScheduledAt,
			&msg.SentAt,
			&msg.Status,
			&msg.CreatedAt,
			&msg.UpdatedAt,
		); err != nil {
			r.logger.Error("Row scan failed", "error", err)
			return nil, fmt.Errorf("row scan failed: %w", err)
		}
		messages = append(messages, msg)
	}

	if err = rows.Err(); err != nil {
		r.logger.Error("GetPendingMessages iteration failed", "error", err)
		return nil, fmt.Errorf("iteration failed: %w", err)
	}

	return messages, nil
}

// MarkMessageSent marks a message as sent.
func (r *Repository) MarkMessageSent(messageID int64) error {
	query := `
		UPDATE linkedin_message_queue
		SET status = 'sent', sent_at = $1, updated_at = $2
		WHERE id = $3
	`

	now := time.Now()
	_, err := r.db.Exec(query, now, now, messageID)
	if err != nil {
		r.logger.Error("MarkMessageSent failed", "id", messageID, "error", err)
		return fmt.Errorf("mark message sent failed: %w", err)
	}

	r.logger.Debug("Message marked sent", "id", messageID)
	return nil
}

// MarkMessageFailed marks a message as failed (not sent).
func (r *Repository) MarkMessageFailed(messageID int64, reason string) error {
	query := `
		UPDATE linkedin_message_queue
		SET status = 'failed', updated_at = $1
		WHERE id = $2
	`

	_, err := r.db.Exec(query, time.Now(), messageID)
	if err != nil {
		r.logger.Error("MarkMessageFailed failed", "id", messageID, "error", err)
		return fmt.Errorf("mark message failed failed: %w", err)
	}

	r.logger.Debug("Message marked failed", "id", messageID, "reason", reason)
	return nil
}

// CreateSequence inserts a new outreach sequence.
func (r *Repository) CreateSequence(seq *OutreachSequence) error {
	query := `
		INSERT INTO outreach_sequences (name, target_icp_min_score, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`

	now := time.Now()
	err := r.db.QueryRow(query, seq.Name, seq.TargetICPMinScore, now, now).Scan(&seq.ID)
	if err != nil {
		r.logger.Error("CreateSequence failed", "error", err)
		return fmt.Errorf("create sequence failed: %w", err)
	}

	r.logger.Debug("Sequence created", "id", seq.ID, "name", seq.Name)
	return nil
}

// GetSequence retrieves a sequence by ID.
func (r *Repository) GetSequence(sequenceID int64) (*OutreachSequence, error) {
	query := `
		SELECT id, name, target_icp_min_score, created_at, updated_at
		FROM outreach_sequences
		WHERE id = $1
	`

	seq := &OutreachSequence{}
	err := r.db.QueryRow(query, sequenceID).Scan(
		&seq.ID,
		&seq.Name,
		&seq.TargetICPMinScore,
		&seq.CreatedAt,
		&seq.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		r.logger.Error("GetSequence failed", "id", sequenceID, "error", err)
		return nil, fmt.Errorf("get sequence failed: %w", err)
	}

	return seq, nil
}
