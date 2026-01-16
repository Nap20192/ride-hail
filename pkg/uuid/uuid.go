package uuid

import (
	"crypto/rand"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"
)

type UUID [16]byte

var (
	ErrLength         = errors.New("not correct length")
	ErrNotValidFormat = errors.New("not valid format")
)

var Nil UUID

func New() UUID {
	var uuid UUID
	_, err := rand.Read(uuid[:])
	if err != nil {
		fmt.Println("Error generating UUID:", err)
		os.Exit(1)
	}
	return uuid
}

func (u UUID) String() string {
	return fmt.Sprintf("%x-%x-%x-%x-%x", u[0:4], u[4:6], u[6:8], u[8:10], u[10:])
}

func (u UUID) IsZero() bool {
	return u == UUID{}
}

func FromString(s string) (UUID, error) {
	var uuid UUID
	s = strings.ReplaceAll(s, "-", "")
	if len(s) != 32 {
		return UUID{}, fmt.Errorf("invalid UUID length: expected 32 hex characters, got %d", len(s))
	}
	for i := 0; i < 16; i++ {
		byteValue, err := strconv.ParseUint(s[i*2:i*2+2], 16, 8)
		if err != nil {
			return UUID{}, fmt.Errorf("invalid hex at position %d: %v", i*2, err)
		}
		uuid[i] = byte(byteValue)
	}

	return uuid, nil
}

func Validate(s string) error {
	if len(s) != 36 {
		return ErrLength
	}
	pattern := regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !pattern.MatchString(s) {
		return ErrNotValidFormat
	}
	return nil
}
