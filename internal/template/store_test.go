package template

import (
	"context"
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"golang.org/x/text/language"

	"authcore.io/authcore/internal/config"
	"authcore.io/authcore/internal/db"
	"authcore.io/authcore/internal/testutil"
)

func TestMain(m *testing.M) {
	testutil.DBSetUp()
	code := m.Run()
	testutil.DBTearDown()
	os.Exit(code)
}

func storeForTest() (*Store, func()) {
	config.InitDefaults()
	viper.Set("base_path", "../..")
	config.InitConfig()

	testutil.FixturesSetUp()
	d := db.NewDBFromConfig()
	store := NewStore(d)

	return store, func() {
		d.Close()
		viper.Reset()
	}
}

func TestFindTemplateByTypeAndLocale(t *testing.T) {
	s, teardown := storeForTest()
	defer teardown()
	name := "VerificationSMS.txt"
	template, err := s.FindTemplateByTypeAndLocale(context.TODO(), name, language.English)
	if assert.NoError(t, err) {
		assert.Equal(t, int64(1), template.ID)
		assert.Equal(t, name, template.Name)
		assert.Equal(t, "en", template.Language)
		assert.Equal(t, "Your verification code is {code}", template.Template)
	}

	template, err = s.FindTemplateByTypeAndLocale(context.TODO(), "testing", language.English)
	assert.Error(t, err)
}

func TestUpdateOrCreateTemplate(t *testing.T) {
	s, teardown := storeForTest()
	defer teardown()

	// Updates an existing template
	err := s.UpdateOrCreateTemplate(
		context.TODO(),
		"VerificationSMS.txt",
		language.English,
		"Hi, your verification code is {code}!",
	)
	assert.NoError(t, err)

	template, err := s.FindTemplateByTypeAndLocale(context.TODO(), "VerificationSMS.txt", language.English)
	assert.NoError(t, err)
	assert.Equal(t, "Hi, your verification code is {code}!", template.Template)

	// Creates a new template
	err = s.UpdateOrCreateTemplate(
		context.TODO(),
		"VerificationSMS.txt",
		language.Chinese,
		"你好，你的認證碼是 {code}！",
	)
	assert.NoError(t, err)

	template, err = s.FindTemplateByTypeAndLocale(context.TODO(), "VerificationSMS.txt", language.Chinese)
	assert.NoError(t, err)
	assert.Equal(t, "你好，你的認證碼是 {code}！", template.Template)
}

func TestDeleteTemplateByTypeAndName(t *testing.T) {
	s, teardown := storeForTest()
	defer teardown()

	_, err := s.FindTemplateByTypeAndLocale(context.TODO(), "VerificationSMS.txt", language.English)
	assert.NoError(t, err)

	// Deletes an existing template
	err = s.DeleteTemplateByNameAndLanguage(
		context.TODO(),
		"VerificationSMS.txt",
		language.English,
	)
	assert.NoError(t, err)

	_, err = s.FindTemplateByTypeAndLocale(context.TODO(), "VerificationSMS.txt", language.English)
	assert.Error(t, err)
}

func TestGetTemplate(t *testing.T) {
	s, teardown := storeForTest()
	defer teardown()

	// Test from DB
	template, err := s.GetTemplate(context.TODO(), "VerificationSMS.txt", language.English)
	assert.NoError(t, err)
	assert.Equal(t, "Your verification code is {code}", template.Template)

	// Test execute as well
	message := template.Execute(map[string]string{"code": "123456"})
	assert.Equal(t, "Your verification code is 123456", message)

	// Test File default fallback
	template, err = s.GetTemplate(context.TODO(), "AuthenticationSMS.txt", language.English)
	assert.NoError(t, err)
	assert.Equal(t, "Your authentication code for {application_name} is {code}.", template.Template)

	message = template.Execute(map[string]string{"code": "123456", "application_name": "Authcore"})
	assert.Equal(t, "Your authentication code for Authcore is 123456.", message)
}
