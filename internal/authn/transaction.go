package authn

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"time"

	"authcore.io/authcore/internal/authn/idp"
	"authcore.io/authcore/internal/authn/verifier"
	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/cryptoutil"
	"authcore.io/authcore/pkg/log"
	"authcore.io/authcore/pkg/nulls"

	"github.com/jmoiron/sqlx"
	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// TransactionController executes authentication transactions.
type TransactionController struct {
	verifierFactory *verifier.Factory
	idpFactory      *idp.Factory
	db              *db.DB
	store           *Store
	userStore       *user.Store
	sessionStore    *session.Store
}

// NewTransactionController returns a new TransactionController.
func NewTransactionController(db *db.DB, store *Store, userStore *user.Store, sessionStore *session.Store) *TransactionController {
	return &TransactionController{
		verifierFactory: verifier.NewFactory(),
		idpFactory:      idp.NewFactory(),
		db:              db,
		store:           store,
		userStore:       userStore,
		sessionStore:    sessionStore,
	}
}

// StartPrimary starts an primary authentication transaction.
func (tc *TransactionController) StartPrimary(ctx context.Context, clientID, handle, redirectURI, codeChallengeMethod, codeChallenge, clientState string) (state *State, err error) {
	if handle == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "user handle cannot be empty")
	}
	if codeChallenge != "" && codeChallengeMethod != "S256" {
		return nil, errors.New(errors.ErrorInvalidArgument, "invalid code challenge method")
	}
	if err := ValidateRedirectURI(clientID, redirectURI); err != nil {
		return nil, err
	}

	u, err := tc.userStore.UserByHandle(ctx, handle)
	if err != nil {
		return
	}

	if u.IsCurrentlyLocked() {
		err = errors.New(errors.ErrorPermissionDenied, "user is locked")
		return
	}

	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return
	}

	state = &State{
		StateToken:          cryptoutil.RandomToken32(),
		Status:              StatusPrimary,
		ClientID:            clientApp.ID,
		UserID:              u.ID,
		Factors:             []string{},
		RedirectURI:         redirectURI,
		PKCEChallengeMethod: codeChallengeMethod,
		PKCEChallenge:       codeChallenge,
		ClientState:         clientState,
	}

	if u.IsPasswordAuthenticationEnabled() {
		verifier, err := u.PasswordVerifier()
		if err != nil {
			return nil, err
		}
		state.AppendFactor(FactorPassword)
		state.PasswordMethod = verifier.Method()
		state.PasswordSalt = verifier.Salt()
	}

	err = tc.store.CheckRateLimiter(ctx, u.ID)
	if err != nil {
		state.Status = StatusBlocked
	}

	err = tc.store.PutState(ctx, state)
	return
}

// RequestPassword performs a password key exchange
func (tc *TransactionController) RequestPassword(ctx context.Context, stateToken string, message []byte) (challenge verifier.Challenge, err error) {
	_, err = tc.stateMutation(ctx, stateToken, StatusPrimary, func(state *State, u *user.User) error {
		verifier, err := u.PasswordVerifier()
		if err != nil {
			return err
		}

		state.PasswordVerifierState, challenge, err = verifier.Request(message)
		return err
	})
	return
}

// VerifyPassword verifies the incoming password confirmation.
func (tc *TransactionController) VerifyPassword(ctx context.Context, stateToken string, in []byte) (*State, error) {
	return tc.stateMutation(ctx, stateToken, StatusPrimary, func(state *State, u *user.User) error {
		verifier, err := u.PasswordVerifier()
		if err != nil {
			return err
		}

		err = tc.store.CheckRateLimiter(ctx, u.ID)
		if err != nil {
			err = errors.New(errors.ErrorUserTemporarilyBlocked, "too many authentication attempts")
			return err
		}

		ok, _ := verifier.Verify(state.PasswordVerifierState, in)
		if !ok {
			log.GetLogger(ctx).WithFields(logrus.Fields{
				"user_id": u.PublicID(),
			}).Warn("password authentication rejected")
			tc.store.IncrementRateLimiter(ctx, u.ID)
			return errors.New(errors.ErrorPermissionDenied, "password incorrect")
		}
		log.GetLogger(ctx).WithFields(logrus.Fields{
			"user_id": u.PublicID(),
		}).Info("password authentication accepted")

		state.PasswordVerified = true

		secondFactors, err := tc.userStore.FindAllSecondFactorsByUserID(ctx, u.ID)
		if err != nil {
			return err
		}
		if !verifier.SkipMFA() && len(*secondFactors) > 0 {
			// MFA_REQUIRED
			state.Status = StatusMFARequired
			state.ClearFactors()
			for _, secondFactor := range *secondFactors {
				var factor string
				factor = secondFactor.Type.String()
				state.AppendFactor(factor)
			}
		} else {
			// SUCCESS
			tc.mutateSuccess(ctx, state)
		}
		return nil
	})
}

// RequestMFA requests a MFA challenge.
func (tc *TransactionController) RequestMFA(ctx context.Context, stateToken, method string, message []byte) (challenge verifier.Challenge, err error) {
	_, err = tc.stateMutation(ctx, stateToken, StatusMFARequired, func(state *State, u *user.User) error {
		factor, err := tc.getSecondFactor(ctx, u, method)
		if err != nil {
			return err
		}

		verifier, err := factor.ToVerifier(tc.verifierFactory)
		if err != nil {
			return err
		}

		state.MFAMethod = verifier.Method()
		state.MFAVerifierState, challenge, err = verifier.Request(message)
		if err != nil {
			return err
		}
		return nil
	})
	return
}

// VerifyMFA verifies a MFA factor.
func (tc *TransactionController) VerifyMFA(ctx context.Context, stateToken, method string, response []byte) (state *State, err error) {
	return tc.stateMutation(ctx, stateToken, StatusMFARequired, func(state *State, u *user.User) error {
		return tc.db.RunInTransaction(ctx, func(tx *sqlx.Tx) error {
			factor, err := tc.getSecondFactor(ctx, u, method)
			if err != nil {
				return err
			}

			err = tc.userStore.LockSecondFactor(ctx, factor)
			if err != nil {
				return err
			}

			v, err := factor.ToVerifier(tc.verifierFactory)
			if err != nil {
				return err
			}

			if state.MFAMethod != "" && state.MFAMethod != v.Method() {
				state.MFAVerifierState = nil
			}

			err = tc.store.CheckRateLimiter(ctx, u.ID)
			if err != nil {
				state.Status = StatusBlocked
				return nil
			}

			ok, updateVerifier := v.Verify(state.MFAVerifierState, response)
			if !ok {
				log.GetLogger(ctx).WithFields(logrus.Fields{
					"user_id": u.PublicID(),
				}).Warn("MFA authentication rejected")
				tc.store.IncrementRateLimiter(ctx, u.ID)
				return errors.New(errors.ErrorPermissionDenied, "MFA authentication rejected")
			}
			log.GetLogger(ctx).WithFields(logrus.Fields{
				"user_id": u.PublicID(),
			}).Info("MFA authentication accepted")

			_, err = tc.userStore.UpdateSecondFactorLastUsedAtByID(ctx, factor.ID)
			if err != nil {
				return err
			}

			if updateVerifier != nil {
				log.GetLogger(ctx).WithFields(logrus.Fields{
					"user_id": u.PublicID(),
				}).Info("updating second factor content")
				factor.UpdateWithVerifier(updateVerifier)
				_, err = tc.userStore.UpdateSecondFactorContent(ctx, factor)
				if err != nil {
					return err
				}
			}

			tx.Commit()

			// SUCCESS
			return tc.mutateSuccess(ctx, state)
		})
	})
}

// SignUp creates a new user.
func (tc *TransactionController) SignUp(ctx context.Context, clientID, redirectURI, email, phone, passwordVerifierJSON, name, lang string) (state *State, err error) {
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		err = errors.New(errors.ErrorInvalidArgument, "invalid client id")
		return
	}
	if !viper.GetBool("sign_up_enabled") {
		err = errors.New(errors.ErrorPermissionDenied, "create account is not allowed")
		return
	}
	if err = ValidateRedirectURI(clientID, redirectURI); err != nil {
		return
	}
	passwordVerifier, err := tc.verifierFactory.Unmarshal([]byte(passwordVerifierJSON))
	if err != nil {
		return
	}

	u := &user.User{
		DisplayNameOld: name,
		Name: nulls.String{
			String: name,
			Valid:  name != "",
		},
		Email: nulls.String{
			String: email,
			Valid:  email != "",
		},
		Phone: nulls.String{
			String: phone,
			Valid:  phone != "",
		},
		Language: nulls.String{
			String: lang,
			Valid:  lang != "",
		},
	}

	if err = u.UpdatePasswordWithVerifier(passwordVerifier); err != nil {
		return
	}

	err = tc.userStore.InsertUser(ctx, u)
	if err != nil {
		return
	}

	state = &State{
		StateToken:  cryptoutil.RandomToken32(),
		Status:      StatusSuccess,
		ClientID:    clientApp.ID,
		UserID:      u.ID,
		RedirectURI: redirectURI,
	}
	code := state.GenerateAuthorizationCode()
	if err = tc.store.PutAuthorizationCode(ctx, code); err != nil {
		return
	}
	if err = tc.store.PutState(ctx, state); err != nil {
		return
	}

	return
}

// RegisterVerifier adds an additional verifier method.
func (tc *TransactionController) RegisterVerifier(method string, unmarshaller verifier.Unmarshaller) {
	tc.verifierFactory.Register(method, unmarshaller)
}

// StartIDP starts a third-party ID provider authentication transaction.
func (tc *TransactionController) StartIDP(ctx context.Context, clientID, idpID, redirectURI, codeChallengeMethod, codeChallenge, clientState string) (state *State, err error) {
	if idpID == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "IDP cannot be empty")
	}
	if codeChallenge != "" && codeChallengeMethod != "S256" {
		return nil, errors.New(errors.ErrorInvalidArgument, "invalid code challenge method")
	}
	if err := ValidateRedirectURI(clientID, redirectURI); err != nil {
		return nil, err
	}
	provider, err := tc.idpFactory.IDP(idpID)
	if err != nil {
		return nil, errors.New(errors.ErrorNotFound, "invalid IDP")
	}
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "invalid client id")
	}
	stateToken := cryptoutil.RandomToken32()
	authorizationURL, idpState, err := provider.AuthorizationURL(stateToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}

	state = &State{
		StateToken:          stateToken,
		Status:              StatusIDP,
		ClientID:            clientApp.ID,
		IDP:                 provider.ID(),
		IDPState:            idpState,
		RedirectURI:         redirectURI,
		PKCEChallengeMethod: codeChallengeMethod,
		PKCEChallenge:       codeChallenge,
		IDPAuthorizationURL: authorizationURL,
		ClientState:         clientState,
	}

	err = tc.store.PutState(ctx, state)
	return
}

// VerifyIDP verifies an IDP authorization grant.
func (tc *TransactionController) VerifyIDP(ctx context.Context, stateToken, code string) (state *State, err error) {
	return tc.stateMutation(ctx, stateToken, StatusIDP, func(state *State, u *user.User) error {
		provider, err := tc.idpFactory.IDP(state.IDP)
		if err != nil {
			return err
		}

		grant, err := provider.Exchange(ctx, state.IDPState, code)
		if err != nil {
			return err
		}

		ident := grant.Identity
		if ident == nil || len(ident.ID) == 0 {
			return errors.New(errors.ErrorPermissionDenied, "invalid authorization grant")
		}

		var localUser *user.User
		oauthService, err := idp.IDToOAuthService(state.IDP)
		if err != nil {
			return err
		}
		oauthFactor, err := tc.userStore.FindOAuthFactorByOAuthIdentity(ctx, oauthService, ident.ID)
		if err != nil && !errors.IsKind(err, errors.ErrorNotFound) {
			return err
		} else if err != nil && errors.IsKind(err, errors.ErrorNotFound) {
			// No IDP binding is found. Check if we can create a new user
			if len(ident.Email) > 0 && ident.EmailVerified {
				u, err = tc.userStore.UserByEmail(ctx, ident.Email)
				if err != nil {
					if !errors.IsKind(err, errors.ErrorNotFound) {
						// a SQL error
						return err
					}
					// Create a new user with the identity below
				} else {
					// Email is used. Exit with StatusIDPAlreadyExists
					log.GetLogger(ctx).WithFields(logrus.Fields{
						"idp":     provider.ID(),
						"idp_id":  ident.ID,
						"user_id": u.PublicID(),
					}).Info("IDP user's email is found but it is not linked with the IDP")
					err = tc.mutateIDPAlreadyExists(ctx, state, u)
					if err != nil {
						return err
					}
					return nil
				}
			}
			if len(ident.PhoneNumber) > 0 && ident.PhoneNumberVerified {
				u, err = tc.userStore.UserByPhone(ctx, ident.PhoneNumber)
				if err != nil {
					if !errors.IsKind(err, errors.ErrorNotFound) {
						// a SQL error
						return err
					}
					// Create a new user with the identity below
				} else {
					// Phone is used. Exit with StatusIDPAlreadyExists
					log.GetLogger(ctx).WithFields(logrus.Fields{
						"idp":     provider.ID(),
						"idp_id":  ident.ID,
						"user_id": u.PublicID(),
					}).Info("IDP user's phone is found but it is not linked with the IDP")
					err = tc.mutateIDPAlreadyExists(ctx, state, u)
					if err != nil {
						return err
					}
					return nil
				}
			}

			if !viper.GetBool("sign_up_enabled") {
				return errors.New(errors.ErrorPermissionDenied, "create user is not allowed")
			}

			// FIXME: we need a verified email or phone number to register for now. It will result
			// in error if the IDP ident have no email and phone.
			emailVerifiedAt := nulls.Time{}
			if ident.EmailVerified {
				emailVerifiedAt = nulls.NewTime(time.Now())
			}
			phoneVerifiedAt := nulls.Time{}
			if ident.PhoneNumberVerified {
				phoneVerifiedAt = nulls.NewTime(time.Now())
			}
			localUser = &user.User{
				DisplayNameOld: ident.Name,
				Name: nulls.String{
					String: ident.Name,
					Valid:  ident.Name != "",
				},
				Email: nulls.String{
					String: ident.Email,
					Valid:  ident.Email != "",
				},
				EmailVerifiedAt: emailVerifiedAt,
				Phone: nulls.String{
					String: ident.PhoneNumber,
					Valid:  ident.PhoneNumber != "",
				},
				PhoneVerifiedAt: phoneVerifiedAt,
			}

			err := tc.userStore.InsertUser(ctx, localUser)
			if err != nil {
				return err
			}

			oauthService, err := idp.IDToOAuthService(state.IDP)
			if err != nil {
				return err
			}

			_, err = tc.userStore.CreateOAuthFactor(ctx, localUser.ID, oauthService, ident.ID, nulls.NewJSON(ident))
			if err != nil {
				return err
			}

			log.GetLogger(ctx).WithFields(logrus.Fields{
				"idp":     provider.ID(),
				"idp_id":  ident.ID,
				"user_id": localUser.ID,
			}).Info("register new IDP user")
		} else {
			// IDP binding is found. Sign in with the user
			localUser, err = tc.getUser(ctx, oauthFactor.UserID)
			if err != nil {
				return err
			}
		}

		// Successfuly login to a registered user
		log.GetLogger(ctx).WithFields(logrus.Fields{
			"idp":     provider.ID(),
			"idp_id":  ident.ID,
			"user_id": localUser.PublicID(),
		}).Info("IDP authentication accepted")
		state.UserID = localUser.ID
		tc.mutateSuccess(ctx, state)
		return nil
	})
}

// StartIDPBinding starts a third-party ID provider binding transaction.
func (tc *TransactionController) StartIDPBinding(ctx context.Context, userID int64, clientID, idpID, redirectURI string) (state *State, err error) {
	if idpID == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "IDP cannot be empty")
	}
	provider, err := tc.idpFactory.IDP(idpID)
	if err != nil {
		return nil, errors.New(errors.ErrorNotFound, "invalid IDP")
	}
	if err := ValidateRedirectURI(clientID, redirectURI); err != nil {
		return nil, err
	}
	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return nil, errors.New(errors.ErrorInvalidArgument, "invalid client id")
	}
	stateToken := cryptoutil.RandomToken32()
	authorizationURL, idpState, err := provider.AuthorizationURL(stateToken)
	if err != nil {
		return nil, errors.Wrap(err, errors.ErrorUnknown, "")
	}
	state = &State{
		StateToken:          stateToken,
		UserID:              userID,
		Status:              StatusIDPBinding,
		ClientID:            clientApp.ID,
		IDP:                 provider.ID(),
		IDPState:            idpState,
		IDPAuthorizationURL: authorizationURL,
		RedirectURI:         redirectURI,
	}

	err = tc.store.PutState(ctx, state)
	return
}

// VerifyIDPBinding verifies an IDP authorization grant for a IDP binding transaction.
func (tc *TransactionController) VerifyIDPBinding(ctx context.Context, stateToken string, userID int64, clientID, code string) (state *State, err error) {
	return tc.stateMutation(ctx, stateToken, StatusIDPBinding, func(state *State, u *user.User) error {
		provider, err := tc.idpFactory.IDP(state.IDP)
		if err != nil {
			return err
		}

		// Make sure the user id and client ID from request match the state
		if userID != state.UserID {
			return errors.New(errors.ErrorPermissionDenied, "illegal state")
		}

		if clientID != state.ClientID {
			return errors.New(errors.ErrorPermissionDenied, "illegal state")
		}

		grant, err := provider.Exchange(ctx, state.IDPState, code)
		if err != nil {
			return err
		}

		ident := grant.Identity
		if ident == nil || len(ident.ID) == 0 {
			return errors.New(errors.ErrorPermissionDenied, "invalid authorization grant")
		}

		oauthService, err := idp.IDToOAuthService(state.IDP)
		if err != nil {
			return err
		}

		oauthFactors, err := tc.userStore.FindAllOAuthFactorsByUserIDAndService(ctx, state.UserID, oauthService)
		if err != nil {
			return err
		}
		if len(*oauthFactors) > 0 {
			return errors.New(errors.ErrorAlreadyExists, "user already bound to the same IDP")
		}

		_, err = tc.userStore.FindOAuthFactorByOAuthIdentity(ctx, oauthService, ident.ID)
		if err != nil && !errors.IsKind(err, errors.ErrorNotFound) {
			return err
		} else if err == nil {
			return errors.New(errors.ErrorAlreadyExists, "IDP user is bound to another user")
		}

		_, err = tc.userStore.CreateOAuthFactor(ctx, state.UserID, oauthService, ident.ID, nulls.NewJSON(ident))
		if err != nil {
			return err
		}

		// Successfuly login to a registered user
		log.GetLogger(ctx).WithFields(logrus.Fields{
			"idp":     provider.ID(),
			"idp_id":  ident.ID,
			"user_id": u.PublicID(),
		}).Info("IDP binding completed successfully")
		state.Status = StatusIDPBindingSuccess

		return nil
	})
}

// StartStepUp starts a step-up password verification transaction with an existing session.
func (tc *TransactionController) StartStepUp(ctx context.Context, sessionID int64) (state *State, err error) {
	sess, err := tc.sessionStore.FindSessionByInternalID(ctx, sessionID)
	if err != nil {
		return
	}
	u, err := tc.userStore.UserByID(ctx, sess.UserID)
	if err != nil {
		return
	}

	if u.IsCurrentlyLocked() {
		err = errors.New(errors.ErrorPermissionDenied, "user is locked")
		return
	}

	if !u.IsPasswordAuthenticationEnabled() {
		err = errors.New(errors.ErrorPermissionDenied, "password is not set")
		return
	}

	stateToken := cryptoutil.RandomToken32()
	state = &State{
		StateToken: stateToken,
		UserID:     u.ID,
		SessionID:  sess.ID,
		Status:     StatusStepUp,
		ClientID:   sess.ClientID.String,
	}

	verifier, err := u.PasswordVerifier()
	if err != nil {
		return nil, err
	}
	state.AppendFactor(FactorPassword)
	state.PasswordMethod = verifier.Method()
	state.PasswordSalt = verifier.Salt()

	err = tc.store.CheckRateLimiter(ctx, u.ID)
	if err != nil {
		state.Status = StatusBlocked
	}

	err = tc.store.PutState(ctx, state)
	return
}

// RequestPasswordStepUp performs a password key exchange for a step-up password verification transaction.
func (tc *TransactionController) RequestPasswordStepUp(ctx context.Context, stateToken string, sessionID int64, message []byte) (challenge verifier.Challenge, err error) {
	_, err = tc.stateMutation(ctx, stateToken, StatusStepUp, func(state *State, u *user.User) error {
		if sessionID != state.SessionID {
			return errors.New(errors.ErrorPermissionDenied, "illegal state")
		}

		verifier, err := u.PasswordVerifier()
		if err != nil {
			return err
		}

		state.PasswordVerifierState, challenge, err = verifier.Request(message)
		return err
	})
	return
}

// VerifyPasswordStepUp verifies the password to complete a step-up password verifiication transaction.
func (tc *TransactionController) VerifyPasswordStepUp(ctx context.Context, stateToken string, sessionID int64, in []byte) (state *State, err error) {
	return tc.stateMutation(ctx, stateToken, StatusStepUp, func(state *State, u *user.User) error {
		if sessionID != state.SessionID {
			return errors.New(errors.ErrorPermissionDenied, "illegal state")
		}

		sess, err := tc.sessionStore.FindSessionByInternalID(ctx, state.SessionID)
		if err != nil {
			return err
		}

		verifier, err := u.PasswordVerifier()
		if err != nil {
			return err
		}

		err = tc.store.CheckRateLimiter(ctx, u.ID)
		if err != nil {
			err = errors.New(errors.ErrorUserTemporarilyBlocked, "too many authentication attempts")
			return err
		}

		ok, _ := verifier.Verify(state.PasswordVerifierState, in)
		if !ok {
			log.GetLogger(ctx).WithFields(logrus.Fields{
				"user_id": u.PublicID(),
			}).Warn("password verification rejected")
			tc.store.IncrementRateLimiter(ctx, u.ID)
			return errors.New(errors.ErrorPermissionDenied, "password incorrect")
		}
		log.GetLogger(ctx).WithFields(logrus.Fields{
			"user_id": u.PublicID(),
		}).Info("password verification accepted")

		state.PasswordVerified = true
		state.Status = StatusStepUpSuccess

		sess.UpdateLastPasswordVerifiedAt()
		_, err = tc.sessionStore.UpdateSession(ctx, sess)
		return err
	})
}

// StartPasswordReset starts a password reset transaction.
func (tc *TransactionController) StartPasswordReset(ctx context.Context, clientID, handle string) (state *State, err error) {
	if handle == "" {
		return nil, errors.New(errors.ErrorInvalidArgument, "user handle cannot be empty")
	}

	u, err := tc.userStore.UserByHandle(ctx, handle)
	if err != nil {
		return
	}

	if u.IsCurrentlyLocked() {
		err = errors.New(errors.ErrorPermissionDenied, "user is locked")
		return
	}

	clientApp, err := clientapp.GetByClientID(clientID)
	if err != nil {
		return
	}

	stateToken := cryptoutil.RandomToken32()

	state = &State{
		StateToken: stateToken,
		Status:     StatusPasswordReset,
		ClientID:   clientApp.ID,
		UserID:     u.ID,
	}

	err = tc.store.CheckRateLimiter(ctx, u.ID)
	if err != nil {
		state.Status = StatusBlocked
		return
	}

	v, err := resetLinkVerifier(tc.verifierFactory, u, handle, clientApp.ID, stateToken)
	if err != nil {
		return
	}

	resetLinkState, _, err := v.Request(nil)
	if err != nil {
		return
	}
	state.ResetLinkState = resetLinkState

	err = tc.store.PutState(ctx, state)
	return
}

// VerifyPasswordReset verifies the reset token to complete a password reset transaction.
func (tc *TransactionController) VerifyPasswordReset(ctx context.Context, stateToken, resetToken, passwordVerifierJSON string) (state *State, err error) {
	return tc.stateMutation(ctx, stateToken, StatusPasswordReset, func(state *State, u *user.User) error {
		err = tc.store.CheckRateLimiter(ctx, u.ID)
		if err != nil {
			err = errors.New(errors.ErrorUserTemporarilyBlocked, "too many authentication attempts")
			return err
		}

		v, err := resetLinkVerifier(tc.verifierFactory, u, u.Email.String, state.ClientID, stateToken)
		if err != nil {
			return err
		}

		ok, _ := v.Verify(state.ResetLinkState, []byte(resetToken))
		if !ok {
			log.GetLogger(ctx).WithFields(logrus.Fields{
				"user_id": u.PublicID(),
			}).Warn("reset link rejected")
			tc.store.IncrementRateLimiter(ctx, u.ID)
			return errors.New(errors.ErrorPermissionDenied, "invalid reset link")
		}

		log.GetLogger(ctx).WithFields(logrus.Fields{
			"user_id": u.PublicID(),
		}).Info("reset link accepted")

		// Set password if a password verifier is defined. Otherwise, just check the token.
		if len(passwordVerifierJSON) > 0 {
			passwordVerifier, err := tc.verifierFactory.Unmarshal([]byte(passwordVerifierJSON))
			if err != nil {
				return err
			}
			if err := u.UpdatePasswordWithVerifier(passwordVerifier); err != nil {
				return err
			}
			if err := tc.userStore.UpdateUser(ctx, u); err != nil {
				return err
			}

			log.GetLogger(ctx).WithFields(logrus.Fields{
				"user_id": u.PublicID(),
			}).Info("reset password completed successfully")

			state.Status = StatusPasswordResetSuccess
		}
		return nil
	})
}

// RegisterIDP adds a third-party IDP
func (tc *TransactionController) RegisterIDP(idp idp.IDP) {
	tc.idpFactory.Register(idp)
}

// ExchangeSession exchanges an authorization code for a session
func (tc *TransactionController) ExchangeSession(ctx context.Context, clientID, redirectURI, code, codeVerifier string) (*session.Session, error) {
	authorizationCode, err := tc.store.GetAuthorizationCode(ctx, code)
	if err != nil {
		return nil, err
	}
	err = tc.store.DeleteAuthorizationCode(ctx, code)
	if err != nil {
		return nil, err
	}

	if authorizationCode.ClientID != clientID {
		return nil, errors.New(errors.ErrorPermissionDenied, "client_id mismatch")
	}

	if authorizationCode.RedirectURI != redirectURI {
		return nil, errors.New(errors.ErrorPermissionDenied, "redirect_uri mismatch")
	}

	if !isCodeChallengeValid(codeVerifier, authorizationCode.PKCEChallenge, authorizationCode.PKCEChallengeMethod) {
		return nil, errors.New(errors.ErrorPermissionDenied, "code_verifier mismatch")
	}

	_, err = tc.getUser(ctx, authorizationCode.UserID) // validates user
	if err != nil {
		return nil, err
	}

	refreshToken := cryptoutil.RandomToken32()
	return tc.sessionStore.CreateSession(ctx, authorizationCode.UserID, 0, authorizationCode.ClientID, refreshToken, authorizationCode.PasswordVerified)
}

func (tc *TransactionController) stateMutation(ctx context.Context, stateToken, expectStatus string, mutateFunc func(*State, *user.User) error) (state *State, err error) {
	state, err = tc.store.GetState(ctx, stateToken)
	if err != nil {
		if errors.IsKind(err, errors.ErrorNotFound) {
			err = errors.New(errors.ErrorPermissionDenied, "state token not found")
		}
		return
	}

	var u *user.User
	userID := ""
	if state.Status != StatusIDP {
		u, err = tc.userStore.UserByID(ctx, state.UserID)
		if err != nil {
			return
		}
		if u.IsCurrentlyLocked() {
			state.Status = StatusBlocked
		}
		userID = u.PublicID()
	}

	if state.Status != expectStatus {
		log.GetLogger(ctx).WithFields(logrus.Fields{
			"user_id":       userID,
			"status":        state.Status,
			"expect_status": expectStatus,
		}).Warn("illegal authn state")
		err = errors.New(errors.ErrorPermissionDenied, "illegal state")
		return
	}

	prevStatus := state.Status

	log.GetLogger(ctx).WithFields(logrus.Fields{
		"user_id": userID,
		"status":  state.Status,
	}).Debug("enter authn state mutation")

	err = mutateFunc(state, u)

	if err != nil {
		log.GetLogger(ctx).WithFields(logrus.Fields{
			"user_id": userID,
			"status":  state.Status,
			"error":   err.Error(),
		}).Warn("exit authn state mutation with error")
		return
	}

	log.GetLogger(ctx).WithFields(logrus.Fields{
		"state_token": stateToken,
		"status":      state.Status,
		"prev_status": prevStatus,
	}).Info("complete state mutation")

	err = tc.store.PutState(ctx, state)
	if err != nil {
		log.GetLogger(ctx).WithFields(logrus.Fields{
			"state_token": stateToken,
			"status":      state.Status,
			"error":       err.Error(),
		}).Error("error saving state mutation")
	}

	return
}

func (tc *TransactionController) getUser(ctx context.Context, id int64) (u *user.User, err error) {
	u, err = tc.userStore.UserByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if u.IsCurrentlyLocked() {
		return nil, errors.New(errors.ErrorPermissionDenied, "user is locked")
	}
	return u, nil
}

func (tc *TransactionController) getSecondFactor(ctx context.Context, u *user.User, method string) (factor *user.SecondFactor, err error) {
	secondFactorType, err := user.SecondFactorTypeFromString(method)
	if err != nil {
		return
	}
	factors, err := tc.userStore.FindAllSecondFactorsByUserIDAndType(ctx, u.ID, secondFactorType)
	if err != nil {
		return
	}
	if len(*factors) == 0 {
		err = errors.New(errors.ErrorPermissionDenied, "factor not found")
		return
	}
	if len(*factors) > 1 {
		err = errors.New(errors.ErrorPermissionDenied, "more than one factor found")
		return
	}
	factor = &(*factors)[0]
	return
}

func (tc *TransactionController) mutateSuccess(ctx context.Context, state *State) (err error) {
	if state.Status == StatusSuccess {
		err = errors.New(errors.ErrorPermissionDenied, "illegal state")
		return
	}
	log.GetLogger(ctx).WithFields(logrus.Fields{
		"user_id": state.UserID,
	}).Info("authentication completed successfully")
	state.Status = StatusSuccess
	code := state.GenerateAuthorizationCode()

	err = tc.store.PutAuthorizationCode(ctx, code)
	return
}

func (tc *TransactionController) mutateIDPAlreadyExists(ctx context.Context, state *State, u *user.User) error {
	// Search for OAuth factors to decide whether to notify the user to sign in with that factor or using password
	oauthFactors, err := tc.userStore.FindAllOAuthFactorsByUserID(ctx, u.ID)
	if err != nil {
		return err
	}
	state.ClearFactors()
	for _, oauthFactor := range *oauthFactors {
		identifier, err := idp.OAuthServiceToID(oauthFactor.Service)
		if err != nil {
			return err
		}
		state.AppendFactor("idp_" + identifier)
	}
	if u.IsPasswordAuthenticationEnabled() {
		state.AppendFactor(verifier.SPAKE2Plus)
	}
	state.Status = StatusIDPAlreadyExists
	return nil
}

func isCodeChallengeValid(codeVerifier, codeChallenge, codeChallengeMethod string) bool {
	// if challenge is not set, then assume that PKCE is not enabled
	if codeChallenge == "" {
		return true
	}
	switch codeChallengeMethod {
	case "plain":
		return false // Not implement as not recommended
	case "S256":
		verifier := []byte(codeVerifier)
		challenge, err := base64.RawURLEncoding.DecodeString(codeChallenge)
		if err != nil {
			return false
		}
		hashVerifier := sha256.Sum256(verifier)
		logrus.Printf("%v %v", codeChallenge, base64.RawURLEncoding.EncodeToString(hashVerifier[:]))
		return bytes.Compare(hashVerifier[:], challenge) == 0
	default:
		return false
	}
}

func resetLinkVerifier(factory *verifier.Factory, u *user.User, handle, clientID, stateToken string) (v verifier.Verifier, err error) {
	// It is quite hackish to create the validator this way. But there is no easier method to
	// create reset link verifier with factory.
	email := ""
	phoneNumber := ""
	if handle == u.Email.String {
		email = u.Email.String
	} else {
		phoneNumber = u.Phone.String
	}
	resetLinkVerifierJSON := map[string]string{
		"method":       "reset_link",
		"email":        email,
		"phone_number": phoneNumber,
		"state_token":  stateToken,
		"client_id":    clientID,
		"lang":         u.RealLanguage(),
	}
	resetLinkVerifierJSONBytes, err := json.Marshal(resetLinkVerifierJSON)
	if err != nil {
		return
	}
	v, err = factory.Unmarshal(resetLinkVerifierJSONBytes)
	return
}
