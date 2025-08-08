package config

type Server struct {
	Address        string   `env:"ADDRESS" envDefault:":6002"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:6001"`
}
