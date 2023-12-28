package digipoauth

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"

	oauth2v4 "github.com/go-oauth2/oauth2/v4"
	digiconfig "github.com/holyhope/digiposte-oauth/config"
)

type AccessGenerator struct {
	setter digiconfig.Setter
	login  LoginMethod
}

var _ oauth2v4.AccessGenerate = (*AccessGenerator)(nil)

const RefreshTokenLength = 32

func (ag *AccessGenerator) Token(
	ctx context.Context,
	generateBasic *oauth2v4.GenerateBasic,
	isGenRefresh bool,
) (string, string, error) {
	if err := generateBasic.Request.ParseForm(); err != nil {
		return "", "", fmt.Errorf("parse form: %w", err)
	}

	creds := &Credentials{
		Username:  generateBasic.Client.GetID(),
		Password:  generateBasic.Client.GetSecret(),
		OTPSecret: generateBasic.Request.Form.Get("otp_secret"),
	}

	if err := areCredentialsValid(creds); err != nil {
		return "", "", fmt.Errorf("invalid credentials: %w", err)
	}

	digiposteToken, cookies, err := ag.login.Login(ctx, creds)
	if err != nil {
		return "", "", fmt.Errorf("using chrome: %w", err)
	}

	if err := digiconfig.SetCookies(ag.setter, cookies); err != nil {
		return "", "", fmt.Errorf("set cookies: %w", err)
	}

	if !isGenRefresh {
		return digiposteToken.AccessToken, "", nil
	}

	if digiposteToken.RefreshToken == "" {
		refreshTokenRunes := make([]byte, base64.RawStdEncoding.DecodedLen(RefreshTokenLength))
		if _, err := rand.Read(refreshTokenRunes); err != nil {
			return "", "", fmt.Errorf("generate refresh token: %w", err)
		}

		refreshToken := bytes.NewBuffer(make([]byte, 0, RefreshTokenLength))

		enc := base64.NewEncoder(base64.StdEncoding, refreshToken)
		defer enc.Close()

		if _, err := enc.Write(refreshTokenRunes); err != nil {
			return "", "", fmt.Errorf("encode refresh token: %w", err)
		}

		digiposteToken.RefreshToken = refreshToken.String()
	}

	return digiposteToken.AccessToken, digiposteToken.RefreshToken, nil
}

var ErrNilCredentials = errors.New("nil credentials")

func areCredentialsValid(creds *Credentials) error {
	if creds == nil {
		return ErrNilCredentials
	}

	if creds.Username == "" {
		return &RequiredFieldError{Field: "username"}
	}

	if creds.Password == "" {
		return &RequiredFieldError{Field: "password"}
	}

	return nil
}

type RequiredFieldError struct {
	Field string
}

func (e *RequiredFieldError) Error() string {
	return fmt.Sprintf("missing field %q", e.Field)
}
