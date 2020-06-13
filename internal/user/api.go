package user

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"github.com/jinzhu/copier"
	"github.com/labstack/echo/v4"

	"authcore.io/authcore/internal/apiutil"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"
)

// APIv2 returns a function that registers API 2.0 endpoints with an Echo instance.
func APIv2(store *Store) func(e *echo.Echo) {
	return func(e *echo.Echo) {
		h := &handler{store: store}

		g := e.Group("/api/v2")
		g.GET("/users", h.ListUsers)
		g.GET("/users/:id", h.GetUser)
		g.DELETE("/users/:id", h.DeleteUser)
		g.PUT("/users/:id", h.UpdateUser)
		g.POST("/users/:id/password", h.UpdateUserPassword)
		g.GET("/users/:id/roles", h.GetUserRoles)
		g.POST("/users/:id/roles", h.AssignUserRole)
		g.DELETE("/users/:id/roles/:role_id", h.UnassignUserRole)
		g.GET("/users/:id/idp", h.ListUserIDP)
		g.DELETE("/users/:id/idp/:service", h.DeleteUserIDP)
		g.GET("/users/:id/mfa", h.ListUserMFA)

		g.GET("/users/current", h.GetCurrentUser)
		g.GET("/users/current/idp", h.ListCurrentUserIDP)
		g.GET("/users/current/mfa", h.ListCurrentUserMFA)
		g.POST("/users/current/mfa", h.CreateCurrentUserMFA)
		g.DELETE("/users/current/mfa/:id", h.DeleteCurrentUserMFA)
		g.DELETE("/users/current/idp/:service", h.DeleteCurrentUserIDP)
		g.PUT("/users/current/password", h.UpdateCurrentUserPassword)

		g.GET("/idp/:id", h.GetIDP)
		g.DELETE("/mfa/:id", h.DeleteMFA)
	}
}

type handler struct {
	store *Store
}

func (h *handler) ListUsers(c echo.Context) error {
	r := UsersQuery{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	if err := c.Validate(&r); err != nil {
		return err
	}

	ctx := c.Request().Context()
	users, page, err := h.store.AllUsersWithQuery(ctx, r)
	if err != nil {
		return err
	}
	jsonUsers := make([]JSONUser, len(*users))
	for i, u := range *users {
		jsonUsers[i], err = NewJSONUser(&u)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(jsonUsers, page)
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) GetUser(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()
	user, err := h.store.UserByID(ctx, id)
	if err != nil {
		return err
	}

	jsonUser, err := NewJSONUser(user)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, jsonUser)
}

func (h *handler) DeleteUser(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()

	// Check if user delete itself
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	if me.ID == id {
		return errors.New(errors.ErrorInvalidArgument, "cannot delete current user")
	}

	err = h.store.DeleteUserByID(ctx, id)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) UpdateUser(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	req := make(map[string]interface{})
	err = json.NewDecoder(c.Request().Body).Decode(&req)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	ctx := c.Request().Context()
	user := &User{
		ID: id,
	}
	if err = h.store.SelectUser(ctx, user); err != nil {
		return err
	}
	loadUpdateUserRequest(req, user)

	if err = h.store.UpdateUser(ctx, user); err != nil {
		return err
	}

	if err = h.store.SelectUser(ctx, user); err != nil {
		return err
	}

	jsonUser, err := NewJSONUser(user)
	if err != nil {
		return err
	}

	return c.JSON(http.StatusOK, jsonUser)
}

func (h *handler) UpdateUserPassword(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	r := PasswordVerifier{}
	if err := c.Bind(&r); err != nil {
		return err
	}

	ctx := c.Request().Context()
	user := &User{
		ID: id,
	}
	if err = h.store.SelectUser(ctx, user); err != nil {
		return err
	}

	err = user.SetPasswordVerifier(r.Salt, r.VerifierW0, r.VerifierL)
	if err != nil {
		return err
	}

	err = h.store.UpdateUser(ctx, user)
	if err != nil {
		return err
	}

	return c.NoContent(http.StatusNoContent)
}

func (h *handler) GetUserRoles(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()
	roles, err := h.store.FindAllRolesByUserID(ctx, id)
	if err != nil {
		return err
	}
	jsonRoles := make([]JSONRole, len(*roles))
	for i, r := range *roles {
		jsonRoles[i], err = NewJSONRole(&r)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(jsonRoles, nil)
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) AssignUserRole(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	r := RolesRequest{}
	if err := c.Bind(&r); err != nil {
		return err
	}

	roleUser := RoleUser{
		RoleID: r.RoleID,
		UserID: id,
	}

	ctx := c.Request().Context()
	err = h.store.AssignRole(ctx, &roleUser)
	if err != nil {
		return err
	}
	return h.GetUserRoles(c)
}

func (h *handler) ListUserIDP(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()
	oauthFactors, err := h.store.FindAllOAuthFactorsByUserID(ctx, id)
	if err != nil {
		return err
	}
	resp := apiutil.NewListPagination(oauthFactors, nil)
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) ListCurrentUserIDP(c echo.Context) error {
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	ctx := c.Request().Context()
	oauthFactors, err := h.store.FindAllOAuthFactorsByUserID(ctx, me.ID)
	if err != nil {
		return err
	}
	results := make([]JSONIDP, len(*oauthFactors))
	for i, f := range *oauthFactors {
		results[i], err = NewJSONIDP(&f)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(results, &paging.Page{})
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) ListUserMFA(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()
	secondFactors, err := h.store.FindAllSecondFactorsByUserID(ctx, id)
	jsonSecondFactors := make([]JSONSecondFactor, len(*secondFactors))
	for i, sf := range *secondFactors {
		jsonSecondFactors[i], err = NewJSONSecondFactor(&sf)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(jsonSecondFactors, &paging.Page{})
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) DeleteMFA(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()
	_, err = h.store.FindSecondFactorByID(ctx, id)
	if err != nil {
		return err
	}

	err = h.store.DeleteSecondFactorByID(ctx, id)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) ListCurrentUserMFA(c echo.Context) error {
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	ctx := c.Request().Context()
	secondFactors, err := h.store.FindAllSecondFactorsByUserID(ctx, me.ID)
	jsonSecondFactors := make([]JSONSecondFactor, len(*secondFactors))
	for i, sf := range *secondFactors {
		jsonSecondFactors[i], err = NewJSONSecondFactor(&sf)
		if err != nil {
			return err
		}
	}
	resp := apiutil.NewListPagination(jsonSecondFactors, &paging.Page{})
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) CreateCurrentUserMFA(c echo.Context) error {
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	r := CreateMFARequest{}
	if err := c.Bind(&r); err != nil {
		return err
	}
	if err := c.Validate(&r); err != nil {
		return err
	}
	secondFactorType, err := SecondFactorTypeFromString(r.Type)
	if err != nil {
		return errors.Errorf(errors.ErrorInvalidArgument, "unknown type %v", r.Type)
	}

	// Only TOTP is supported in v2 API
	if secondFactorType != SecondFactorTOTP {
		return errors.Errorf(errors.ErrorInvalidArgument, "unknown type %v", r.Type)
	}

	secondFactor := &SecondFactor{
		UserID: me.ID,
		Type:   secondFactorType,
		Content: SecondFactorContent{
			Secret: nulls.NewString(r.Secret),
		},
	}
	v, err := secondFactor.ToVerifier(h.store.verifierFactory)
	if err != nil {
		return err
	}

	ok, _ = v.Verify([]byte{}, r.Verifier)
	if !ok {
		return errors.New(errors.ErrorInvalidArgument, "verifier is invalid")
	}

	ctx := c.Request().Context()
	secondFactors, err := h.store.FindAllSecondFactorsByUserIDAndType(ctx, me.ID, secondFactorType)
	if err != nil {
		return err
	}
	if len(*secondFactors) > 0 {
		return errors.Errorf(errors.ErrorAlreadyExists, "%v factor already exists", r.Type)
	}

	secondFactor, err = h.store.CreateSecondFactor(ctx, secondFactor)
	if err != nil {
		return err
	}

	resp, err := NewJSONSecondFactor(secondFactor)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) DeleteCurrentUserMFA(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	ctx := c.Request().Context()
	secondFactor, err := h.store.FindSecondFactorByID(ctx, id)
	if err != nil {
		return err
	}
	if secondFactor.UserID != me.ID {
		return errors.New(errors.ErrorNotFound, "")
	}
	err = h.store.DeleteSecondFactorByID(ctx, id)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

// DeleteUserIDP deletes all IDP by given user id and service
func (h *handler) DeleteUserIDP(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	service, err := ToOAuthService(c.Param("service"))
	if err != nil {
		return err
	}
	ctx := c.Request().Context()
	err = h.store.DeleteOAuthFactorByUserIDAndService(ctx, id, service)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) DeleteCurrentUserIDP(c echo.Context) error {
	service, err := ToOAuthService(c.Param("service"))
	if err != nil {
		return err
	}
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	ctx := c.Request().Context()
	err = h.store.DeleteOAuthFactorByUserIDAndService(ctx, me.ID, service)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) UnassignUserRole(c echo.Context) error {
	s := c.Param("id")
	uid, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	srid := c.Param("role_id")
	rid, err := strconv.ParseInt(srid, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}

	ctx := c.Request().Context()
	err = h.store.UnassignByRoleIDAndUserID(ctx, rid, uid)
	if err != nil {
		return err
	}
	return c.NoContent(http.StatusNoContent)
}

func (h *handler) GetCurrentUser(c echo.Context) error {
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}

	ctx := c.Request().Context()
	roles, err := h.store.FindAllRolesByUserID(ctx, me.ID)
	if err != nil {
		return err
	}
	jsonRoles := make([]JSONRole, len(*roles))
	for i, r := range *roles {
		jsonRoles[i], err = NewJSONRole(&r)
		if err != nil {
			return err
		}
	}
	resp, err := NewCurrentUser(me, jsonRoles)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) UpdateCurrentUserPassword(c echo.Context) error {
	me, ok := FromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}
	sess, ok := sessionFromContext(c)
	if !ok {
		return errors.New(errors.ErrorUnauthenticated, "")
	}

	if me.IsPasswordAuthenticationEnabled() && !sess.UpdateCurrentUserPasswordAllowed() {
		return errors.New(errors.ErrorPermissionDenied, "step-up authentication is required")
	}

	r := new(PasswordVerifier)
	if err := c.Bind(&r); err != nil {
		return err
	}

	ctx := c.Request().Context()

	if err := h.store.SelectUser(ctx, me); err != nil {
		return err
	}

	if err := me.SetPasswordVerifier(r.Salt, r.VerifierW0, r.VerifierL); err != nil {
		return err
	}

	if err := h.store.UpdateUser(ctx, me); err != nil {
		return err
	}

	roles, err := h.store.FindAllRolesByUserID(ctx, me.ID)
	if err != nil {
		return err
	}
	jsonRoles := make([]JSONRole, len(*roles))
	for i, r := range *roles {
		jsonRoles[i], err = NewJSONRole(&r)
		if err != nil {
			return err
		}
	}
	resp, err := NewCurrentUser(me, jsonRoles)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, resp)
}

func (h *handler) GetIDP(c echo.Context) error {
	s := c.Param("id")
	id, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return errors.Wrap(err, errors.ErrorInvalidArgument, "")
	}
	ctx := c.Request().Context()

	idp, err := h.store.FindOAuthFactorByID(ctx, id)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, idp)
}

// JSONUser represents a user in management API.
type JSONUser struct {
	ID                              int64        `json:"id"`
	Name                            nulls.String `json:"name"`
	Username                        nulls.String `json:"preferred_username"`
	Email                           nulls.String `json:"email"`
	EmailVerified                   bool         `json:"email_verified"`
	Phone                           nulls.String `json:"phone_number"`
	PhoneVerified                   bool         `json:"phone_number_verified"`
	IsPasswordAuthenticationEnabled bool         `json:"is_password_set"`
	IsLocked                        bool         `json:"locked"`
	UserMetadata                    nulls.JSON   `json:"user_metadata"`
	AppMetadata                     nulls.JSON   `json:"app_metadata"`
	Language                        string       `json:"language"`
	UpdatedAt                       time.Time    `json:"updated_at"`
	CreatedAt                       time.Time    `json:"created_at"`
	LastSeenAt                      time.Time    `json:"last_seen_at"`
}

// NewJSONUser converts a User into JSONUser.
func NewJSONUser(user *User) (JSONUser, error) {
	j := JSONUser{}
	err := copier.Copy(&j, user)
	j.Language = user.RealLanguage()
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}

// CurrentUser represents the current user.
type CurrentUser struct {
	ID                              int64        `json:"id"`
	Name                            nulls.String `json:"name"`
	Username                        nulls.String `json:"preferred_username"`
	Email                           nulls.String `json:"email"`
	EmailVerified                   bool         `json:"email_verified"`
	Phone                           nulls.String `json:"phone_number"`
	PhoneVerified                   bool         `json:"phone_number_verified"`
	IsPasswordAuthenticationEnabled bool         `json:"is_password_set"`
	UserMetadata                    nulls.JSON   `json:"user_metadata"`
	Language                        string       `json:"language"`
	Roles                           []JSONRole   `json:"roles"`
	UpdatedAt                       time.Time    `json:"updated_at"`
	CreatedAt                       time.Time    `json:"created_at"`
	LastSeenAt                      time.Time    `json:"last_seen_at"`
}

// NewCurrentUser converts a User into JSONUser.
func NewCurrentUser(user *User, roles []JSONRole) (CurrentUser, error) {
	j := CurrentUser{}
	err := copier.Copy(&j, user)
	j.Language = user.RealLanguage()
	j.Roles = roles
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}

// loadUpdateUserRequest load up a map from json response to user
func loadUpdateUserRequest(m map[string]interface{}, user *User) {
	if name, ok := m["name"].(string); ok {
		user.Name = nulls.NewString(name)
	}
	if username, ok := m["preferred_username"].(string); ok {
		user.Username = nulls.NewString(username)
	}
	if email, ok := m["email"].(string); ok {
		user.Email = nulls.NewString(email)
	}
	if phone, ok := m["phone_number"].(string); ok {
		user.Phone = nulls.NewString(phone)
	}
	if emailVerified, ok := m["email_verified"].(bool); ok {
		if emailVerified != user.EmailVerified() {
			if emailVerified {
				user.EmailVerifiedAt = nulls.NewTime(time.Now())
			} else {
				user.EmailVerifiedAt = nulls.Time{}
			}
		}
	}
	if phoneVerified, ok := m["phone_number_verified"].(bool); ok {
		if phoneVerified != user.PhoneVerified() {
			if phoneVerified {
				user.PhoneVerifiedAt = nulls.NewTime(time.Now())
			} else {
				user.PhoneVerifiedAt = nulls.Time{}
			}
		}
	}
	if appMetadata, ok := m["app_metadata"]; ok {
		user.AppMetadata = nulls.NewJSON(appMetadata)
	}
	if userMetadata, ok := m["user_metadata"]; ok {
		user.UserMetadata = nulls.NewJSON(userMetadata)
	}
	if isLocked, ok := m["is_locked"].(bool); ok {
		user.IsLocked = isLocked
		if isLocked {
			t, _ := time.Parse(time.RFC3339, "2038-01-19T00:00:00Z")
			user.LockExpiredAt = nulls.NewTime(t)
		} else {
			user.LockExpiredAt = nulls.Time{}
		}
	}
}

// PasswordVerifier represents a password verifier.
type PasswordVerifier struct {
	Method     string `json:"method"`
	Salt       []byte `json:"salt"`
	VerifierW0 []byte `json:"w0"`
	VerifierL  []byte `json:"l"`
}

// RolesRequest represents a user roles request in management API.
type RolesRequest struct {
	RoleID int64 `json:"role_id"`
}

// CreateMFARequest is a request for CreateCurrentUserMFA.
type CreateMFARequest struct {
	Type     string `json:"type" validate:"required"`
	Secret   string `json:"secret" validate:"required"`
	Verifier []byte `json:"verifier" validate:"required"`
}

// JSONRole represents a role record in management API.
type JSONRole struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

// NewJSONRole converts a Role to JSONRole.
func NewJSONRole(r *Role) (JSONRole, error) {
	j := JSONRole{}
	err := copier.Copy(&j, r)
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}

// JSONSecondFactor represents a second factor record in management API.
type JSONSecondFactor struct {
	ID         int64     `json:"id"`
	Type       string    `json:"type"`
	Value      string    `json:"value"`
	UpdatedAt  time.Time `json:"updated_at"`
	CreatedAt  time.Time `json:"created_at"`
	LastUsedAt time.Time `json:"last_used_at"`
}

// NewJSONSecondFactor converts a SecondFactor to JSONSecondFactor.
func NewJSONSecondFactor(sf *SecondFactor) (JSONSecondFactor, error) {
	j := JSONSecondFactor{}
	err := copier.Copy(&j, sf)
	j.Type = sf.Type.String()
	if j.Type == "sms_otp" {
		j.Value = sf.Content.PhoneNumber.String
	}
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}

// JSONIDP represents a IDP binding.
type JSONIDP struct {
	ServiceName string    `json:"service"`
	OAuthUserID string    `json:"oauth_user_id"`
	UpdatedAt   time.Time `json:"updated_at"`
	CreatedAt   time.Time `json:"created_at"`
	LastUsedAt  time.Time `json:"last_used_at"`
}

// NewJSONIDP converts *OAuthFactor to JSONIDP.
func NewJSONIDP(f *OAuthFactor) (JSONIDP, error) {
	j := JSONIDP{}
	err := copier.Copy(&j, f)
	return j, errors.Wrap(err, errors.ErrorUnknown, "")
}
