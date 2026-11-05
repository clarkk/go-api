package auth_hmac

import (
	"io"
	"fmt"
	"time"
	"bytes"
	"strings"
	"context"
	"strconv"
	"net/http"
	"crypto/hmac"
	"github.com/clarkk/go-util/hash"
)

const expiry_multiplier = 3

type Auth struct {
	API_key		string
	Request		*http.Request
	Hash_prefix	string
	Expires		int
	Use_nonce	func(ctx context.Context, hash_key string, expire int) (bool, error)
	time		int64
	nonce		string
}

func (a *Auth) Validate_time(header_name string) error {
	s := strings.TrimSpace(a.Request.Header.Get(header_name))
	if s == "" {
		return fmt.Errorf("Timestamp is missing")
	}
	
	var err error
	a.time, err = strconv.ParseInt(s, 10, 64)
	if err != nil {
		return fmt.Errorf("Parse timestamp: %w", err)
	}
	
	window	:= time.Duration(a.Expires) * time.Second
	delta	:= time.Since(time.Unix(a.time, 0))
	if delta < -window || delta > window {
		return fmt.Errorf("Timestamp outside allowed window")
	}
	
	return nil
}

func (a *Auth) Validate_nonce(header_name string) error {
	a.nonce = strings.TrimSpace(a.Request.Header.Get(header_name))
	if a.nonce == "" {
		return fmt.Errorf("Nonce is missing")
	}
	return nil
}

func (a *Auth) Validate_signature(id uint64, api_secret, received_signature string) (bool, int, error){
	signature, received, err := a.generate_signature(id)
	if err != nil {
		return false, received, err
	}
	
	expected	:= hash.HMAC_SHA256([]byte(api_secret), signature)
	actual		:= strings.ToLower(received_signature)
	
	if !hmac.Equal(expected, []byte(actual)) {
		return false, received, nil
	}
	
	if err := a.use_nonce(); err != nil {
		return false, received, err
	}
	
	return true, received, nil
}

func (a *Auth) generate_signature(id uint64) ([]byte, int, error){
	var (
		received	int
		b			bytes.Buffer
	)
	
	b.WriteString(a.Request.Method)
	b.WriteByte('\n')
	b.WriteString(a.Request.Host)
	b.WriteString(a.Request.URL.Path)
	if a.Request.URL.RawQuery != "" {
		b.WriteByte('?')
		b.WriteString(a.Request.URL.RawQuery)
	}
	b.WriteByte('\n')
	
	if a.Request.Method != http.MethodGet {
		body, err := io.ReadAll(a.Request.Body)
		if err != nil {
			return nil, 0, fmt.Errorf("Read request body: %w", err)
		}
		a.Request.Body.Close()
		
		received = len(body)
		
		b.Write(body)
		a.Request.Body = io.NopCloser(bytes.NewReader(body))
		b.WriteByte('\n')
	}
	
	b.WriteString(strconv.FormatUint(id, 10))
	b.WriteByte('\n')
	b.WriteString(a.API_key)
	b.WriteByte('\n')
	b.WriteString(strconv.FormatInt(a.time, 10))
	b.WriteByte('\n')
	b.WriteString(a.nonce)
	
	return b.Bytes(), received, nil
}

func (a *Auth) use_nonce() error {
	hash_key := fmt.Sprintf("%s:%s:%d:%s", a.Hash_prefix, a.API_key, a.time, a.nonce)
	ok, err := a.Use_nonce(a.Request.Context(), hash_key, a.Expires * expiry_multiplier)
	if err != nil {
		return fmt.Errorf("Could not validate nonce")
	}
	if !ok {
		return fmt.Errorf("Nonce is already used")
	}
	return nil
}