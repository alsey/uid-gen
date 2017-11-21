package config

import (
	"github.com/alsey/uid-gen/logger"
		
	"os"
	"net"
	"strconv"
)

var (
	mysql_dsn 	string
	redis_addr	string
	serv_port 	string
)

func init() {
	
	serv_port = env("PORT0", "3000")
	
	mysql_dsn = env("MYSQL_DSN", "root:password@tcp(127.0.0.1:3306)/counter?charset=utf8")
	logger.Info("mysql dsn is %s", mysql_dsn)

	redis_service := os.Getenv("REDIS_SERVICE")
	redis_host    := env("REDIS_HOST", "127.0.0.1")
	redis_port, _ := strconv.Atoi(env("REDIS_PORT", "6379"))

	if len(redis_service) > 0 {
			
		var (
			addrs 	[]*net.SRV
			err 	error
		)
		if _, addrs, err = net.LookupSRV("", "", redis_service); nil != err {
			logger.Error("dns lookup %s failed", redis_service)
			goto LookupServiceEnd
		}
			
		var hosts []string
		if hosts, err = net.LookupHost(addrs[0].Target); nil != err {
			logger.Error("lookup host %v failed", addrs)
			goto LookupServiceEnd
		}
		
		redis_host = hosts[0]
		redis_port = int(addrs[0].Port)
	}

	LookupServiceEnd:

	if len(redis_host) == 0 || redis_port == 0 {
		logger.Fatal("Environment variable(s) either REDIS_SERVICE or REDIS_HOST and REDIS_PORT are missing.")
	}
	
	redis_addr = net.JoinHostPort(redis_host, strconv.Itoa(redis_port))

	logger.Info("redis address is %s", redis_addr)
}

func GetMySqlDsn() string {
	return mysql_dsn
}

func GetRedisAddr() string {
	return redis_addr
}

func GetServPort() string {
	return serv_port
}

func env(nme string, def ...string) (val string) {
	val = os.Getenv(nme)
	if len(val) == 0 {
		if (len(def) > 0) {
			val = def[0]
		} else {
			logger.Fatal("Missing environment variable " + nme + ".")
		}
	}
	return
}