package managementapi

import (
	"context"

	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/v1/authentication"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/rbac"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/template"
	"authcore.io/authcore/internal/user"
)

// Service provides GRPC Management API implementations.
type Service struct {
	DB                    *db.DB
	UserStore             *user.Store
	SessionStore          *session.Store
	AuthenticationService *authentication.Service
	AuditStore                 *audit.Store
	RBACService           *rbac.Service
	TemplateStore         *template.Store
	EmailService          *email.Service
	SMSService            *sms.Service
}

// NewService initialize a new Service.
func NewService(db *db.DB,
	userStore *user.Store,
	sessionStore *session.Store,
	authenticationService *authentication.Service,
	auditStore *audit.Store,
	rbacService *rbac.Service,
	templateStore *template.Store,
	emailService *email.Service,
	smsService *sms.Service) *Service {

	return &Service{
		DB:                    db,
		UserStore:             userStore,
		SessionStore:          sessionStore,
		AuthenticationService: authenticationService,
		AuditStore:                 auditStore,
		RBACService:           rbacService,
		TemplateStore:         templateStore,
		EmailService:          emailService,
		SMSService:            smsService,
	}
}

// authorize is a convenient method to authorize a request.
func (s *Service) authorize(ctx context.Context, permissions ...rbac.Permission) error {
	err := s.RBACService.Authorize(ctx, permissions...)
	if err != nil {
		return errors.Wrap(err, errors.ErrorPermissionDenied, "")
	}
	return nil
}
