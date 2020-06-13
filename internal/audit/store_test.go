package audit

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/testutil"
	"authcore.io/authcore/pkg/nulls"

	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func storeForTest() (*Store, func()) {
	testutil.FixturesSetUp()
	config.InitDefaults()
	config.InitConfig()
	d := db.NewDBFromConfig()
	a := NewStore(d)

	return a, func() {
		d.Close()
		config.Reset()
	}
}

func TestInsert(t *testing.T) {
	a, teardown := storeForTest()
	defer teardown()

	event := &Event{
		ActorID:      nulls.NewInt64(1),
		ActorDisplay: nulls.NewString("bob"),
		Action:       "test",
		Target:       nulls.JSON{},
		Result:       EventResultSuccess,
		IP:           nulls.NewString("127.0.0.1"),
	}

	err := a.InsertEvent(context.Background(), event)
	assert.NoError(t, err)
	assert.NotEmpty(t, event.ID)
}

func TestInsertValidation(t *testing.T) {
	a, teardown := storeForTest()
	defer teardown()

	event := &Event{
		ActorID:      nulls.NewInt64(1),
		ActorDisplay: nulls.NewString("bob"),
		Action:       "test",
		Target:       nulls.JSON{},
		Result:       EventResultSuccess,
		IP:           nulls.NewString("invalid"),
	}

	err := a.InsertEvent(context.Background(), event)
	assert.Error(t, err)
}

func TestAllEventsWithQuery(t *testing.T) {
	a, teardown := storeForTest()
	defer teardown()

	eventsQuery := EventsQuery{
		ActorID: "1",
	}

	auditLogs, _, err := a.AllEventsWithQuery(context.TODO(), eventsQuery)

	if assert.NoError(t, err) {
		assert.Len(t, *auditLogs, 3)
	}
}

func TestLogEvent(t *testing.T) {
	a, teardown := storeForTest()
	defer teardown()

	e := echo.New()
	req := httptest.NewRequest(http.MethodPost, "/", nil)
	req.Header.Set("User-Agent", "Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.Set("user", &mockUser{ID: 3, Name: "test"})

	target := map[string]interface{}{"test": "1"}

	a.LogEvent(c, nil, "test", target, EventResultSuccess)

	eventsQuery := EventsQuery{
		ActorID: "3",
	}
	auditLogs, _, err := a.AllEventsWithQuery(context.TODO(), eventsQuery)

	if assert.NoError(t, err) {
		assert.Len(t, *auditLogs, 1)
		event := (*auditLogs)[0]
		assert.Equal(t, int64(3), event.ActorID.Int64)
		assert.Equal(t, "test", event.ActorDisplay.String)
		assert.Equal(t, "test", event.Action)
		assert.Equal(t, EventResultSuccess, event.Result)
		assert.Equal(t, "192.0.2.1", event.IP.String)
		assert.Equal(t, "Chrome 71.0.3578.98 (Mac OS X 10.14.2)", event.UserAgent.String)
	}
}

type mockUser struct {
	ID   int64
	Name string
}

func (u *mockUser) ActorID() int64 {
	return u.ID
}

func (u *mockUser) DisplayName() string {
	return u.Name
}
