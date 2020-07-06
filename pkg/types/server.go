package types

type server struct {
	Host string
	Port int
	Ssl bool
}

var (
	ProductionServer = &server{
		Host: "service-prod.xdapp.com",
		Port: 8900,
		Ssl: true,
	}

	DevServer = &server{
		Host: "dev.xdapp.com",
		Port: 8100,
		Ssl: true,
	}

	GlobalServer = &server{
		Host: "service-gcp.xdapp.com",
		Port: 8900,
		Ssl: true,
	}
)
