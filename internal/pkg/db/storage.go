package db

import (
	"context"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
)

type dsn struct {
	Host string
	Port string
	Name string
	User string
	Pass string
}

func (d dsn) String() string {
	dsnStr := "postgres://"

	if len(d.User) > 0 {
		dsnStr += d.User

		if len(d.Pass) > 0 {
			dsnStr += ":" + d.Pass
		}

		dsnStr += "@"
	}

	if len(d.Host) > 0 && len(d.Port) > 0 {
		dsnStr += d.Host + ":" + d.Port
	} else {
		dsnStr += "localhost:5432"
	}

	dsnStr += "/" + d.Name
	dsnStr += "?sslmode=disable"
	return dsnStr
}

func New(
	ctx context.Context,
	host string,
	port string,
	name string,
	user string,
	pass string,

) (*sqlx.DB, error) {
	return sqlx.ConnectContext(ctx, "postgres", dsn{
		Host: host,
		Port: port,
		Name: name,
		User: user,
		Pass: pass,
	}.String())
}
