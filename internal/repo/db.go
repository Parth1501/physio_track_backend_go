package repo

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	_ "github.com/sijms/go-ora/v2"
	goora "github.com/sijms/go-ora/v2"
)

type DBConfig struct {
	User          string
	Password      string
	ConnectString string // TNS alias, e.g., sf1qflnhz887u1f0_high
	TNSAdmin      string // wallet dir
}

// NewDB connects to Oracle using godror and the wallet/TNS alias.
func NewDB(ctx context.Context, cfg DBConfig) (*sql.DB, error) {
	cleanDir := strings.ReplaceAll(cfg.TNSAdmin, `\`, `/`)

	tns, err := resolveTNS(cfg.ConnectString, cfg.TNSAdmin)
	if err != nil {
		return nil, fmt.Errorf("parse tns: %w", err)
	}

	opts := map[string]string{
		"WALLET":     cleanDir,
		"SSL":        "enable",
		"SSL VERIFY": "true",
	}
	dsn := goora.BuildUrl(tns.host, tns.port, tns.service, cfg.User, cfg.Password, opts)

	fmt.Printf("DB connect host=%s port=%d service=%s\n", tns.host, tns.port, tns.service)

	db, err := sql.Open("oracle", dsn)
	if err != nil {
		return nil, err
	}
	db.SetMaxIdleConns(2)
	db.SetMaxOpenConns(10)
	db.SetConnMaxIdleTime(15 * time.Minute)
	db.SetConnMaxLifetime(2 * time.Hour)

	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("db ping failed: %w", err)
	}
	return db, nil
}

type tnsEntry struct {
	host    string
	port    int
	service string
}

var (
	hostRe    = regexp.MustCompile(`host=([^)]+)`)
	portRe    = regexp.MustCompile(`port=([0-9]+)`)
	serviceRe = regexp.MustCompile(`service_name=([^)]+)`)
)

// resolveTNS parses tnsnames.ora to extract host/port/service for a given alias.
func resolveTNS(alias, tnsAdmin string) (tnsEntry, error) {
	data, err := os.ReadFile(filepath.Join(tnsAdmin, "tnsnames.ora"))
	if err != nil {
		return tnsEntry{}, err
	}
	content := string(data)
	lines := strings.Split(content, "\n")
	lowerAlias := strings.ToLower(alias)
	for _, line := range lines {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(strings.ToLower(trim), lowerAlias+" ") || strings.HasPrefix(strings.ToLower(trim), lowerAlias+"=") {
			host := hostRe.FindStringSubmatch(strings.ToLower(trim))
			port := portRe.FindStringSubmatch(trim)
			service := serviceRe.FindStringSubmatch(strings.ToLower(trim))
			if len(host) < 2 || len(port) < 2 || len(service) < 2 {
				return tnsEntry{}, fmt.Errorf("alias %s not fully resolvable", alias)
			}
			p, _ := strconv.Atoi(port[1])
			return tnsEntry{
				host:    host[1],
				port:    p,
				service: service[1],
			}, nil
		}
	}
	return tnsEntry{}, fmt.Errorf("alias %s not found in tnsnames.ora", alias)
}
