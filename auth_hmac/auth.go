package auth_hmac

import (
	"fmt"
	"time"
	"strings"
	"context"
	"strconv"
)

type Auth struct {
	Expires		int
	Use_nonce	func(ctx context.Context, api_key, nonce string, timestamp int64, expire int) (bool, error)
	time		int64
}

func (a *Auth) Validate_time(header string) error {
	s := strings.TrimSpace(header)
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

func (a *Auth) Validate_nonce(ctx context.Context, api_key, header string) error {
	nonce := strings.TrimSpace(header)
	if nonce == "" {
		return fmt.Errorf("Nonce is missing")
	}
	
	ok, err := a.Use_nonce(ctx, api_key, nonce, a.time, a.Expires * 3)
	if err != nil {
		return fmt.Errorf("Could not validate nonce")
	}
	if !ok {
		return fmt.Errorf("Nonce is already used")
	}
	
	return nil
}