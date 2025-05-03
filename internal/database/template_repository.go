package database

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/google/uuid"
)

// TemplateRepository handles database operations for document templates
type TemplateRepository struct {
	db *sql.DB
}

// Template represents a document template entity
type Template struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ContentType string    `json:"contentType"`
	Content     []byte    `json:"content,omitempty"`
	CreatedAt   time.Time `json:"createdAt"`
}

// NewTemplateRepository creates a new template repository
func NewTemplateRepository(database *Database) *TemplateRepository {
	return &TemplateRepository{
		db: database.GetDB(),
	}
}

// CreateTemplate creates a new template
func (r *TemplateRepository) CreateTemplate(name, description, contentType string, content []byte) (*Template, error) {
	id := uuid.New().String()
	now := time.Now()

	_, err := r.db.Exec(
		"INSERT INTO templates (id, name, description, content_type, content, created_at) VALUES (?, ?, ?, ?, ?, ?)",
		id, name, description, contentType, content, now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create template: %w", err)
	}

	return &Template{
		ID:          id,
		Name:        name,
		Description: description,
		ContentType: contentType,
		Content:     content,
		CreatedAt:   now,
	}, nil
}

// GetTemplate gets a template by ID
func (r *TemplateRepository) GetTemplate(id string) (*Template, error) {
	var template Template
	var createdAt string

	err := r.db.QueryRow(
		"SELECT id, name, description, content_type, content, created_at FROM templates WHERE id = ?",
		id,
	).Scan(&template.ID, &template.Name, &template.Description, &template.ContentType, &template.Content, &createdAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("template not found: %s", id)
		}
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Parse created_at
	template.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
	if err != nil {
		return nil, fmt.Errorf("failed to parse created_at: %w", err)
	}

	return &template, nil
}

// GetTemplates gets all templates
func (r *TemplateRepository) GetTemplates() ([]*Template, error) {
	rows, err := r.db.Query(
		"SELECT id, name, description, content_type, created_at FROM templates ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to get templates: %w", err)
	}
	defer rows.Close()

	templates := []*Template{}
	for rows.Next() {
		var template Template
		var createdAt string

		err := rows.Scan(&template.ID, &template.Name, &template.Description, &template.ContentType, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan template: %w", err)
		}

		// Parse created_at
		template.CreatedAt, err = time.Parse(time.RFC3339, createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to parse created_at: %w", err)
		}

		templates = append(templates, &template)
	}

	return templates, nil
}

// UpdateTemplate updates a template
func (r *TemplateRepository) UpdateTemplate(id, name, description, contentType string, content []byte) (*Template, error) {
	_, err := r.db.Exec(
		"UPDATE templates SET name = ?, description = ?, content_type = ?, content = ? WHERE id = ?",
		name, description, contentType, content, id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update template: %w", err)
	}

	return r.GetTemplate(id)
}

// DeleteTemplate deletes a template by ID
func (r *TemplateRepository) DeleteTemplate(id string) error {
	_, err := r.db.Exec("DELETE FROM templates WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete template: %w", err)
	}
	return nil
}

// CreateDocumentFromTemplate creates a new document from a template
func (r *TemplateRepository) CreateDocumentFromTemplate(templateID, title string) (*Document, error) {
	// Get the template
	template, err := r.GetTemplate(templateID)
	if err != nil {
		return nil, fmt.Errorf("failed to get template: %w", err)
	}

	// Create a document repository
	docRepo := NewDocumentRepository(&Database{db: r.db})

	// Create the document
	doc, err := docRepo.CreateDocument(title, template.ContentType, template.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to create document from template: %w", err)
	}

	return doc, nil
}

// GetTemplateCount gets the total number of templates
func (r *TemplateRepository) GetTemplateCount() (int, error) {
	var count int
	err := r.db.QueryRow("SELECT COUNT(*) FROM templates").Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get template count: %w", err)
	}
	return count, nil
}
