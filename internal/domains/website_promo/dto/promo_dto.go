package dto

import (
	"time"

	"github.com/google/uuid"
)

// Hero Promo DTOs

type CreateHeroPromoRequest struct {
	Title           string     `json:"title" validate:"required,max=255"`
	Subtitle        string     `json:"subtitle" validate:"max=255"`
	Description     string     `json:"description"`
	ImageURL        string     `json:"image_url" validate:"required,url"`
	ButtonText      string     `json:"button_text" validate:"max=100"`
	ButtonLink      string     `json:"button_link"`
	DisplayOrder    int        `json:"display_order"`
	DurationSeconds int        `json:"duration_seconds"`
	IsActive        bool       `json:"is_active"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
}

type UpdateHeroPromoRequest struct {
	Title           string     `json:"title" validate:"required,max=255"`
	Subtitle        string     `json:"subtitle" validate:"max=255"`
	Description     string     `json:"description"`
	ImageURL        string     `json:"image_url" validate:"required,url"`
	ButtonText      string     `json:"button_text" validate:"max=100"`
	ButtonLink      string     `json:"button_link"`
	DisplayOrder    int        `json:"display_order"`
	DurationSeconds int        `json:"duration_seconds"`
	IsActive        bool       `json:"is_active"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
}

type HeroPromoResponse struct {
	ID              uuid.UUID  `json:"id"`
	Title           string     `json:"title"`
	Subtitle        *string    `json:"subtitle"`
	Description     *string    `json:"description"`
	ImageURL        string     `json:"image_url"`
	ButtonText      *string    `json:"button_text"`
	ButtonLink      *string    `json:"button_link"`
	DisplayOrder    int        `json:"display_order"`
	DurationSeconds int        `json:"duration_seconds"`
	IsActive        bool       `json:"is_active"`
	StartDate       *time.Time `json:"start_date"`
	EndDate         *time.Time `json:"end_date"`
	CreatedAt       time.Time  `json:"created_at"`
	UpdatedAt       time.Time  `json:"updated_at"`
}

// Feature Card DTOs

type CreateFeatureCardRequest struct {
	Title        string     `json:"title" validate:"required,max=255"`
	Description  string     `json:"description"`
	ImageURL     string     `json:"image_url" validate:"required,url"`
	ButtonText   string     `json:"button_text" validate:"max=100"`
	ButtonLink   string     `json:"button_link"`
	DisplayOrder int        `json:"display_order"`
	IsActive     bool       `json:"is_active"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
}

type UpdateFeatureCardRequest struct {
	Title        string     `json:"title" validate:"required,max=255"`
	Description  string     `json:"description"`
	ImageURL     string     `json:"image_url" validate:"required,url"`
	ButtonText   string     `json:"button_text" validate:"max=100"`
	ButtonLink   string     `json:"button_link"`
	DisplayOrder int        `json:"display_order"`
	IsActive     bool       `json:"is_active"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
}

type FeatureCardResponse struct {
	ID           uuid.UUID  `json:"id"`
	Title        string     `json:"title"`
	Description  *string    `json:"description"`
	ImageURL     string     `json:"image_url"`
	ButtonText   *string    `json:"button_text"`
	ButtonLink   *string    `json:"button_link"`
	DisplayOrder int        `json:"display_order"`
	IsActive     bool       `json:"is_active"`
	StartDate    *time.Time `json:"start_date"`
	EndDate      *time.Time `json:"end_date"`
	CreatedAt    time.Time  `json:"created_at"`
	UpdatedAt    time.Time  `json:"updated_at"`
}
