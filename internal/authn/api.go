package authn

import (
	"encoding/json"
	"net/http"
	"net/url"

	"authcore.io/authcore/internal/audit"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/session"
	"authcore.io/authcore/pkg/log"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"
	"github.com/sirupsen/logrus"
)

// APIv2 returns a function that registers Auth API 2.0 endpoints with an Echo instance.
func APIv2(tc *TransactionController, auditor audit.Auditor) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{tc: tc, auditor: auditor}

		g := e.Group("/api/v2")
		g.POST("/authn", h.StartPrimary)
		g.POST("/authn/password", h.RequestPassword)
		g.POST("/authn/password/verify", h.VerifyPassword)
		g.POST("/authn/mfa/:method", h.RequestMFA)
		g.POST("/authn/mfa/:method/verify", h.VerifyMFA)
		g.POST("/authn/idp/:provider", h.StartIDP)
		g.POST("/authn/idp/:provider/verify", h.VerifyIDP)
		g.POST("/authn/idp_binding/:provider", h.StartIDPBinding)
		g.POST("/authn/idp_binding/:provider/verify", h.VerifyIDPBinding)
		g.POST("/authn/step_up", h.StartStepUp)
		g.POST("/authn/step_up/password", h.RequestPasswordStepUp)
		g.POST("/authn/step_up/password/verify", h.VerifyPasswordStepUp)
		g.POST("/authn/password_reset", h.StartPasswordReset)
		g.POST("/authn/password_reset/verify", h.VerifyPasswordReset)
		g.POST("/signup", h.SignUp)
		g.POST("/authn/get_state", h.GetState)

		// Endpoints for handling third-party OAuth IDP. They are defined in this package because
		// they are part of the IDP authn flow.
		e.GET("/oauth/redirect", h.OauthRedirect)
		e.GET("/oauth/arbiter-redirect", h.OauthArbiterRedirect)
	}
}

type handler struct {
	tc      *TransactionController
	auditor audit.Auditor
}

func (h *handler) StartPrimary(c echo.Context) error {
	r := new(StartPrimaryRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.StartPrimary(c.Request().Context(), r.ClientID, r.Handle, r.RedirectURI, r.CodeChallengeMethod, r.CodeChallenge, r.ClientState)
	if err != nil {
		return err
	}
	return sendState(c, state)
}

func (h *handler) RequestPassword(c echo.Context) error {
	r := new(PasswordRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	challenge, err := h.tc.RequestPassword(c.Request().Context(), r.StateToken, r.Message)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &PasswordResponse{
		Challenge: challenge,
	})
}

func (h *handler) VerifyPassword(c echo.Context) error {
	r := new(VerifyPasswordRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.VerifyPassword(c.Request().Context(), r.StateToken, r.Verifier)
	if err != nil {
		return err
	}

	if state.Status == StatusSuccess {
		target := map[string]interface{}{"method": "password"}
		h.logStateAuditEvent(c, state, "user.authn", true, target)
	} else if state.Status == StatusBlocked {
		target := map[string]interface{}{"method": "password", "blocked": true}
		h.logStateAuditEvent(c, state, "user.authn", false, target)
	}

	return sendState(c, state)
}

func (h *handler) RequestMFA(c echo.Context) error {
	method := c.Param("method")
	r := new(MFARequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	challenge, err := h.tc.RequestMFA(c.Request().Context(), r.StateToken, method, r.Message)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &MFAResponse{
		Challenge: challenge,
	})
}

func (h *handler) VerifyMFA(c echo.Context) error {
	method := c.Param("method")
	r := new(VerifyMFARequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.VerifyMFA(c.Request().Context(), r.StateToken, method, r.Verifier)
	if err != nil {
		return err
	}

	if state.Status == StatusSuccess {
		target := map[string]interface{}{"method": "mfa"}
		h.logStateAuditEvent(c, state, "user.authn", true, target)
	} else if state.Status == StatusBlocked {
		target := map[string]interface{}{"method": "mfa", "blocked": true}
		h.logStateAuditEvent(c, state, "user.authn", false, target)
	}

	return sendState(c, state)
}

func (h *handler) StartIDP(c echo.Context) error {
	idpID := c.Param("provider")
	r := new(StartIDPRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.StartIDP(c.Request().Context(), r.ClientID, idpID, r.RedirectURI, r.CodeChallengeMethod, r.CodeChallenge, r.ClientState)
	if err != nil {
		return err
	}
	return sendState(c, state)
}

func (h *handler) VerifyIDP(c echo.Context) error {
	r := new(VerifyIDPRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	ctx := c.Request().Context()
	state, err := h.tc.VerifyIDP(ctx, r.StateToken, r.Code)
	if err != nil {
		return err
	}

	if state.Status == StatusSuccess {
		target := map[string]interface{}{"method": "idp", "provider": state.IDP}
		h.logStateAuditEvent(c, state, "user.authn", true, target)
	}

	return sendState(c, state)
}

func (h *handler) StartIDPBinding(c echo.Context) error {
	idpID := c.Param("provider")
	currentSess, ok := session.FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	r := new(StartIDPBindingRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.StartIDPBinding(c.Request().Context(), currentSess.UserID, currentSess.ClientID.String, idpID, r.RedirectURI)
	if err != nil {
		return err
	}
	return sendState(c, state)
}

func (h *handler) VerifyIDPBinding(c echo.Context) error {
	currentSess, ok := session.FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	r := new(VerifyIDPBindingRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.VerifyIDPBinding(c.Request().Context(), r.StateToken, currentSess.UserID, currentSess.ClientID.String, r.Code)
	if err != nil {
		return err
	}

	if state.Status == StatusSuccess {
		target := map[string]interface{}{"provider": state.IDP}
		h.logStateAuditEvent(c, state, "user.bind_idp", true, target)
	}

	return sendState(c, state)
}

func (h *handler) StartStepUp(c echo.Context) error {
	currentSess, ok := session.FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	state, err := h.tc.StartStepUp(c.Request().Context(), currentSess.ID)
	if err != nil {
		return err
	}
	return sendState(c, state)
}

func (h *handler) RequestPasswordStepUp(c echo.Context) error {
	currentSess, ok := session.FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	r := new(PasswordRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	challenge, err := h.tc.RequestPasswordStepUp(c.Request().Context(), r.StateToken, currentSess.ID, r.Message)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, &PasswordResponse{
		Challenge: challenge,
	})
}

func (h *handler) VerifyPasswordStepUp(c echo.Context) error {
	currentSess, ok := session.FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	r := new(VerifyPasswordRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.VerifyPasswordStepUp(c.Request().Context(), r.StateToken, currentSess.ID, r.Verifier)
	if err != nil {
		return err
	}

	if state.Status == StatusStepUpSuccess {
		target := map[string]interface{}{"method": "password"}
		h.logStateAuditEvent(c, state, "user.step_up_authn", true, target)
	} else if state.Status == StatusBlocked {
		target := map[string]interface{}{"method": "password", "blocked": true}
		h.logStateAuditEvent(c, state, "user.step_up_authn", false, target)
	}

	return sendState(c, state)
}

func (h *handler) StartPasswordReset(c echo.Context) error {
	r := new(StartPasswordResetRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.StartPasswordReset(c.Request().Context(), r.ClientID, r.Handle)
	if err != nil {
		return err
	}
	return sendState(c, state)
}

func (h *handler) VerifyPasswordReset(c echo.Context) error {
	r := new(VerifyPasswordResetRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	var verifierJSON []byte
	if len(r.PasswordVerifier) > 0 {
		var err error
		verifierJSON, err = json.Marshal(r.PasswordVerifier)
		if err != nil {
			return errors.New(errors.ErrorInvalidArgument, "invalid password_verifier")
		}
	}
	state, err := h.tc.VerifyPasswordReset(c.Request().Context(), r.StateToken, r.ResetToken, string(verifierJSON))
	if err != nil {
		return err
	}
	if state.Status == StatusPasswordResetSuccess {
		h.logStateAuditEvent(c, state, "user.password_reset", true, nil)
	}
	return sendState(c, state)
}

func (h *handler) GetState(c echo.Context) error {
	r := new(GetStateRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	state, err := h.tc.store.GetState(c.Request().Context(), r.StateToken)
	if err != nil {
		return err
	}
	return sendState(c, state)
}

func (h *handler) SignUp(c echo.Context) error {
	r := new(SignUpRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	verifierJSON, err := json.Marshal(r.PasswordVerifier)
	if err != nil {
		return errors.New(errors.ErrorInvalidArgument, "invalid password_verifier")
	}
	ctx := c.Request().Context()
	state, err := h.tc.SignUp(ctx, r.ClientID, r.RedirectURI, r.Email, r.Phone, string(verifierJSON), r.Name, r.Language)
	if err != nil {
		return err
	}

	h.logStateAuditEvent(c, state, "user.sign_up", true, nil)

	return sendState(c, state)
}

func (h *handler) OauthRedirect(c echo.Context) error {
	r := new(OauthRedirectRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}

	state, err := h.tc.store.GetState(c.Request().Context(), r.State)
	if err != nil {
		return errors.Wrap(err, errors.ErrorPermissionDenied, "state not found")
	}

	if state.Status != StatusIDP && state.Status != StatusIDPBinding {
		return errors.New(errors.ErrorPermissionDenied, "illegal state")
	}

	// Redirect to client-side arbiter to verify the code grant
	redirectURL, err := url.Parse("/widgets/oauth/arbiter")
	if err != nil {
		logrus.Fatal(err.Error())
	}
	q := redirectURL.Query()
	q.Add("clientId", state.ClientID)
	q.Add("state", r.State)
	q.Add("code", r.Code)
	q.Add("oauth_verifier", r.OauthVerifier)
	redirectURL.RawQuery = q.Encode()
	c.Redirect(http.StatusFound, redirectURL.String())
	return nil
}

// OauthArbiterRedirect endpoint for oauth arbiter widget to redirect to a destination URL. It only
// redirects if the URL is allowed according to the client_id.
func (h *handler) OauthArbiterRedirect(c echo.Context) error {
	r := new(OauthRedirectArbiterRequest)
	if err := c.Bind(r); err != nil {
		return err
	}
	if err := c.Validate(r); err != nil {
		return err
	}
	if err := ValidateRedirectURI(r.ClientID, r.RedirectURI); err != nil {
		return err
	}
	c.Redirect(http.StatusFound, r.RedirectURI)
	return nil
}

func (h *handler) logStateAuditEvent(c echo.Context, state *State, action string, success bool, target interface{}) {
	var actor audit.Actor
	var err error
	if state.UserID != 0 {
		ctx := c.Request().Context()
		actor, err = h.tc.userStore.UserByID(ctx, state.UserID)
		if err != nil {
			log.GetLogger(ctx).Errorf("error when writing audit log: %v", err)
			return
		}
	}
	var result audit.EventResult
	if success {
		result = audit.EventResultSuccess
	} else {
		result = audit.EventResultFail
	}
	h.auditor.LogEvent(c, actor, action, target, result)
}

// StartPrimaryRequest is the request for StartPrimary.
type StartPrimaryRequest struct {
	ClientID            string `json:"client_id"`
	Handle              string `json:"handle"`
	RedirectURI         string `json:"redirect_uri"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	CodeChallenge       string `json:"code_challenge"`
	ClientState         string `json:"client_state"`
}

// PasswordRequest is the request for RequestPassword.
type PasswordRequest struct {
	StateToken string `json:"state_token" validate:"required"`
	Message    []byte `json:"message"`
}

// VerifyPasswordRequest is the request for VerifyPassword.
type VerifyPasswordRequest struct {
	StateToken string `json:"state_token" validate:"required"`
	Verifier   []byte `json:"verifier" validate:"required"`
}

// MFARequest is the request for RequestMFA.
type MFARequest struct {
	StateToken string `json:"state_token" validate:"required"`
	Message    []byte `json:"message"`
}

// VerifyMFARequest is the request for VerifyMFA.
type VerifyMFARequest struct {
	StateToken string `json:"state_token" validate:"required"`
	Verifier   []byte `json:"verifier" validate:"required"`
}

// StartIDPRequest is the request for StartIDP.
type StartIDPRequest struct {
	ClientID            string `json:"client_id"`
	RedirectURI         string `json:"redirect_uri"`
	CodeChallengeMethod string `json:"code_challenge_method"`
	CodeChallenge       string `json:"code_challenge"`
	ClientState         string `json:"client_state"`
}

// VerifyIDPRequest is the request for VerifyIDP.
type VerifyIDPRequest struct {
	StateToken string `json:"state_token" validate:"required"`
	Code       string `json:"code" validate:"required"`
}

// StartIDPBindingRequest is the request for StartIDPBinding.
type StartIDPBindingRequest struct {
	RedirectURI string `json:"redirect_uri"`
}

// VerifyIDPBindingRequest is the request for VerifyIDP.
type VerifyIDPBindingRequest struct {
	StateToken string `json:"state_token" validate:"required"`
	Code       string `json:"code" validate:"required"`
}

// StartPasswordResetRequest is the request for StartPasswordReset.
type StartPasswordResetRequest struct {
	Handle   string `json:"handle"`
	ClientID string `json:"client_id"`
}

// VerifyPasswordResetRequest is the request for VerifyPasswordReset.
type VerifyPasswordResetRequest struct {
	StateToken       string                 `json:"state_token" validate:"required"`
	ResetToken       string                 `json:"reset_token" validate:"required"`
	PasswordVerifier map[string]interface{} `json:"password_verifier"`
}

// GetStateRequest is the request for GetState.
type GetStateRequest struct {
	StateToken string `json:"state_token" validate:"required"`
}

// OauthRedirectRequest is the request OauthRedirect.
type OauthRedirectRequest struct {
	State         string `query:"state" validate:"required"`
	Code          string `query:"code"`
	OauthVerifier string `query:"oauth_verifier"`
}

// OauthRedirectArbiterRequest is the request for OauthRedirectArbiter.
type OauthRedirectArbiterRequest struct {
	ClientID    string `query:"client_id" validate:"required"`
	RedirectURI string `query:"redirect_uri" validate:"required"`
}

// SignUpRequest is the request for SignUp.
type SignUpRequest struct {
	ClientID         string                 `json:"client_id" validate:"required"`
	RedirectURI      string                 `json:"redirect_uri" validate:"required"`
	PasswordVerifier map[string]interface{} `json:"password_verifier" validate:"required"`
	// check phone exist first. if phone exist then it omitempty and dont check for email format.
	// otherwise if phone not exist, it is required so it is not empty, and have to pass email check.
	// Similar logic for phone validator.
	Email    string `json:"email" validate:"required_without=Phone,omitempty,email"`
	Phone    string `json:"phone" validate:"required_without=Email,omitempty,phone"`
	Name     string `json:"name"`
	Language string `json:"language"`
}

// PasswordResponse is the response body for RequestPassword.
type PasswordResponse struct {
	Challenge []byte `json:"challenge"`
}

// MFAResponse is the response for RequestMFA
type MFAResponse struct {
	Challenge []byte `json:"challenge"`
}

// JSONState represents a AuthnState in Authn API.
type JSONState struct {
	StateToken          string   `json:"state_token" validate:"required"`
	Status              string   `json:"status" validate:"required"`
	PasswordMethod      string   `json:"password_method"`
	PasswordSalt        []byte   `json:"password_salt"`
	Factors             []string `json:"factors"`
	IDP                 string   `json:"idp"`
	IDPAuthorizationURL string   `json:"idp_authorization_url"`
	AuthorizationCode   string   `json:"authorization_code"`
	RedirectURI         string   `json:"redirect_uri"`
	ClientState         string   `json:"client_state"`
}

// NewJSONState converts a State into JSONState.
func NewJSONState(state *State) (JSONState, error) {
	j := JSONState{}
	err := copier.Copy(&j, state)
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}

func sendState(c echo.Context, state *State) error {
	resp, err := NewJSONState(state)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}
