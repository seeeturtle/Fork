package config

type Config struct {
	DB     *DBConfig
	Emails []*EmailConfig
}

type DBConfig struct {
	Dialect  string
	Username string
	Password string
	Name     string
}

type EmailConfig struct {
	From     string
	To       string
	Password string
}

func GetConfig() *Config {
	return &Config{
		DB: &DBConfig{
			Dialect:  "postgres",
			Username: "bumblebee",
			Password: "fire2823",
			Name:     "schoolmealdb",
		},
	}
}
