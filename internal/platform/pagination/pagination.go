package pagination

import (
	"net/url"
	"strconv"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

const (
	DefaultPage  = 1
	DefaultLimit = 20
	MaxLimit     = 100
)

type Params struct {
	Page  int `json:"page"`
	Limit int `json:"limit"`
}

type Metadata struct {
	Page       int   `json:"page"`
	Limit      int   `json:"limit"`
	Total      int64 `json:"total"`
	TotalPages int   `json:"totalPages"`
}

func Parse(values url.Values) (Params, error) {
	params := Params{
		Page:  DefaultPage,
		Limit: DefaultLimit,
	}

	var details []apierror.Detail

	if rawPage := values.Get("page"); rawPage != "" {
		page, err := strconv.Atoi(rawPage)
		if err != nil || page <= 0 {
			details = append(details, apierror.Detail{
				Field:       "page",
				Constraints: []string{"page must be a positive integer"},
			})
		} else {
			params.Page = page
		}
	}

	if rawLimit := values.Get("limit"); rawLimit != "" {
		limit, err := strconv.Atoi(rawLimit)
		if err != nil || limit <= 0 {
			details = append(details, apierror.Detail{
				Field:       "limit",
				Constraints: []string{"limit must be a positive integer"},
			})
		} else if limit > MaxLimit {
			details = append(details, apierror.Detail{
				Field:       "limit",
				Constraints: []string{"limit must be less than or equal to 100"},
			})
		} else {
			params.Limit = limit
		}
	}

	if len(details) > 0 {
		return Params{}, apierror.Validation(details)
	}

	return params, nil
}

func NewMetadata(params Params, total int64) Metadata {
	totalPages := 0
	if total > 0 {
		totalPages = int((total + int64(params.Limit) - 1) / int64(params.Limit))
	}

	return Metadata{
		Page:       params.Page,
		Limit:      params.Limit,
		Total:      total,
		TotalPages: totalPages,
	}
}
