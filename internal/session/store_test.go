package session

import (
	"context"
	"os"
	"testing"

	"authcore.io/authcore/internal/config"
	dbx "authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/internal/user"
	"authcore.io/authcore/pkg/paging"

	"github.com/mssola/user_agent"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/peer"
)

func TestMain(m *testing.M) {
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func storeForTest() (*Store, func()) {
	config.InitDefaults()
	viper.Set("access_token_private_key", accessTokenPrivateKeyForTest)
	viper.Set("service_account_public_key", serviceAccountPublicKeyForTest)
	viper.Set("service_account_id", "123456")
	viper.Set("applications.test-client.name", "test")
	config.InitConfig()

	testutil.FixturesSetUp()
	db := dbx.NewDBFromConfig()
	redis := testutil.RedisForTest()
	encryptor := testutil.EncryptorForTest()
	userStore := user.NewStore(db, redis, encryptor)
	store := NewStore(db, redis, userStore)

	return store, func() {
		db.Close()
		viper.Reset()
		redis.FlushAll()
	}
}

func TestCreateSession(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	mDWithIPv4 := metadata.New(map[string]string{
		"ip-address": "127.0.0.1",
	})
	mDWithIPv6 := metadata.New(map[string]string{
		"ip-address": "2001:0db8:85a3:0000:0000:8a2e:0370:7334",
	})

	contextWithMDIPv4 := metadata.NewIncomingContext(context.Background(), mDWithIPv4)
	contextWithMDIPv6 := metadata.NewIncomingContext(context.Background(), mDWithIPv6)

	// Test for IP field with normal IP v4 value
	session, err := store.CreateSession(contextWithMDIPv4, 1, 1, "test-client", "REFRESH", false)
	if assert.NoError(t, err) {
		assert.Equal(t, "test-client", session.ClientID.String)
		assert.Equal(t, int64(1), session.UserID)
		assert.Equal(t, int64(1), session.DeviceID.Int64)
		assert.Equal(t, false, session.IsMachine)
		assert.NotNil(t, session.LastSeenAt)
		assert.Equal(t, "127.0.0.1", session.LastSeenIP)
		assert.Equal(t, "null", session.LastSeenLocation)
		assert.Equal(t, "REFRESH", session.RefreshToken)
	}

	// Test for IP field with normal IP v6 value
	session, err = store.CreateSession(contextWithMDIPv6, 1, 1, "test-client", "REFRESH1", false)
	if assert.NoError(t, err) {
		assert.Equal(t, "test-client", session.ClientID.String)
		assert.Equal(t, int64(1), session.UserID)
		assert.Equal(t, int64(1), session.DeviceID.Int64)
		assert.Equal(t, false, session.IsMachine)
		assert.NotNil(t, session.LastSeenAt)
		assert.Equal(t, "2001:0db8:85a3:0000:0000:8a2e:0370:7334", session.LastSeenIP)
		assert.Equal(t, "null", session.LastSeenLocation)
		assert.Equal(t, "REFRESH1", session.RefreshToken)
	}

	// Test for IP field with "null" value
	session, err = store.CreateSession(context.Background(), 1, 1, "test-client", "REFRESH2", false)
	if assert.NoError(t, err) {
		assert.Equal(t, "test-client", session.ClientID.String)
		assert.Equal(t, int64(1), session.UserID)
		assert.Equal(t, int64(1), session.DeviceID.Int64)
		assert.Equal(t, false, session.IsMachine)
		assert.NotNil(t, session.LastSeenAt)
		assert.Equal(t, "", session.LastSeenIP)
		assert.Equal(t, "null", session.LastSeenLocation)
		assert.Equal(t, "REFRESH2", session.RefreshToken)
	}
}

func TestCreateSessionWithUserAgent(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	contextWithPeerIPv4 := peer.NewContext(context.Background(), &peer.Peer{
		Addr:     &TestAddrIPv4{},
		AuthInfo: nil,
	})

	userAgent := user_agent.New("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	ip := "127.0.0.1"

	ctx := context.WithValue(contextWithPeerIPv4, UserAgentKey{}, userAgent)
	ctx = context.WithValue(ctx, IPKey{}, ip)

	session, err := store.CreateSession(ctx, 1, 1, "test-client", "REFRESH", false)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1), session.UserID)
		assert.Equal(t, int64(1), session.DeviceID.Int64)
		assert.Equal(t, false, session.IsMachine)
		assert.NotNil(t, session.LastSeenAt)
		assert.Equal(t, "127.0.0.1", session.LastSeenIP)
		assert.Equal(t, "null", session.LastSeenLocation)
		assert.Equal(t, "REFRESH", session.RefreshToken)
		assert.Equal(t, "Mac OS X 10.14.2 Chrome 71.0.3578.98", session.UserAgent)
	}
}

func TestCreateMachineSession(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	mDWithIPv4 := metadata.New(map[string]string{
		"ip-address": "127.0.0.1",
	})

	contextWithMDIPv4 := metadata.NewIncomingContext(context.Background(), mDWithIPv4)

	// Test for IP field with normal IP v4 value
	session, err := store.CreateMachineSession(contextWithMDIPv4, 1)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1), session.UserID)
		assert.Equal(t, false, session.DeviceID.Valid)
		assert.Equal(t, true, session.IsMachine)
		assert.NotNil(t, session.LastSeenAt)
		assert.Equal(t, "127.0.0.1", session.LastSeenIP)
		assert.Equal(t, "null", session.LastSeenLocation)
		assert.NotEmpty(t, session.RefreshToken)
	}
}

func TestUpdate(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	session, err := store.FindSessionByInternalID(context.TODO(), 2)
	session.LastSeenLocation = "Taiwan"
	session.LastSeenIP = "null"
	newSession, err := store.UpdateSession(context.TODO(), session)
	if assert.NoError(t, err) {
		assert.Equal(t, "Taiwan", newSession.LastSeenLocation)
		assert.Equal(t, "null", newSession.LastSeenIP)
	}
}

func TestFindSessionByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	session, err := store.FindSessionByInternalID(context.TODO(), 1)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1), session.UserID)
		assert.NotNil(t, session.LastSeenAt)
		assert.Equal(t, "1.1.1.1", session.LastSeenIP)
		assert.Equal(t, "HK", session.LastSeenLocation)
		assert.Equal(t, computeRefreshTokenHash("BOBREFRESHTOKEN1"), session.RefreshTokenHash)
	}
}

func TestFindSessionByRefreshToken(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	session, err := store.FindSessionByRefreshToken(context.TODO(), "BOBREFRESHTOKEN1")
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1), session.UserID)
		assert.NotNil(t, session.LastSeenAt)
		assert.Equal(t, "1.1.1.1", session.LastSeenIP)
		assert.Equal(t, "HK", session.LastSeenLocation)
		assert.Equal(t, computeRefreshTokenHash("BOBREFRESHTOKEN1"), session.RefreshTokenHash)
	}
}

func TestFindAllSessionsByUser(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	pageOptions := paging.PageOptions{
		SortColumn:    "created_at",
		UniqueColumn:  "id",
		SortDirection: paging.Desc,
		Limit:         50,
	}

	sessionsWithUser, _, err := store.FindAllSessionsByUser(context.TODO(), pageOptions, "1")

	sessionsArray := *sessionsWithUser

	if assert.NoError(t, err) {
		assert.Len(t, sessionsArray, 3)
	}

	sessions, _, err := store.FindAllSessionsByUser(context.TODO(), pageOptions, "")

	sessionsArray = *sessions

	if assert.NoError(t, err) {
		assert.Len(t, sessionsArray, 5)
	}
}

func TestInvalidateSessionByID(t *testing.T) {
	store, teardown := storeForTest()
	defer teardown()

	_, err := store.InvalidateSessionByID(context.TODO(), 1)
	assert.Nil(t, err)

	_, err = store.FindSessionByInternalID(context.TODO(), 1)
	assert.Error(t, err)
}

// TestAddrIPv4 serves a mock of Addr for test case usage.
type TestAddrIPv4 struct{}

// Network name for TestAddrIPv4
func (addr *TestAddrIPv4) Network() string {
	return "test-network"
}

func (addr *TestAddrIPv4) String() string {
	return "127.0.0.1:45138"
}

// TestAddrIPv6 serves a mock of Addr for test case usage.
type TestAddrIPv6 struct{}

// Network name for TestAddrIPv6
func (addr *TestAddrIPv6) Network() string {
	return "test-network"
}

func (addr *TestAddrIPv6) String() string {
	return "[2001:0db8:85a3:0000:0000:8a2e:0370:7334]:13245"
}
