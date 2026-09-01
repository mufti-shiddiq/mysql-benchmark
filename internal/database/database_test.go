package database

import (
	"errors"
	"strings"
	"testing"

	"github.com/go-sql-driver/mysql"
)

func TestFriendlyConnectionErrorAccessDenied(t *testing.T) {
	err := friendlyConnectionError(&mysql.MySQLError{Number: 1045, Message: "Access denied"})
	if !strings.Contains(err.Error(), "access denied") {
		t.Fatalf("error = %q", err.Error())
	}
	if strings.Contains(err.Error(), "password=root") {
		t.Fatalf("error leaked credential detail: %q", err.Error())
	}
}

func TestFriendlyConnectionErrorOperationNotPermitted(t *testing.T) {
	err := friendlyConnectionError(errors.New("dial tcp 127.0.0.1:3306: operation not permitted"))
	if !strings.Contains(err.Error(), "sandboxed terminal") {
		t.Fatalf("error = %q", err.Error())
	}
}
