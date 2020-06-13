package registration

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/spf13/viper"

	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/clientapp"
	"authcore.io/authcore/internal/email"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/internal/sms"
	"authcore.io/authcore/internal/user"
)

// APIv2 returns a function that registers API 2.0 endpoints with an Echo instance.
func APIv2(userStore *user.Store, sessionStore *session.Store, emailService *email.Service, smsService *sms.Service) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{
			userStore:    userStore,
			sessionStore: sessionStore,
			emailService: emailService,
			smsService:   smsService,
		}

		g := e.Group("/api/v2")
		g.POST("/users", h.CreateUser)
	}
}

type handler struct {
	userStore    *user.Store
	sessionStore *session.Store
	emailService *email.Service
	smsService   *sms.Service
}

func (h *handler) CreateUser(c echo.Context) error {
	r := CreateUserRequest{}

	if err := c.Bind(&r); err != nil {
		return err
	}

	u := &user.User{
		Username: db.NullableString(r.Username),
		Email:    db.NullableString(r.Email),
		Phone:    db.NullableString(r.Phone),
		// Fixed value for language field. Support for deprecated API
		Language: db.NullableString(viper.GetStringSlice("available_languages")[0]),
	}

	if r.PasswordVerifier != nil {
		pv := r.PasswordVerifier

		if err := u.SetPasswordVerifier(pv.Salt, pv.VerifierW0, pv.VerifierL); err != nil {
			return err
		}
	}

	ctx := c.Request().Context()

	clientID := clientapp.AdminPortalClientID
	session, err := RegisterUser(ctx, h.userStore, h.sessionStore, h.emailService, h.smsService, u, clientID, false, true)
	if err != nil {
		return err
	}

	jsonUser, err := user.NewJSONUser(u)
	if err != nil {
		return err
	}

	resp := CreateUserResponse{
		User:         jsonUser,
		RefreshToken: session.RefreshToken,
	}


	return c.JSON(http.StatusOK, resp)
}

// CreateUserRequest represents a create user request.
type CreateUserRequest struct {
	Username         string                 `json:"username"`
	Email            string                 `json:"email"`
	Phone            string                 `json:"phone_number"`
	PasswordVerifier *user.PasswordVerifier `json:"verifier"`
}

// CreateUserResponse represents a create user response.
type CreateUserResponse struct {
	User         user.JSONUser `json:"user"`
	RefreshToken string        `json:"refresh_token"`
}
