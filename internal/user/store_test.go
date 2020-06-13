package user

import (
	"context"
	"os"
	"testing"
	"time"

	"authcore.io/authcore/internal/config"
	dbx "authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/errors"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/validator"
	"authcore.io/authcore/pkg/nulls"
	"authcore.io/authcore/pkg/paging"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
)

func TestMain(m *testing.M) {
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func storeForTest() (*Store, func()) {
	config.InitDefaults()
	viper.Set("require_user_email_or_phone", false)
	viper.Set("require_user_phone", false)
	viper.Set("require_user_email", false)
	viper.Set("require_user_username", false)
	viper.Set("secret_key_base", "855edf399835e9c9deb61877c1a76bf14eed7c35a167e10ff1b7d43db4363268")
	viper.Set("base_path", "../..")
	config.InitConfig()

	testutil.FixturesSetUp()
	db := dbx.NewDBFromConfig()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	store := NewStore(db, redis, encryptor)

	return store, func() {
		db.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

func TestUserValidate(t *testing.T) {
	u := &User{}

	err := validator.Validate.Struct(u)
	if assert.NotNil(t, err) {
		err = errors.WithValidateError(err)
		ie := err.(*errors.Error)
		assert.Len(t, ie.FieldViolations(), 1)
		assert.Equal(t, ie.FieldViolations()[0].Field, "email")
	}
}

func TestCreateUser(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	u := &User{
		Username:       nulls.NewString("alice"),
		Email:          nulls.NewString("alice@example.com"),
		Phone:          nulls.NewString("+85291234567"),
		DisplayNameOld: "Alice",
		Language:       nulls.NewString("en"),
	}

	mDWithUserAgent := metadata.New(map[string]string{"grpcgateway-user-agent": "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"})

	contextWithMD := metadata.NewIncomingContext(context.Background(), mDWithUserAgent)

	err := store.InsertUser(contextWithMD, u)
	if assert.Nil(t, err) {
		assert.NotNil(t, u)
		assert.NotZero(t, u.ID)
		assert.Equal(t, "alice", u.Username.String)
		assert.NotZero(t, u.CreatedAt)
		assert.NotZero(t, u.UpdatedAt)
	}
}

func TestUpdateUser(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	u, err := store.UserByPublicID(context.Background(), "2")
	if assert.NoError(t, err) {
		u.Username = nulls.NewString("carol_updated")
		err := store.UpdateUser(context.Background(), u)
		if assert.NoError(t, err) {
			u, err = store.UserByPublicID(context.Background(), "2")
			assert.NoError(t, err)
			assert.Equal(t, "carol_updated", u.Username.String)
		}
	}
}

func TestUpdateUserMetadata(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	u, err := store.UserByID(context.Background(), 1)
	assert.NoError(t, err)

	u.UserMetadata = nulls.NewJSON(`{
		"favourite_links": [
			"https://github.com",
			"https://google.com",
			"https://blocksq.com"
		]
	}`)
	err = store.UpdateUser(context.Background(), u)
	assert.NoError(t, err)

	err = store.SelectUser(context.Background(), u)
	assert.NoError(t, err)

	updatedUserMetadata, err := u.UserMetadata.String()
	assert.NoError(t, err)

	assert.Equal(t, `{"favourite_links":["https://github.com","https://google.com","https://blocksq.com"]}`, updatedUserMetadata)
}

func TestUpdateAppMetadata(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	u, err := store.UserByID(context.Background(), 1)
	assert.NoError(t, err)

	u.AppMetadata = nulls.NewJSON(`{
		"kyc_status": true
	}`)
	err = store.UpdateUser(context.Background(), u)
	assert.NoError(t, err)

	err = store.SelectUser(context.Background(), u)
	assert.NoError(t, err)

	updatedAppMetadata, err := u.AppMetadata.String()
	assert.NoError(t, err)

	assert.Equal(t, `{"kyc_status":true}`, updatedAppMetadata)
}

func TestUserByPublicID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	assert := assert.New(t)

	u, err := store.UserByPublicID(context.Background(), "2")
	if assert.NoError(err) {
		assert.NotNil(u)
		assert.Equal(int64(2), u.ID)
		assert.Equal("carol", u.Username.String)
		assert.Equal("carol@example.com", u.Email.String)
		assert.True(u.EmailVerifiedAt.Valid)
		assert.Equal("+85221111111", u.Phone.String)
		assert.False(u.PhoneVerifiedAt.Valid)
	}
}

func TestUserByPublicIDNotFound(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	_, err := store.UserByPublicID(context.Background(), "none")
	assert.Error(t, err)
}

func TestUserByHandle(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	u, err := store.UserByHandle(context.Background(), "benny@example.com")
	if assert.NoError(t, err) {
		assert.NotNil(t, u)
		assert.Equal(t, int64(6), u.ID)
	}
}

func TestAllUsersWithPageOptions(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	var options paging.PageOptions
	options = paging.PageOptions{
		SortColumn:   "updated_at",
		UniqueColumn: "id",
		Limit:        50,
	}
	users, _, err := store.AllUsersWithPageOptions(context.Background(), options)
	usersArray := *users

	if assert.NoError(t, err) {
		assert.Equal(t, "bob", usersArray[0].Username.String)
		assert.Equal(t, "carol", usersArray[1].Username.String)
	}
}

func TestAllUsersWithPageOptionsSortByLock(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	var options paging.PageOptions
	options = paging.PageOptions{
		SortColumn:    "is_locked",
		UniqueColumn:  "id",
		SortDirection: paging.Desc,
		Limit:         2,
	}
	users, page, err := store.AllUsersWithPageOptions(context.Background(), options)
	usersArray := *users

	if assert.NoError(t, err) {
		assert.Equal(t, "benny", usersArray[0].Username.String)
		assert.Equal(t, "last", usersArray[1].Username.String)
	}

	options.PageToken = page.NextPageToken
	users, _, err = store.AllUsersWithPageOptions(context.Background(), options)
	usersArray = *users

	if assert.NoError(t, err) {
		assert.Equal(t, "lastplusone", usersArray[0].Username.String)
		assert.Equal(t, "oliver", usersArray[1].Username.String)
	}
}

func TestAllUsersWithPageOptionsSortByCreatedAt(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	var options paging.PageOptions
	options = paging.PageOptions{
		SortColumn:   "created_at",
		UniqueColumn: "id",
		Limit:        2,
	}
	users, page, err := store.AllUsersWithPageOptions(context.Background(), options)
	usersArray := *users

	if assert.NoError(t, err) {
		assert.Equal(t, "bob", usersArray[0].Username.String)
		assert.Equal(t, "carol", usersArray[1].Username.String)
	}

	options.PageToken = page.NextPageToken
	users, _, err = store.AllUsersWithPageOptions(context.Background(), options)
	usersArray = *users

	if assert.NoError(t, err) {
		assert.Equal(t, "factor", usersArray[0].Username.String)
		assert.Equal(t, "twofactor", usersArray[1].Username.String)
	}
}

func TestAllUsersWithPageOptionsSortByLastSeenAt(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	var options paging.PageOptions
	options = paging.PageOptions{
		SortColumn:    "last_seen_at",
		UniqueColumn:  "id",
		SortDirection: paging.Desc,
		Limit:         2,
	}
	users, page, err := store.AllUsersWithPageOptions(context.Background(), options)
	usersArray := *users

	if assert.NoError(t, err) {
		assert.Equal(t, "bob", usersArray[0].Username.String)
		assert.Equal(t, "benny", usersArray[1].Username.String)
	}

	options.PageToken = page.NextPageToken
	users, _, err = store.AllUsersWithPageOptions(context.Background(), options)
	usersArray = *users

	if assert.NoError(t, err) {
		assert.Equal(t, "last", usersArray[0].Username.String)
		assert.Equal(t, "lastplusone", usersArray[1].Username.String)
	}
}

func TestAllUsersWithQuery(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	query := UsersQuery{
		Limit:    2,
		SortBy:   "id desc",
		Username: "plusone",
	}
	users, _, err := store.AllUsersWithQuery(context.Background(), query)
	usersArray := *users

	if assert.NoError(t, err) {
		assert.Equal(t, "lastplusone", usersArray[0].Username.String)
	}
}

func TestDeleteUserByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	err := store.DeleteUserByID(context.Background(), 1)
	assert.NoError(t, err)

	// not found
	err = store.DeleteUserByID(context.Background(), 1)
	assert.Error(t, err)
}

func TestCreateSecondFactor(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	secondFactor := &SecondFactor{
		UserID: int64(1),
		Type:   SecondFactorTOTP,
		Content: SecondFactorContent{
			Identifier: dbx.NullableString("jPhone"),
			Secret:     dbx.NullableString("THISISATOTPSECRETXXXXXXXXXXXXXXX"),
		},
	}
	secondFactor, err := store.CreateSecondFactor(context.TODO(), secondFactor)
	if assert.Nil(t, err) {
		assert.NotNil(t, secondFactor)
		assert.Equal(t, int64(1), secondFactor.UserID)
		assert.Equal(t, "THISISATOTPSECRETXXXXXXXXXXXXXXX", secondFactor.Content.Secret.String)
		assert.Equal(t, "jPhone", secondFactor.Content.Identifier.String)
	}
}

func TestFindSecondFactorByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	secondFactor, err := store.FindSecondFactorByID(context.TODO(), 2)
	if assert.Nil(t, err) {
		assert.NotNil(t, secondFactor)
		assert.Equal(t, SecondFactorTOTP, secondFactor.Type)
		assert.Equal(t, "THISISAWEAKTOTPSECRETFORTESTSXX2", secondFactor.Content.Secret.String)
		assert.Equal(t, int64(3), secondFactor.UserID)
	}
}

func TestFindAllSecondFactorsByUser(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	secondFactors, err := store.FindAllSecondFactorsByUserID(context.TODO(), 4)
	if assert.Nil(t, err) {
		assert.Len(t, *secondFactors, 3)
	}
}

func TestCreateOAuthFactor(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	metadata := nulls.NewJSON(map[string]interface{}{
		"name": "testing",
	})
	metadataString, err := metadata.String()
	assert.NoError(t, err)
	oauthFactor, err := store.CreateOAuthFactor(context.TODO(), 1, OAuthFacebook, "1337", metadata)
	if assert.NoError(t, err) {
		assert.NotNil(t, oauthFactor)
		assert.Equal(t, int64(1), oauthFactor.UserID)
		assert.Equal(t, "1337", oauthFactor.OAuthUserID)
		newMetadataString, err := oauthFactor.Metadata.String()
		assert.NoError(t, err)
		assert.Equal(t, metadataString, newMetadataString)
	}
}

func TestFindOAuthFactorByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	oauthFactor, err := store.FindOAuthFactorByID(context.TODO(), 1)
	if assert.Nil(t, err) {
		assert.NotNil(t, oauthFactor)
		assert.Equal(t, int64(11), oauthFactor.UserID)
		assert.Equal(t, OAuthFacebook, oauthFactor.Service)
		assert.Equal(t, "1337", oauthFactor.OAuthUserID)
	}
}

func TestFindAllOAuthFactorsByUser(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	oauthFactors, err := store.FindAllOAuthFactorsByUserID(context.TODO(), 11)
	if assert.Nil(t, err) {
		assert.Len(t, *oauthFactors, 3)
	}
}

func TestDeleteOAuthFactorByUserIDAndService(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	err := store.DeleteOAuthFactorByUserIDAndService(context.TODO(), 11, 999)
	assert.NoError(t, err)

	oauthFactors, err := store.FindAllOAuthFactorsByUserID(context.TODO(), 11)
	if assert.NoError(t, err) {
		assert.Len(t, *oauthFactors, 2)
	}

	err = store.DeleteOAuthFactorByUserIDAndService(context.TODO(), 1, 999)
	if assert.Error(t, err) {
		assert.True(t, errors.IsKind(err, errors.ErrorNotFound))
	}
}

func TestUpdateOAuthFactorMetadata(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	oauthFactor, err := store.FindOAuthFactorByID(context.TODO(), 1)
	metadata := nulls.NewJSON(map[string]interface{}{
		"name": "test",
	})
	metadataString, err := metadata.String()
	assert.NoError(t, err)
	updatedOauthFactor, err := store.UpdateOAuthFactorMetadata(context.TODO(), oauthFactor.ID, metadata)
	if assert.NoError(t, err) {
		assert.Equal(t, oauthFactor.ID, updatedOauthFactor.ID)
		newMetadataString, err := updatedOauthFactor.Metadata.String()
		assert.NoError(t, err)
		assert.Equal(t, metadataString, newMetadataString)
	}
}

func TestUpdateOAuthFactorLastUsedAt(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	oauthFactor, err := store.FindOAuthFactorByID(context.TODO(), 1)
	time, _ := time.Parse(time.RFC3339, "2020-01-01T00:00:00Z")
	updatedOauthFactor, err := store.UpdateOAuthFactorLastUsedAt(context.TODO(), oauthFactor.ID, time)
	if assert.Nil(t, err) {
		assert.Equal(t, "2020-01-01 00:00:00 +0000 UTC", updatedOauthFactor.LastUsedAt.String())
	}
}

func TestRoleValidate(t *testing.T) {
	role := &Role{
		ID:           1,
		Name:         "authcore.admin",
		IsSystemRole: true,
	}

	err := role.Validate()
	assert.Nil(t, err)
}

func TestCreateRole(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// Create the role with name being "snowdrop.editor"
	role := &Role{
		Name: "snowdrop.editor",
	}
	role, err := store.CreateRole(context.TODO(), role)
	if assert.Nil(t, err) {
		assert.NotNil(t, role)
		assert.False(t, role.IsSystemRole)
		assert.Equal(t, "snowdrop.editor", role.Name)
	}
}

func TestCreateRoleWithDuplicatedName(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// Create the role with name being "snowdrop.editor"
	role := &Role{
		Name: "snowdrop.editor",
	}
	role, err := store.CreateRole(context.TODO(), role)
	assert.Nil(t, err)

	// Create the role with name being "snowdrop.editor" again
	role, err = store.CreateRole(context.TODO(), role)
	assert.Error(t, err)
}

func TestFindRoleByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// Find the role with id = 1
	role, err := store.FindRoleByID(context.TODO(), int64(1))
	if assert.Nil(t, err) {
		assert.Equal(t, int64(1), role.ID)
		assert.Equal(t, "authcore.admin", role.Name)
		assert.True(t, role.IsSystemRole)
	}
}

func TestFindRoleByName(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	role, err := store.FindRoleByName(context.TODO(), "authcore.admin")
	if assert.Nil(t, err) {
		assert.Equal(t, int64(1), role.ID)
		assert.Equal(t, "authcore.admin", role.Name)
		assert.True(t, role.IsSystemRole)
	}
}

func TestFindAllRoles(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// Find all the roles
	roles, err := store.FindAllRoles(context.TODO())
	if assert.Nil(t, err) {
		assert.Len(t, *roles, 3)

		role1 := (*roles)[0]
		assert.Equal(t, int64(1), role1.ID)
		assert.Equal(t, "authcore.admin", role1.Name)

		role2 := (*roles)[1]
		assert.Equal(t, int64(2), role2.ID)
		assert.Equal(t, "authcore.editor", role2.Name)

		role3 := (*roles)[2]
		assert.Equal(t, int64(3), role3.ID)
		assert.Equal(t, "snowdrop.admin", role3.Name)
	}
}

func TestDeleteRoleByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// The role exists
	_, err := store.FindRoleByID(context.TODO(), int64(3))
	assert.Nil(t, err)

	// The role can be deleted (it is not a system role)
	err = store.DeleteRoleByID(context.TODO(), int64(3))
	assert.Nil(t, err)

	// The role no longer exists
	_, err = store.FindRoleByID(context.TODO(), int64(3))
	assert.Error(t, err)
}

func TestDeleteSystemRoleByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// The role exists
	_, err := store.FindRoleByID(context.TODO(), int64(1))
	assert.Nil(t, err)

	// The role can be deleted (it is not a system role)
	err = store.DeleteRoleByID(context.TODO(), int64(1))
	assert.Error(t, err)

	// The role still exists
	_, err = store.FindRoleByID(context.TODO(), int64(1))
	assert.Nil(t, err)
}

func TestRoleUserValidate(t *testing.T) {
	roleUser := &RoleUser{
		ID:     1,
		RoleID: 1,
		UserID: 1,
	}

	err := roleUser.Validate()
	assert.Nil(t, err)
}

func TestAssignRole(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// Assign the role to a user
	roleUser := &RoleUser{
		RoleID: 3,
		UserID: 6,
	}
	err := store.AssignRole(context.TODO(), roleUser)
	assert.Nil(t, err)
}

func TestAssignRoleDuplicate(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// Assign the role to a user
	roleUser := &RoleUser{
		RoleID: 3,
		UserID: 6,
	}
	err := store.AssignRole(context.TODO(), roleUser)
	assert.Nil(t, err)

	// Assign the role to the user again
	err = store.AssignRole(context.TODO(), roleUser)
	assert.Error(t, err)
}

func TestUnassignRoleByRoleIDAndUserID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	// The subject assignment exists
	_, err := store.FindSubjectAssignmentByRoleIDAndUserID(context.TODO(), int64(2), int64(2))
	assert.Nil(t, err)

	// Unassign the role from the user
	err = store.UnassignByRoleIDAndUserID(context.TODO(), int64(2), int64(2))
	assert.Nil(t, err)

	// The subject assignment no longer exists
	_, err = store.FindSubjectAssignmentByRoleIDAndUserID(context.TODO(), int64(2), int64(2))
	assert.Error(t, err)
}

func TestFindAllRolesByUserID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	roles, err := store.FindAllRolesByUserID(context.TODO(), int64(1))
	if assert.Nil(t, err) {
		assert.Len(t, *roles, 2)
		role1 := (*roles)[0]
		assert.Equal(t, int64(1), role1.ID)
		role2 := (*roles)[1]
		assert.Equal(t, int64(2), role2.ID)
	}
}
func TestAllUsersByRoleID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	users, err := store.AllUsersByRoleID(context.TODO(), int64(2))
	if assert.Nil(t, err) {
		assert.Len(t, *users, 2)
		user1 := (*users)[0]
		assert.Equal(t, int64(1), user1.ID)
		user2 := (*users)[1]
		assert.Equal(t, int64(2), user2.ID)
	}
}
