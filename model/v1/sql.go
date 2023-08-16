package v1

import (
	"context"
	"fmt"
)

// MySQLOpts holds MySQL configurations
type MySQLOpts struct {
	Username string
	Password string
	Host     string
	Port     string
	DBName   string
}
type MySqlConfig struct {
	DSN string
}

func NewMySQL(ctx context.Context, config MySQLOpts) (MySqlConfig, error) {
	if config.Username == "" || config.Password == "" || config.Host == "" || config.Port == "" || config.DBName == "" {
		return MySqlConfig{}, fmt.Errorf("参数不对")
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=true", config.Username, config.Password, config.Host, config.Port, config.DBName)
	return MySqlConfig{dsn}, nil
}
