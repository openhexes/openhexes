package config

type Server struct {
	Address        string   `env:"ADDRESS" envDefault:":8080"`
	AllowedOrigins []string `env:"ALLOWED_ORIGINS" envDefault:"http://localhost:5173"`
}
