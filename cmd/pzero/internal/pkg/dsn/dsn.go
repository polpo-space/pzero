// Unless explicitly stated otherwise all files in this repository are licensed
// under the Apache License Version 2.0.
// This product includes software developed at Datadog (https://www.datadoghq.com/).
// Copyright 2016 Datadog, Inc.

package dsn

import (
	"errors"
	"net/url"
	"strings"
)

const (
	Database = "database"
	User     = "user"
	Password = "password"
	Host     = "host"
	Port     = "port"
)

// ParseDSN parses postgres/pgx DSNs into a map of key/value pairs.
func ParseDSN(driverName, dsn string) (meta map[string]string, err error) {
	meta = make(map[string]string)
	switch driverName {
	case "postgres", "pgx":
		meta, err = parsePostgresDSN(dsn)
		if err != nil {
			return meta, err
		}
	default:
		// Try to parse the DSN and see if the scheme contains a known driver name.
		u, e := url.Parse(dsn)
		if e != nil {
			return meta, e
		}
		if driverName != u.Scheme {
			// In some cases the driver is registered under a non-official name.
			// For example, "Test" may be the registered name with a DSN of "postgres://postgres:postgres@127.0.0.1:5432/fakepreparedb"
			// for the purposes of testing/mocking.
			// In these cases, we try to parse the DSN based upon the DSN itself, instead of the registered driver name
			return ParseDSN(u.Scheme, dsn)
		}
		return meta, &url.Error{Op: "parse", URL: dsn, Err: errors.New("unsupported driver")}
	}
	return meta, nil
}

// parsePostgresDSN parses a postgres-type dsn into a map.
func parsePostgresDSN(dsn string) (map[string]string, error) {
	var err error
	if strings.HasPrefix(dsn, "postgres://") || strings.HasPrefix(dsn, "postgresql://") {
		// url form, convert to opts
		dsn, err = parseURL(dsn)
		if err != nil {
			return nil, err
		}
	}
	meta := make(map[string]string)
	if err := parseOpts(dsn, meta); err != nil {
		return nil, err
	}
	// remove sensitive information
	delete(meta, "password")
	return meta, nil
}
