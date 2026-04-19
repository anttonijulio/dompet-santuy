package domain

import "time"

type Category struct {
	ID        string
	UserID    string
	Name      string
	Icon      string
	Color     string
	Type      string
	CreatedAt time.Time
}

type Transaction struct {
	ID         string
	UserID     string
	CategoryID string
	Amount     int64
	Type       string
	Note       string
	Date       time.Time
	CreatedAt  time.Time
	Category   *Category
}

type CreateCategoryRequest struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

type UpdateCategoryRequest struct {
	Name  string `json:"name"`
	Icon  string `json:"icon"`
	Color string `json:"color"`
	Type  string `json:"type"`
}

type CreateTransactionRequest struct {
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"`
	Type       string `json:"type"`
	Note       string `json:"note"`
	Date       string `json:"date"` // RFC3339 or "2006-01-02"
}

type UpdateTransactionRequest struct {
	CategoryID string `json:"category_id"`
	Amount     int64  `json:"amount"`
	Type       string `json:"type"`
	Note       string `json:"note"`
	Date       string `json:"date"`
}

type ListTransactionsFilter struct {
	StartDate string
	EndDate   string
	Type      string
	Limit     int
	Offset    int
}

type CategoryResponse struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Icon  string `json:"icon,omitempty"`
	Color string `json:"color,omitempty"`
	Type  string `json:"type"`
}

type TransactionResponse struct {
	ID        string           `json:"id"`
	Category  CategoryResponse `json:"category"`
	Amount    int64            `json:"amount"`
	Type      string           `json:"type"`
	Note      string           `json:"note,omitempty"`
	Date      time.Time        `json:"date"`
	CreatedAt time.Time        `json:"created_at"`
}

func (c *Category) ToResponse() CategoryResponse {
	return CategoryResponse{
		ID:    c.ID,
		Name:  c.Name,
		Icon:  c.Icon,
		Color: c.Color,
		Type:  c.Type,
	}
}

func (t *Transaction) ToResponse() TransactionResponse {
	r := TransactionResponse{
		ID:        t.ID,
		Amount:    t.Amount,
		Type:      t.Type,
		Note:      t.Note,
		Date:      t.Date,
		CreatedAt: t.CreatedAt,
	}
	if t.Category != nil {
		r.Category = t.Category.ToResponse()
	}
	return r
}
