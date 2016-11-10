package session

import (
	"encoding/base64"
	"net/http"
	"net/url"
	"time"

	"golang.org/x/net/context"
	"gottb.io/goru"
	"gottb.io/goru/crypto"
)

const (
	COOKIE_PREFIX = "goru_"
)

type cookieStore struct {
}

func (s *cookieStore) Exists(ctx *goru.Context, group, key string) bool {
	values, err := s.load(ctx, group)
	if err != nil {
		return false
	}
	_, ok := values[key]
	return ok
}

func (s *cookieStore) Get(ctx *goru.Context, group, key string) (string, bool) {
	values, err := s.load(ctx, group)
	if err != nil {
		return "", false
	}
	value, ok := values[key]
	return value, ok
}

func (s *cookieStore) Set(ctx *goru.Context, group, key, value string) {
	values, err := s.load(ctx, group)
	if err != nil {
		return
	}
	values[key] = value
	s.save(ctx, group, values)
}

func (s *cookieStore) Remove(ctx *goru.Context, group, key string) {
	values, err := s.load(ctx, group)
	if err != nil {
		return
	}
	delete(values, key)
	s.save(ctx, group, values)
}

func (s *cookieStore) Clear(ctx *goru.Context, group string) {
	s.remove(ctx, group)
}

func (s *cookieStore) load(ctx *goru.Context, group string) (map[string]string, error) {
	valueStore, ok := ctx.NetContext.Value("cookie-store").(map[string]map[string]string)
	if ok {
		values, ok := valueStore[group]
		if ok {
			return values, nil
		}
	}
	if valueStore == nil {
		valueStore = make(map[string]map[string]string)
	}
	defer func() {
		ctx.NetContext = context.WithValue(ctx.NetContext, "cookie-store", valueStore)
	}()
	cookie, err := ctx.Request.Cookie(COOKIE_PREFIX + group)
	values := make(map[string]string)
	if err != nil {
		valueStore[group] = values
		return values, nil
	}
	urlValues, err := s.decryptCookie(cookie.Value)
	if err != nil {
		return nil, err
	}
	for k, v := range urlValues {
		if len(v) > 0 {
			values[k] = v[0]
		}
	}
	valueStore[group] = values
	return values, nil
}

func (s *cookieStore) save(ctx *goru.Context, group string, values map[string]string) error {
	valueStore, ok := ctx.NetContext.Value("cookie-store").(map[string]map[string]string)
	if !ok {
		valueStore = make(map[string]map[string]string)
	}
	valueStore[group] = values
	ctx.NetContext = context.WithValue(ctx.NetContext, "cookie-store", valueStore)
	encryptedValue, err := s.encryptCookie(values)
	if err != nil {
		return err
	}
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:  COOKIE_PREFIX + group,
		Value: encryptedValue,
	})
	return nil
}

func (s *cookieStore) remove(ctx *goru.Context, group string) error {
	valueStore, ok := ctx.NetContext.Value("cookie-store").(map[string]map[string]string)
	if !ok {
		valueStore = make(map[string]map[string]string)
	}
	values := make(map[string]string)
	valueStore[group] = values
	ctx.NetContext = context.WithValue(ctx.NetContext, "cookie-store", valueStore)
	encryptedValue, err := s.encryptCookie(values)
	if err != nil {
		return err
	}
	http.SetCookie(ctx.ResponseWriter, &http.Cookie{
		Name:    COOKIE_PREFIX + group,
		Value:   encryptedValue,
		Expires: time.Unix(0, 0),
		MaxAge:  -1,
	})
	return nil
}

func (s *cookieStore) decryptCookie(value string) (url.Values, error) {
	cipherText, err := base64.StdEncoding.DecodeString(value)
	if err != nil {
		return nil, err
	}
	plainText, err := crypto.Decrypt(cipherText)
	if err != nil {
		return nil, err
	}
	return url.ParseQuery(string(plainText))
}

func (s *cookieStore) encryptCookie(values map[string]string) (string, error) {
	urlValues := make(url.Values)
	for k, v := range values {
		urlValues.Add(k, v)
	}
	plainText := urlValues.Encode()
	cipherText, err := crypto.Encrypt([]byte(plainText))
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(cipherText), nil
}
