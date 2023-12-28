package chrome

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"time"

	cu "github.com/Davincible/chromedp-undetected"
	digioauth "github.com/holyhope/digiposte-oauth"
)

type Validatable interface {
	Validate() error
}

var chromeOpts = []cu.Option{ //nolint:gochecknoglobals
	func(c *cu.Config) {
		c.Language = "fr-FR"
	},
}

var errNegativeFreq = fmt.Errorf("frequency must be positive")

type WithRefreshFrequency struct {
	Frequency time.Duration
}

func (o *WithRefreshFrequency) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.refreshFrequency = o.Frequency

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

func (o *WithRefreshFrequency) Validate() error {
	if o.Frequency <= 0 {
		return &digioauth.InvalidOptionError{
			Name: "WithRefreshFrequency",
			Err:  errNegativeFreq,
		}
	}

	return nil
}

var errNegativeTimeout = fmt.Errorf("timeout must be positive")

type WithTimeout struct {
	Timeout time.Duration
}

func (o *WithTimeout) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.timeout = o.Timeout

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

func (o *WithTimeout) Validate() error {
	if o.Timeout <= 0 {
		return &digioauth.InvalidOptionError{
			Name: "WithTimeout",
			Err:  errNegativeTimeout,
		}
	}

	return nil
}

var errEmptyURL = errors.New("url is empty")

type WithURL struct {
	URL string
}

func (o *WithURL) Validate() error {
	if o.URL == "" {
		return &digioauth.InvalidOptionError{
			Name: "WithURL",
			Err:  errEmptyURL,
		}
	}

	return nil
}

func (o *WithURL) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.url = o.URL

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

type WithCookies struct {
	Cookies []*http.Cookie
}

func (o *WithCookies) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.cookies = o.Cookies

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

type WithScreenShortOnError struct{}

func (o *WithScreenShortOnError) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		chrome.screenShortOnError = true

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}

type InvalidTypeOptionError struct {
	instance interface{}
}

func (e *InvalidTypeOptionError) Error() string {
	return fmt.Sprintf("invalid instance type: %T", e.instance)
}

const (
	ConfigURL       = "url"
	ConfigCookieJar = "cookie"
)

type MissingOptionError struct {
	Option string
}

func (e *MissingOptionError) Error() string {
	return fmt.Sprintf("missing option %q", e.Option)
}

func (e *MissingOptionError) Is(target error) bool {
	if target, ok := target.(*MissingOptionError); ok {
		return reflect.ValueOf(e.Option).Pointer() == reflect.ValueOf(target.Option).Pointer()
	}

	return false
}

type WithScreenshotError struct {
	Err        error
	Screenshot []byte
}

func (e *WithScreenshotError) Error() string {
	return e.Err.Error()
}

func (e *WithScreenshotError) Unwrap() error {
	return e.Err
}

type WithLoggers struct {
	Info  *log.Logger
	Error *log.Logger
}

func (o *WithLoggers) Apply(instance interface{}) error {
	if chrome, ok := instance.(*chromeLogin); ok {
		if o.Info != nil {
			chrome.infoLogger = o.Info
		}

		if o.Error != nil {
			chrome.errorLogger = o.Error
		}

		return nil
	}

	return &InvalidTypeOptionError{instance: instance}
}
