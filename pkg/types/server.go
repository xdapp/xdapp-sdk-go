package types

type Server struct {
	Host string
	Port int
	Ssl bool
}
const (
	EnvironmentProd   = "prod"
	EnvironmentDev    = "dev"
	EnvironmentGlobal = "global"
)

var (
	ProductionServer = &Server{
		Host: "service-prod.xdapp.com",
		Port: 8900,
		Ssl: true,
	}

	DevServer = &Server{
		Host: "dev.xdapp.com",
		Port: 8100,
		Ssl: true,
	}

	GlobalServer = &Server{
		Host: "service-gcp.xdapp.com",
		Port: 8900,
		Ssl: true,
	}
)
