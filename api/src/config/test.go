package config

type Test struct {
	Enabled bool `env:"ENABLED"`
	ID      string
	Tokens  Tokens `envPrefix:"TOKENS__"`
}

type Tokens struct {
	Owner      string `env:"OWNER" envDefault:"owner"`
	Unverified string `env:"UNVERIFIED" envDefault:"unverified"`
	Alfa       string `env:"ALFA" envDefault:"alfa"`
	Bravo      string `env:"BRAVO" envDefault:"bravo"`
}
