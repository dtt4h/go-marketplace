package pgutil

import (
	"errors"
	"mime"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
)

func IsUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		return pgErr.Code == "23505"
	}
	return err != nil && strings.Contains(err.Error(), "23505")
}

func UniqueViolationConstraint(err error) string {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return pgErr.ConstraintName
	}
	return ""
}

func IsNoRows(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

func NullText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func TextToPtr(t pgtype.Text) *string {
	if t.Valid {
		return &t.String
	}
	return nil
}

func NullInt8(v *int64) pgtype.Int8 {
	if v == nil {
		return pgtype.Int8{Valid: false}
	}
	return pgtype.Int8{Int64: *v, Valid: true}
}

func Int8ToPtr(v pgtype.Int8) *int64 {
	if v.Valid {
		return &v.Int64
	}
	return nil
}

func ToNumeric(v *string) (pgtype.Numeric, error) {
	if v == nil {
		return pgtype.Numeric{Valid: false}, nil
	}
	var n pgtype.Numeric
	if err := n.Scan(*v); err != nil {
		return pgtype.Numeric{}, err
	}
	return n, nil
}

func GenerateUploadKey(contentType string) string {
	ext := ".bin"
	if exts, _ := mime.ExtensionsByType(contentType); len(exts) > 0 && exts[0] != "" {
		ext = exts[0]
	}
	return uuid.New().String() + ext
}
