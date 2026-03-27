package validation

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"strings"

	"github.com/go-playground/validator/v10"

	"github.com/denis-maiorov-brightsec/go-net-http-ecommerce/internal/platform/apierror"
)

const defaultMaxBodyBytes int64 = 1 << 20

type Decoder struct {
	validator *validator.Validate
	maxBytes  int64
}

func NewDecoder() *Decoder {
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterTagNameFunc(func(field reflect.StructField) string {
		name := strings.Split(field.Tag.Get("json"), ",")[0]
		if name == "" {
			return field.Name
		}
		if name == "-" {
			return ""
		}

		return name
	})

	return &Decoder{
		validator: validate,
		maxBytes:  defaultMaxBodyBytes,
	}
}

func DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	return NewDecoder().DecodeJSON(w, r, dst)
}

func (d *Decoder) DecodeJSON(w http.ResponseWriter, r *http.Request, dst any) error {
	r.Body = http.MaxBytesReader(w, r.Body, d.maxBytes)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(dst); err != nil {
		return decodeError(err)
	}

	if err := decoder.Decode(&struct{}{}); err != nil && !errors.Is(err, io.EOF) {
		return apierror.Validation([]apierror.Detail{{
			Field:       "body",
			Constraints: []string{"request body must contain a single JSON object"},
		}})
	}

	if err := d.validator.Struct(dst); err != nil {
		return validationError(err)
	}

	return nil
}

func validationError(err error) error {
	var validationErrs validator.ValidationErrors
	if !errors.As(err, &validationErrs) {
		return apierror.Internal(err)
	}

	details := make([]apierror.Detail, 0, len(validationErrs))
	for _, validationErr := range validationErrs {
		field := validationErr.Field()
		if field == "" {
			field = strings.ToLower(validationErr.StructField())
		}

		details = append(details, apierror.Detail{
			Field:       field,
			Constraints: []string{constraintMessage(field, validationErr)},
		})
	}

	return apierror.Validation(details)
}

func decodeError(err error) error {
	var syntaxErr *json.SyntaxError
	if errors.As(err, &syntaxErr) {
		return apierror.Validation([]apierror.Detail{{
			Field:       "body",
			Constraints: []string{"request body must be valid JSON"},
		}})
	}

	var typeErr *json.UnmarshalTypeError
	if errors.As(err, &typeErr) {
		field := typeErr.Field
		if field == "" {
			field = "body"
		}

		return apierror.Validation([]apierror.Detail{{
			Field:       field,
			Constraints: []string{fmt.Sprintf("%s must be of type %s", field, typeErr.Type.String())},
		}})
	}

	var maxBytesErr *http.MaxBytesError
	if errors.As(err, &maxBytesErr) {
		return apierror.Validation([]apierror.Detail{{
			Field:       "body",
			Constraints: []string{fmt.Sprintf("request body must be smaller than %d bytes", maxBytesErr.Limit)},
		}})
	}

	if errors.Is(err, io.EOF) {
		return apierror.Validation([]apierror.Detail{{
			Field:       "body",
			Constraints: []string{"request body must not be empty"},
		}})
	}

	if strings.HasPrefix(err.Error(), "json: unknown field ") {
		field := strings.Trim(strings.TrimPrefix(err.Error(), "json: unknown field "), "\"")
		return apierror.Validation([]apierror.Detail{{
			Field:       field,
			Constraints: []string{fmt.Sprintf("%s is not allowed", field)},
		}})
	}

	return apierror.Validation([]apierror.Detail{{
		Field:       "body",
		Constraints: []string{"request body must be valid JSON"},
	}})
}

func constraintMessage(field string, validationErr validator.FieldError) string {
	switch validationErr.Tag() {
	case "required":
		return fmt.Sprintf("%s must not be empty", field)
	case "min":
		return fmt.Sprintf("%s must be at least %s", field, validationErr.Param())
	case "max":
		return fmt.Sprintf("%s must be at most %s", field, validationErr.Param())
	default:
		return fmt.Sprintf("%s failed %s validation", field, validationErr.Tag())
	}
}
