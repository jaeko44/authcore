package apiutil

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatUserAgent(t *testing.T) {
	assert.Equal(t, "Chrome 71.0.3578.98 (Mac OS X 10.14.2)", FormatUserAgent("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_14_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/71.0.3578.98 Safari/537.36"))
	assert.Equal(t, "Internet Explorer 9.0 (Windows Phone OS 7.5)", FormatUserAgent("Mozilla/5.0 (compatible; MSIE 9.0; Windows Phone OS 7.5; Trident/5.0; IEMobile/9.0)"))
	assert.Equal(t, "Safari 10.0 (iPhone OS 10.3.1)", FormatUserAgent("Mozilla/5.0 (iPhone; CPU iPhone OS 10_3_1 like Mac OS X) AppleWebKit/603.1.30 (KHTML, like Gecko) Version/10.0 Mobile/14E304 Safari/602.1"))
	assert.Equal(t, "Googlebot/2.1 (+http://www.google.com/bot.html)", FormatUserAgent("Googlebot/2.1 (+http://www.google.com/bot.html)"))
}
