package dto

import (
	"net/url"
	"strings"
	"time"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

type ListFilters struct {
	Status string
	From   *time.Time
	To     *time.Time
}

func ParseListFilters(values url.Values) (ListFilters, error) {
	filters := ListFilters{
		Status: strings.TrimSpace(values.Get("status")),
	}

	var details []apierror.Detail

	if rawFrom := values.Get("from"); rawFrom != "" {
		from, err := time.Parse(time.DateOnly, rawFrom)
		if err != nil {
			details = append(details, apierror.Detail{
				Field:       "from",
				Constraints: []string{"from must be a valid date in YYYY-MM-DD format"},
			})
		} else {
			from = from.UTC()
			filters.From = &from
		}
	}

	if rawTo := values.Get("to"); rawTo != "" {
		to, err := time.Parse(time.DateOnly, rawTo)
		if err != nil {
			details = append(details, apierror.Detail{
				Field:       "to",
				Constraints: []string{"to must be a valid date in YYYY-MM-DD format"},
			})
		} else {
			to = to.UTC().Add(24 * time.Hour)
			filters.To = &to
		}
	}

	if filters.From != nil && filters.To != nil && !filters.From.Before(*filters.To) {
		details = append(details, apierror.Detail{
			Field:       "from",
			Constraints: []string{"from must be on or before to"},
		})
	}

	if len(details) > 0 {
		return ListFilters{}, apierror.Validation(details)
	}

	return filters, nil
}
