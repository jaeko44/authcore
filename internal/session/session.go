package session

import (
	"context"
	"strconv"
	"time"

	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/cryptoutil"
	"authcore.io/authcore/pkg/nulls"

	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

// Session represents an authenticated or unauthenticated session.
type Session struct {
	ID                     int64        `db:"id"`
	UserID                 int64        `db:"user_id" validate:"min=1"`
	ClientID               nulls.String `db:"client_id"`
	DeviceID               nulls.Int64  `db:"device_id"` // TODO: Device ID should be required
	IsMachine              bool         `db:"is_machine"`
	RefreshTokenHash       string       `db:"refresh_token"`
	RefreshToken           string       // Stores plaintext token for new session
	LastSeenAt             time.Time    `db:"last_seen_at"`
	LastSeenIP             string       `db:"last_seen_ip" validate:"omitempty,ip"`
	LastSeenLocation       string       `db:"last_seen_location"`
	LastPasswordVerifiedAt nulls.Time   `db:"last_password_verified_at"`
	UserAgent              string       `db:"user_agent"`
	IsInvalid              bool         `db:"is_invalid"`
	ExpiredAt              time.Time    `db:"expired_at"`
	UpdatedAt              time.Time    `db:"updated_at"`
	CreatedAt              time.Time    `db:"created_at"`
}

// Validate validates if a given session does not contains invalid data.
func (s *Session) Validate() error {
	if len(s.RefreshTokenHash) == 0 {
		return errors.New(errors.ErrorInvalidArgument, "")
	}

	return validator.Validate.Struct(s)
}

// IsExpired returns whether the session is expired.
func (s *Session) IsExpired() bool {
	return s.ExpiredAt.Before(time.Now())
}

// VerifyRefreshToken verifies the given token
func (s *Session) VerifyRefreshToken(refreshToken string) bool {
	return s.RefreshTokenHash == computeRefreshTokenHash(refreshToken)
}

// SetRefreshToken changes the refresh token to the given value.
func (s *Session) SetRefreshToken(refreshToken string) {
	s.RefreshToken = refreshToken
	s.RefreshTokenHash = computeRefreshTokenHash(refreshToken)
}

// Refresh refreshes the session and optionally generates new refresh token.
func (s *Session) Refresh(ctx context.Context, newRefreshToken bool) string {
	expiresIn := viper.GetDuration("session_expires_in")

	// Getting IP address from grpc is deprecated, keep it for compatibility
	fromMD, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ipAddress := fromMD.Get("ip-address")
		if len(ipAddress) > 0 {
			s.LastSeenIP = ipAddress[0]
		}
	}

	s.LastSeenAt = time.Now()
	s.LastSeenLocation = "null"

	s.ExpiredAt = time.Now().Add(expiresIn)

	refreshToken := ""
	if newRefreshToken {
		refreshToken = cryptoutil.RandomToken32()
		s.SetRefreshToken(refreshToken)
	}
	return refreshToken
}

// UpdateLastSeen updates the last seen metadata in the session according to the given ctx.
func (s *Session) UpdateLastSeen(ctx context.Context) {
	// Getting IP address from grpc is deprecated, keep it for compatibility
	fromMD, ok := metadata.FromIncomingContext(ctx)
	if ok {
		ipAddress := fromMD.Get("ip-address")
		if len(ipAddress) > 0 {
			s.LastSeenIP = ipAddress[0]
		}
	}

	s.LastSeenAt = time.Now()
	s.LastSeenLocation = "null"
}

// PublicID returns a textual unique identifier of the session.
func (s *Session) PublicID() string {
	return strconv.FormatInt(s.ID, 10)
}

// PublicUserID returns a textual unique identifier of the user for the session.
func (s *Session) PublicUserID() string {
	return strconv.FormatInt(s.UserID, 10)
}

// UpdateLastPasswordVerifiedAt updates the last password verification timestamp.
func (s *Session) UpdateLastPasswordVerifiedAt() {
	s.LastPasswordVerifiedAt = nulls.NewTime(time.Now())
}

// UpdateCurrentUserPasswordAllowed returns whether the session allow changing user password.
func (s *Session) UpdateCurrentUserPasswordAllowed() bool {
	recentTime := time.Now().Add(-5 * time.Minute)
	return s.LastPasswordVerifiedAt.Valid && s.LastPasswordVerifiedAt.Time.After(recentTime)
}
