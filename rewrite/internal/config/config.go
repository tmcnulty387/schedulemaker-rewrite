package config

import (
	"fmt"
	"os"
)

type Config struct {
	// Database
	DatabaseServer string // MySQL hostname
	DatabaseUser   string // MySQL username
	DatabasePass   string // MySQL password
	DatabaseDB     string // MySQL database name
	CookieStore    string // Path for SIS session cookies. Default: `/tmp/sis_cookies.json`

	// Application
	Addr            string // Hostname. Default: localhost
	Port            string // Port number. Default: 8080
	HttpRootAddress string // Public base URL with trailing slash (https://schedule.csh.rit.edu/)
	ServerType      string // Server type ("production", "development") Default: "development"

	// Google Analytics and RUM
	GoogleAnalytics1 string // GA tracking ID used when SERVER_TYPE=production
	GoogleAnalytics2 string // GA tracking ID used when SERVER_TYPE=development
	RumClientToken   string // Datadog RUM client token
	RumApplicationID string // Datadog RUM application ID

	// S3 Image Storage
	S3Server      string // S3-compatible endpoint URL. Default: https://s3.csh.rit.edu
	S3Key         string // S3 access key
	S3Secret      string // S3 secret key
	S3ImageBucket string // S3 bucket name. Default: schedulemaker

	// Course Data Import
	DumpClasses   string // Course data dump file. Default: `/mnt/share/cshclass.dat`
	DumpClassAttr string // Course attribute data dump file. Default: `/mnt/share/cshattrib.dat`
	DumpInstruct  string // Instructor data dump file. Default: `/mnt/share/cshinstr.dat`
	DumpMeeting   string // Meeting patterns dump file. Default: `/mnt/share/cshmtgpat.dat`
	DumpNotes     string // Course notes dump file. Default: `/mnt/share/cshnotes.dat`
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}

	return defaultValue
}

func Load() Config {
	return Config{
		DatabaseServer: getEnv("DATABASE_SERVER", "localhost"),
		DatabaseUser:   getEnv("DATABASE_USER", "schedulemaker"),
		DatabasePass:   getEnv("DATABASE_PASS", "schedulemaker"),
		DatabaseDB:     getEnv("DATABASE_DB", "schedulemaker"),
		CookieStore:    getEnv("COOKIE_STORE", "/tmp/sis_cookies.json"),

		Addr:       getEnv("ADDR", "localhost"),
		Port:       getEnv("PORT", "8080"),
		ServerType: getEnv("SERVER_TYPE", "development"),

		GoogleAnalytics1: getEnv("GOOGLEANALYTICS1", ""),
		GoogleAnalytics2: getEnv("GOOGLEANALYTICS2", ""),
		RumClientToken:   getEnv("RUM_CLIENT_TOKEN", ""),
		RumApplicationID: getEnv("RUM_APPLICATION_ID", ""),

		S3Server:      getEnv("S3_SERVER", "https://s3.csh.rit.edu"),
		S3Key:         getEnv("S3_KEY", ""),
		S3Secret:      getEnv("S3_SECRET", ""),
		S3ImageBucket: getEnv("S3_IMAGE_BUCKET", "schedulemaker"),

		DumpClasses:   getEnv("DUMPCLASSES", "/mnt/share/cshclass.dat"),
		DumpClassAttr: getEnv("DUMPCLASSATTR", "/mnt/share/cshattrib.dat"),
		DumpInstruct:  getEnv("DUMPINSTRUCT", "/mnt/share/cshinstr.dat"),
		DumpMeeting:   getEnv("DUMPMEETING", "/mnt/share/cshmtgpat.dat"),
		DumpNotes:     getEnv("DUMPNOTES", "/mnt/share/cshnotes.dat"),
	}
}

func (c *Config) GetDataSourceName() string {
	return fmt.Sprintf("%s:%s@tcp(%s)/%s?parseTime=true&collation=utf8mb4_unicode_ci",
		c.DatabaseUser, c.DatabasePass, c.DatabaseServer, c.DatabaseDB)
}