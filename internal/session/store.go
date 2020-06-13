package session

import (
	"context"
	"crypto"
	"crypto/ecdsa"
	"database/sql"
	"encoding/base64"
	"fmt"
	"strconv"

	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"
	"authcore.io/authcore/pkg/secret"
	"gopkg.in/square/go-jose.v2"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/go-redis/redis"
	"github.com/jmoiron/sqlx"
	"github.com/mssola/user_agent"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Store manages Session models.
type Store struct {
	db        *db.DB
	redis     *redis.Client
	userStore *user.Store

	accessTokenPrivateKey   *ecdsa.PrivateKey
	accessTokenPublicKey    *ecdsa.PublicKey
	serviceAccountPublicKey *ecdsa.PublicKey
	serviceAccountsMap         map[string]ServiceAccount
}

// NewStore retrusn a new Store instance.
func NewStore(db *db.DB, redis *redis.Client, userStore *user.Store) *Store {
	s := &Store{
		db:        db,
		redis:     redis,
		userStore: userStore,
	}
	s.initPrivateKeys()
	s.initServiceAccounts()
	return s
}

// CreateSession creates new authenticated session and saves it to database.
func (s *Store) CreateSession(ctx context.Context, userID, deviceID int64, clientID, refreshToken string, passwordVerified bool) (*Session, error) {
	var userAgentValue string
	// Get the agent from context
	userAgent, ok := ctx.Value(UserAgentKey{}).(*user_agent.UserAgent)
	if ok {
		osInfo := userAgent.OSInfo()
		browserName, browserVersion := userAgent.Browser()
		// Agent format as {OS Name} {OS Version} {Browser name} {Browser version}
		userAgentValue = fmt.Sprintf("%s %s %s %s", osInfo.Name, osInfo.Version, browserName, browserVersion)
	} else {
		userAgentValue = "null"
	}
	// Get the IP from context
	ip, ok := ctx.Value(IPKey{}).(string)
	if !ok {
		ip = ""
	}
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	session := &Session{
		UserID:     userID,
		ClientID:   nulls.NewString(clientApp.ID),
		DeviceID:   nulls.NewInt64(deviceID),
		IsMachine:  false,
		LastSeenIP: ip,
		UserAgent:  userAgentValue,
	}
	if len(refreshToken) > 0 {
		session.SetRefreshToken(refreshToken)
		session.Refresh(ctx, false)
	} else {
		refreshToken = session.Refresh(ctx, true)
	}

	if passwordVerified {
		session.UpdateLastPasswordVerifiedAt()
	}

	id, err := s.createSession(ctx, session)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	session, err = s.FindSessionByInternalID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Save refresh token in the returning session
	session.RefreshToken = refreshToken

	s.userStore.UpdateUserLastSeenAt(ctx, userID, session.LastSeenAt)

	return session, nil
}

// CreateMachineSession creates new machine-to-machine session and saves it to database.
func (s *Store) CreateMachineSession(ctx context.Context, userID int64) (*Session, error) {
	session := &Session{
		UserID:     userID,
		DeviceID:   nulls.Int64{},
		IsMachine:  true,
		LastSeenIP: "",
	}
	refreshToken := session.Refresh(ctx, true)

	id, err := s.createSession(ctx, session)
	if err != nil {
		return nil, err
	}

	session, err = s.FindSessionByInternalID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Save refresh token in the returning session
	session.RefreshToken = refreshToken

	return session, nil
}

func (s *Store) createSession(ctx context.Context, session *Session) (int64, error) {
	err := session.Validate()
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	result, err := sqlx.NamedExecContext(
		ctx,
		s.db,
		`INSERT INTO sessions (
			user_id,
			client_id,
			device_id,
			is_machine,
			refresh_token,
			last_seen_at,
			last_seen_location,
			last_seen_ip,
			user_agent,
			expired_at,
			last_password_verified_at
		) VALUES (
			:user_id,
			:client_id,
			:device_id,
			:is_machine,
			:refresh_token,
			:last_seen_at,
			:last_seen_location,
			:last_seen_ip,
			:user_agent,
			:expired_at,
			:last_password_verified_at
		)`,
		&session)
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	id, err := result.LastInsertId()
	if err != nil {
		return 0, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return id, nil
}

// UpdateSession updates the given session record in database.
func (s *Store) UpdateSession(ctx context.Context, session *Session) (*Session, error) {
	_, err := sqlx.NamedExecContext(
		ctx,
		s.db,
		`UPDATE sessions SET
			refresh_token=:refresh_token,
			last_seen_at=:last_seen_at,
			last_seen_location=:last_seen_location,
			last_seen_ip=:last_seen_ip,
			last_password_verified_at=:last_password_verified_at
		WHERE id=:id`, &session)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	session, err = s.FindSessionByInternalID(ctx, session.ID)
	if err != nil {
		return nil, err
	}

	err = s.userStore.UpdateUserLastSeenAt(ctx, session.UserID, session.LastSeenAt)
	if err != nil {
		return nil, err
	}

	return session, err
}

// FindSessionByInternalID lookups a authenticated session primary ID.
func (s *Store) FindSessionByInternalID(ctx context.Context, id int64) (*Session, error) {
	session := &Session{}
	err := s.db.QueryRowxContext(ctx, "SELECT * FROM sessions WHERE id = ? AND is_invalid = 0 AND expired_at > NOW()", id).StructScan(session)
	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return session, nil
}

// FindSessionByPublicID lookups a authenticated session public ID.
func (s *Store) FindSessionByPublicID(ctx context.Context, publicID string) (*Session, error) {
	internalID, err := strconv.ParseInt(publicID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return s.FindSessionByInternalID(ctx, internalID)
}

// FindSessionByRefreshToken lookups an authenticated session with a refresh token.
func (s *Store) FindSessionByRefreshToken(ctx context.Context, refreshToken string) (*Session, error) {
	refreshTokenHash := computeRefreshTokenHash(refreshToken)
	session := &Session{}
	err := s.db.QueryRowxContext(ctx,
		"SELECT * FROM sessions WHERE refresh_token=? AND is_invalid = 0 AND expired_at > NOW()",
		refreshTokenHash).StructScan(session)
	if err == sql.ErrNoRows {
		return nil, errors.New(errors.ErrorNotFound, "")
	} else if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return session, nil
}

// FindAllSessionsByUser lookups all active (not expired, not invalidated) Sessions by certain user in the database. If userPublicID is empty string it finds all data in the database.
func (s *Store) FindAllSessionsByUser(ctx context.Context, pageOptions paging.PageOptions, userPublicID string) (*[]Session, *paging.Page, error) {
	pageOptions.UniqueColumn = "id"
	var page *paging.Page
	var err error

	sessions := []Session{}
	if userPublicID == "" {
		page, err = paging.SelectContext(ctx, s.db, pageOptions, &sessions, "SELECT sessions.* FROM sessions LEFT JOIN users ON users.id = sessions.user_id WHERE sessions.is_invalid = 0 AND sessions.expired_at > NOW()")
	} else {
		userID, err := strconv.ParseInt(userPublicID, 10, 64)
		if err != nil {
			return nil, nil, errors.Wrap(err, errors.ErrorUnknown, "")
		}
		page, err = paging.SelectContext(ctx, s.db, pageOptions, &sessions, "SELECT sessions.* FROM sessions LEFT JOIN users ON users.id = sessions.user_id WHERE users.id = ? AND sessions.is_invalid = 0 AND sessions.expired_at > NOW()", userID)
	}

	if err != nil {
		return nil, nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	return &sessions, page, nil
}

// InvalidateSessionByID invalidates a session by id
func (s *Store) InvalidateSessionByID(ctx context.Context, id int64) (int64, error) {
	_, err := s.FindSessionByInternalID(ctx, id)
	if err != nil {
		return -1, err
	}

	_, err = s.db.QueryxContext(ctx, "UPDATE sessions set is_invalid = 1 WHERE id = ?", id)
	if err != nil {
		return -1, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return id, nil
}

// GenerateAccessToken generates a new JWT token that assert the given session. If
// idTokenData is not nil, an ID Token will be generated with the given user information.
func (s *Store) GenerateAccessToken(ctx context.Context, session *Session, idToken bool) (AccessToken, error) {
	u := &user.User{ID: session.UserID}
	err := s.userStore.SelectUser(ctx, u)
	if err != nil {
		return AccessToken{}, err
	}
	userID := u.PublicID()
	if u.IsCurrentlyLocked() {
		return AccessToken{}, errors.New(errors.ErrorPermissionDenied, "cannot generate access token for a locked user")
	}
	if !idToken {
		// clear it to skip id token
		u = nil
	}
	return generateAccessToken(s.accessTokenPrivateKey, userID, session.PublicID(), session.ClientID.String, u)
}

// VerifyAccessToken verifies the signature of a give JWT token and returns the asserted userID and sessionID.
func (s *Store) VerifyAccessToken(ctx context.Context, token string) (userID string, sessionID string, err error) {
	return verifyAccessToken(s.accessTokenPublicKey, s.serviceAccountsMap, token)
}

// AccessTokenPublicKey gets the access token public key in jose.JSONWebKey.
func (s *Store) AccessTokenPublicKey() (*jose.JSONWebKey, error) {
	kid, err := kidFromECPublicKey(s.accessTokenPublicKey)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	publicKey := jose.JSONWebKey{
		Key:       s.accessTokenPublicKey,
		Algorithm: "ES256",
		Use:       "sig",
		KeyID:     kid,
	}
	return &publicKey, nil
}

// AllServiceAccounts returns all service accounts.
func (s *Store) AllServiceAccounts() ([]ServiceAccount, error) {
	as := make([]ServiceAccount, 0, len(s.serviceAccountsMap))
	for _, v := range s.serviceAccountsMap {
		as = append(as, v)
	}
	return as, nil
}

// kidFromECPublicKey return kid from ecdsa public key.
func kidFromECPublicKey(key *ecdsa.PublicKey) (string, error) {
	publicKey := jose.JSONWebKey{
		Key:       key,
		Algorithm: "ES256",
		Use:       "sig",
	}
	keyhash, err := publicKey.Thumbprint(crypto.SHA256)
	if err != nil {
		return "", errors.Wrap(err, errors.ErrorUnknown, "")
	}
	return base64.RawURLEncoding.EncodeToString(keyhash), nil
}

func (s *Store) initPrivateKeys() {
	privateKeyPEM := viper.Get("access_token_private_key").(secret.String).SecretString()
	if privateKeyPEM != "" {
		privateKey, err := jwt.ParseECPrivateKeyFromPEM([]byte(privateKeyPEM))
		if err != nil {
			log.Fatalf("invalid access_token_private_key: %v", err)
		}
		s.accessTokenPrivateKey = privateKey
		s.accessTokenPublicKey = &privateKey.PublicKey
	}

	serviceAccountPublicKeyPEM := viper.GetString("service_account_public_key")
	if serviceAccountPublicKeyPEM != "" {
		publicKey, err := jwt.ParseECPublicKeyFromPEM([]byte(serviceAccountPublicKeyPEM))
		if err != nil {
			log.Fatalf("invalid service_account_public_key: %v", err)
		}
		s.serviceAccountPublicKey = publicKey
	}
}

func (s *Store) initServiceAccounts() {
	var err error
	s.serviceAccountsMap, err = LoadServiceAccounts()
	if err != nil {
		log.Fatalf("unable to load service accounts: %v", err)
	}

	for _, a := range s.serviceAccountsMap {
		kid, err := a.KeyID()
		if err != nil {
			log.Fatalf("invalid service account public key: %v", err)
		}
		log.WithFields(log.Fields{
			"id":    a.ID,
			"kid":   kid,
			"roles": a.Roles,
		}).Info("add service account")
	}
}
