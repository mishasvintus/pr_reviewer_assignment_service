package unit_tests

import (
	"database/sql/driver"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/mishasvintus/avito_backend_internship/internal/domain"
)

func TestPRStatus_NewPRStatus(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      domain.PRStatus
		wantError bool
	}{
		{
			name:      "valid - OPEN",
			input:     "OPEN",
			want:      domain.StatusOpen,
			wantError: false,
		},
		{
			name:      "valid - MERGED",
			input:     "MERGED",
			want:      domain.StatusMerged,
			wantError: false,
		},
		{
			name:      "invalid - empty string",
			input:     "",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid - random string",
			input:     "INVALID",
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid - lowercase",
			input:     "open",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			status, err := domain.NewPRStatus(tt.input)

			if tt.wantError {
				assert.Error(t, err)
				assert.Empty(t, status)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, status)
			}
		})
	}
}

func TestPRStatus_IsValid(t *testing.T) {
	tests := []struct {
		name   string
		status domain.PRStatus
		want   bool
	}{
		{
			name:   "valid - OPEN",
			status: domain.StatusOpen,
			want:   true,
		},
		{
			name:   "valid - MERGED",
			status: domain.StatusMerged,
			want:   true,
		},
		{
			name:   "invalid - empty",
			status: "",
			want:   false,
		},
		{
			name:   "invalid - random",
			status: "INVALID",
			want:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.status.IsValid())
		})
	}
}

func TestPRStatus_Scan(t *testing.T) {
	tests := []struct {
		name      string
		input     interface{}
		want      domain.PRStatus
		wantError bool
	}{
		{
			name:      "valid - string OPEN",
			input:     "OPEN",
			want:      domain.StatusOpen,
			wantError: false,
		},
		{
			name:      "valid - string MERGED",
			input:     "MERGED",
			want:      domain.StatusMerged,
			wantError: false,
		},
		{
			name:      "valid - []byte OPEN",
			input:     []byte("OPEN"),
			want:      domain.StatusOpen,
			wantError: false,
		},
		{
			name:      "invalid - nil",
			input:     nil,
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid - int",
			input:     123,
			want:      "",
			wantError: true,
		},
		{
			name:      "invalid - invalid string",
			input:     "INVALID",
			want:      "",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var status domain.PRStatus
			err := status.Scan(tt.input)

			if tt.wantError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.want, status)
			}
		})
	}
}

func TestPRStatus_Value(t *testing.T) {
	tests := []struct {
		name      string
		status    domain.PRStatus
		want      driver.Value
		wantError bool
	}{
		{
			name:      "valid - OPEN",
			status:    domain.StatusOpen,
			want:      "OPEN",
			wantError: false,
		},
		{
			name:      "valid - MERGED",
			status:    domain.StatusMerged,
			want:      "MERGED",
			wantError: false,
		},
		{
			name:      "invalid - empty",
			status:    "",
			want:      nil,
			wantError: true,
		},
		{
			name:      "invalid - invalid status",
			status:    "INVALID",
			want:      nil,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value, err := tt.status.Value()

			if tt.wantError {
				assert.Error(t, err)
				assert.Nil(t, value)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, value)
			}
		})
	}
}
