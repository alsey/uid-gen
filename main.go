package main

import (
	"github.com/alsey/uid-gen/config"
	"github.com/alsey/uid-gen/health"
	"github.com/alsey/uid-gen/logger"
	"github.com/alsey/uid-gen/util"

	"database/sql"
	"fmt"
	"log"
	"net/http"
	"strconv"

	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/mux"
	"gopkg.in/redis.v5"
)

type Item struct {
	Step  int
	Value int
}

func main() {

	var (
		db  *sql.DB
		cli *redis.Client
		err error
	)

	// mysql
	mysql_dsn := config.GetMySqlDsn()

	if db, err = sql.Open("mysql", mysql_dsn); nil != err {
		logger.Fatal("failed to open %s, %v", mysql_dsn, err)
	}

	if err = db.Ping(); nil != err {
		logger.Fatal("failed to connect %s, %v", mysql_dsn, err)
	}

	defer db.Close()

	// redis
	redis_addr := config.GetRedisAddr()

	cli = redis.NewClient(&redis.Options{Addr: redis_addr})
	if _, err = cli.Ping().Result(); nil != err {
		logger.Fatal("failed to connect %s, %v", redis_addr, err)
	}

	// sync
	go func() {

		logger.Info("sync start...")

		// read data from mysql
		var (
			name  string
			step  int
			value int
			val   string
			item  *Item
			rows  *sql.Rows
			err   error
		)

		if rows, err = db.Query("select name, step, value from state"); nil != err {
			logger.Error("failed to query from mysql %s, %v", mysql_dsn, err)
			return
		}

		defer rows.Close()

		for rows.Next() {

			if err = rows.Scan(&name, &step, &value); nil != err {
				logger.Error("failed to read data from mysql %s, %v", mysql_dsn, err)
				continue
			}

			val, err = cli.Get(name).Result()

			if err == redis.Nil {

				item = &Item{
					Step:  step,
					Value: value,
				}

				var item_str string
				if item_str, err = util.Stringify(item); nil != err {
					logger.Error("failed stringify %v, %v", item, err)
					continue
				}

				if err = cli.Set(name, item_str, 0).Err(); nil != err {
					logger.Error("failed to set redis, name = %s, value = %s, %v", name, item_str, err)
					continue
				}

				logger.Info("set redis with key = %s, value = %s", name, item_str)
				continue
			} else if err != nil {
				logger.Error("read %s from redis %s failed, %v", name, redis_addr, err)
				continue
			}

			util.Parse(val, &item)

			val = strconv.Itoa(item.Value)

			if (step > 0 && item.Value > value) || (step < 0 && item.Value < value) {

				var stmt *sql.Stmt
				if stmt, err = db.Prepare("update state set value = ? where name = ?"); nil != err {
					logger.Error("failed prepared statement of update sql, name = %s, value = %s, %v", name, val, err)
					continue
				}

				if _, err = stmt.Exec(val, name); nil != err {
					logger.Error("failed to update mysql with name = %s, value = %s, %v", name, val, err)
					continue
				}

				logger.Info("update mysql with key = %s, value = %s", name, val)
			}
		}

		if err = rows.Err(); nil != err {
			logger.Error("failed to synchronize, %v", err)
		}
		
		logger.Info("sync END")
	}()

	// update mysql
	c := make(chan struct {
		name string
		Item
	})

	go func() {
		
		logger.Info("mysql writer start...")
		
		for v := range c {

			logger.Info("write to mysql %v", v)

			var stmt *sql.Stmt
			if stmt, err = db.Prepare("select * from state where name = ?"); nil != err {
				logger.Error("failed prepared statement, %v", err)
				continue
			}

			defer stmt.Close()

			var rows *sql.Rows
			if rows, err = stmt.Query(v.name); nil != err {
				logger.Error("failed to query mysql, name = %s, %v", v.name, err)
				continue
			}

			defer rows.Close()

			if rows.Next() {

				if stmt, err = db.Prepare("update state set value = ?, step = ? where name = ?"); nil != err {
					logger.Error("failed prepared statement, %v", err)
					continue
				}

				if _, err = stmt.Exec(v.Value, v.Step, v.name); nil != err {
					logger.Error("failed update mysql, name = %s, step = %s, value = %s, %v", v.name, v.Step, v.Value, err)
					continue
				}

			} else {

				if stmt, err = db.Prepare("insert into state (name, step, value) values (?, ?, ?)"); nil != err {
					logger.Error("failed prepared statement, %v", err)
					continue
				}

				if _, err = stmt.Exec(v.name, v.Step, v.Value); nil != err {
					logger.Error("failed insert into mysql, name = %s, step = %s, value = %s, %v", v.name, v.Step, v.Value, err)
					continue
				}

			}
		}
		
		logger.Info("mysql writer END")
	}()

	r := mux.NewRouter()

	r.HandleFunc("/health", health.Health)
	r.HandleFunc("/env", health.Env)
	r.HandleFunc("/favicon.ico", health.Favicon)

	r.HandleFunc("/{name}", func(w http.ResponseWriter, r *http.Request) {
		var (
			status int
			err    error
		)
		defer func() {
			if nil != err {
				http.Error(w, err.Error(), status)
			}
		}()

		vars := mux.Vars(r)
		name := vars["name"]

		if len(name) == 0 {
			status = http.StatusBadRequest
			return
		}

		start_str := r.URL.Query().Get("start")

		var (
			counter_start int
			is_start_set  = false
		)
		if len(start_str) > 0 {

			if counter_start, err = strconv.Atoi(start_str); nil != err {
				status = http.StatusBadRequest
				return
			}

			is_start_set = true
		}

		step_str := r.URL.Query().Get("step")

		var (
			counter_step int
			is_step_set  = false
		)
		if len(step_str) > 0 {

			if counter_step, err = strconv.Atoi(step_str); nil != err {
				status = http.StatusBadRequest
				return
			}

			is_step_set = true
		}

		var item Item

		item_str, err := cli.Get(name).Result()

		if err == redis.Nil {

			item.Step = 1
			if is_step_set {
				item.Step = counter_step
			}

			item.Value = 1
			if is_start_set {
				item.Value = counter_start
			}

		} else if err != nil {
			status = http.StatusInternalServerError
			return
		} else {
			var last_item *Item
			util.Parse(item_str, &last_item)

			item.Step = last_item.Step
			if is_step_set {
				item.Step = counter_step
			}

			item.Value = last_item.Value + item.Step
			if is_start_set {
				item.Value = counter_start
			}
		}

		logger.Info("counter %s step %d is %d", name, item.Step, item.Value)

		fmt.Fprintf(w, "%d", item.Value)

		go func() {
			
			c <- struct {
				name string
				Item
			}{name, item}
			
			if item_str, err = util.Stringify(item); nil != err {
				logger.Error("%v to string failed, %v", item, err)
				return
			}
	
			if err = cli.Set(name, item_str, 0).Err(); nil != err {
				logger.Error("set %s %s to redis failed, %v", name, item_str, err)
				return
			}			
			
		}()
	})

	port := config.GetServPort()
	logger.Info("listen on port %s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
