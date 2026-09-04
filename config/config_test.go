package config

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadDefaults(t *testing.T) {
	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Server.Port <= 0 {
		t.Errorf("expected positive port, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host == "" {
		t.Errorf("expected default host, got empty string")
	}
	if cfg.Database.URL == "" {
		t.Errorf("expected default database url, got empty string")
	}
	if cfg.Auth.OTPTTLMinutes <= 0 {
		t.Errorf("expected positive OTPTTLMinutes, got %d", cfg.Auth.OTPTTLMinutes)
	}
	if cfg.Auth.PasswordResetTTLMinutes <= 0 {
		t.Errorf("expected positive PasswordResetTTLMinutes, got %d", cfg.Auth.PasswordResetTTLMinutes)
	}
}

func TestLoadEnvOverrides(t *testing.T) {
	os.Setenv("PORT", "9090")
	os.Setenv("SERVER_HOST", "127.0.0.1")
	os.Setenv("DATABASE_URL", "postgres://user:secret@postgres.cloud:5432/co_drive_prod?sslmode=require")
	os.Setenv("OTP_TTL_MINUTES", "5")
	os.Setenv("PASSWORD_RESET_TTL_MINUTES", "20")
	os.Setenv("SMTP_HOST", "smtp.custom.com")
	os.Setenv("SMTP_PORT", "2525")
	os.Setenv("SMTP_USER", "smtp_user")
	os.Setenv("SMTP_PASS", "smtp_pass")
	os.Setenv("SMTP_FROM", "alerts@custom.com")

	defer func() {
		os.Unsetenv("PORT")
		os.Unsetenv("SERVER_HOST")
		os.Unsetenv("DATABASE_URL")
		os.Unsetenv("OTP_TTL_MINUTES")
		os.Unsetenv("PASSWORD_RESET_TTL_MINUTES")
		os.Unsetenv("SMTP_HOST")
		os.Unsetenv("SMTP_PORT")
		os.Unsetenv("SMTP_USER")
		os.Unsetenv("SMTP_PASS")
		os.Unsetenv("SMTP_FROM")
	}()

	cfg, err := Load("")
	if err != nil {
		t.Fatalf("unexpected error loading config: %v", err)
	}

	if cfg.Server.Port != 9090 {
		t.Errorf("expected port 9090 from PORT env, got %d", cfg.Server.Port)
	}
	if cfg.Server.Host != "127.0.0.1" {
		t.Errorf("expected host 127.0.0.1, got %s", cfg.Server.Host)
	}
	if cfg.Database.URL != "postgres://user:secret@postgres.cloud:5432/co_drive_prod?sslmode=require" {
		t.Errorf("expected db url postgres://user:secret@postgres.cloud:5432/co_drive_prod?sslmode=require, got %s", cfg.Database.URL)
	}
	if cfg.Auth.OTPTTLMinutes != 5 {
		t.Errorf("expected OTPTTLMinutes 5, got %d", cfg.Auth.OTPTTLMinutes)
	}
	if cfg.Auth.PasswordResetTTLMinutes != 20 {
		t.Errorf("expected PasswordResetTTLMinutes 20, got %d", cfg.Auth.PasswordResetTTLMinutes)
	}
	if cfg.SMTP.Host != "smtp.custom.com" {
		t.Errorf("expected smtp host smtp.custom.com, got %s", cfg.SMTP.Host)
	}
	if cfg.SMTP.Port != 2525 {
		t.Errorf("expected smtp port 2525, got %d", cfg.SMTP.Port)
	}
	if cfg.SMTP.Username != "smtp_user" {
		t.Errorf("expected smtp user smtp_user, got %s", cfg.SMTP.Username)
	}
	if cfg.SMTP.Password != "smtp_pass" {
		t.Errorf("expected smtp pass smtp_pass, got %s", cfg.SMTP.Password)
	}
	if cfg.SMTP.FromEmail != "alerts@custom.com" {
		t.Errorf("expected smtp from alerts@custom.com, got %s", cfg.SMTP.FromEmail)
	}
}

// TestStrictEnvIsolationInvariant statically inspects the codebase using the Go AST
// to guarantee that no non-test .go file in the workspace calls os.Getenv or os.LookupEnv
// outside of config/config.go.
func TestStrictEnvIsolationInvariant(t *testing.T) {
	rootPath, err := filepath.Abs("..")
	if err != nil {
		t.Fatalf("failed to determine repository root: %v", err)
	}

	fset := token.NewFileSet()
	var violations []string

	err = filepath.Walk(rootPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Skip vendor, git, temporary dirs, hidden dirs, and test files
		if info.IsDir() {
			base := filepath.Base(path)
			if strings.HasPrefix(base, ".") || base == "vendor" || base == "bin" {
				return filepath.SkipDir
			}
			return nil
		}

		if !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return nil
		}

		relPath, _ := filepath.Rel(rootPath, path)
		// config/config.go is the ONLY allowed file
		if relPath == filepath.Join("config", "config.go") {
			return nil
		}

		node, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return nil // Skip unparseable files
		}

		ast.Inspect(node, func(n ast.Node) bool {
			call, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}

			sel, ok := call.Fun.(*ast.SelectorExpr)
			if !ok {
				return true
			}

			ident, ok := sel.X.(*ast.Ident)
			if !ok {
				return true
			}

			if ident.Name == "os" && (sel.Sel.Name == "Getenv" || sel.Sel.Name == "LookupEnv" || sel.Sel.Name == "Environ") {
				pos := fset.Position(call.Pos())
				violations = append(violations, pos.String())
			}

			return true
		})

		return nil
	})

	if err != nil {
		t.Fatalf("failed to walk repository tree: %v", err)
	}

	if len(violations) > 0 {
		t.Fatalf("Architectural invariant violated! Environment variables must only be accessed from config/config.go. Found unauthorized calls:\n%s", strings.Join(violations, "\n"))
	}
}
