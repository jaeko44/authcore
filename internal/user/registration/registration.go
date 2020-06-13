package registration

import (
	"context"

	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/user"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"google.golang.org/grpc/metadata"
)

// RegisterUser function register the account into the system. Use this method instead of CreateUser to complete the
// user registration.
func RegisterUser(ctx context.Context, userStore *user.Store, sessionStore *session.Store, emailService *email.Service, smsService *sms.Service, u *user.User, clientID string, sendVerification, createSession bool) (*session.Session, error) {
	err := userStore.InsertUser(ctx, u)
	if err != nil {
		return nil, err
	}

	if u.Email.Valid && sendVerification {
		// Get origin from the context
		incomingCtx, _ := metadata.FromIncomingContext(ctx)
		origin := incomingCtx["grpcgateway-origin"][0]
		// Send verification email
		closedLoopCodePartialKeyForEmail := userStore.GetClosedLoopCodePartialKeyByUserIDAndContactValue(u.ID, u.Email.String)
		verificationCodeExpiryForEmail := viper.GetDuration("contact_verification_expiry_for_email")
		closedLoopCodeForEmail, err := userStore.CreateClosedLoopCode(ctx, closedLoopCodePartialKeyForEmail, verificationCodeExpiryForEmail)
		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{
			"user id": u.ID,
			"email":   u.Email.String,
		}).Info("send email")

		err = emailService.SendVerificationMail(
			ctx,
			origin,
			u.DisplayName(),
			u.Email.String,
			u.Language.String,
			closedLoopCodeForEmail.Code,
			closedLoopCodeForEmail.Token,
		)
		if err != nil {
			return nil, err
		}
	}

	if u.Phone.Valid && sendVerification {
		// Send verification SMS
		closedLoopCodePartialKeyForPhone := userStore.GetClosedLoopCodePartialKeyByUserIDAndContactValue(u.ID, u.Phone.String)
		verificationCodeExpiryForPhone := viper.GetDuration("contact_verification_expiry_for_phone")
		closedLoopCodeForPhone, err := userStore.CreateClosedLoopCode(ctx, closedLoopCodePartialKeyForPhone, verificationCodeExpiryForPhone)
		if err != nil {
			return nil, err
		}

		log.WithFields(log.Fields{
			"user id": u.ID,
			"phone":   u.Phone.String,
		}).Info("send sms")

		err = smsService.SendVerificationSMS(
			ctx,
			u.DisplayName(),
			u.Phone.String,
			closedLoopCodeForPhone.Code,
		)
		if err != nil {
			return nil, err
		}
	}

	log.WithFields(log.Fields{
		"username": u.Username,
		"email":    u.Email,
		"phone":    u.Phone,
	}).Info("user registration success")

	var session *session.Session
	if createSession {
		session, err = sessionStore.CreateSession(ctx, u.ID, 0, clientID, "", false)
		if err != nil {
			log.WithFields(log.Fields{
				"user_id": u.PublicID(),
				"error":   err,
			}).Info("failed to update session")
			return nil, err
		}
		log.WithFields(log.Fields{
			"user_id":    u.PublicID(),
			"session_id": session.PublicID(),
		}).Info("session created for user")
	}

	return session, nil
}
