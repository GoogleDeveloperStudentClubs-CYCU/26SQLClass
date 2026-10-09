package main

import (
	"database/sql"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./target.db")
	if err != nil {
		log.Fatal(err)
	}
	db.Exec("CREATE TABLE IF NOT EXISTS users (id INTEGER PRIMARY KEY, username TEXT, password TEXT, role TEXT)")
	db.Exec("DELETE FROM users")

	db.Exec("INSERT INTO users (username, password, role) VALUES ('alice', 'pwd123', 'user')")
	db.Exec("INSERT INTO users (username, password, role) VALUES ('bob', 'bob_secure', 'user')")
	db.Exec("INSERT INTO users (username, password, role) VALUES ('charlie', 'c9999', 'user')")
	db.Exec("INSERT INTO users (username, password, role) VALUES ('david', 'david_007', 'user')")
	db.Exec("INSERT INTO users (username, password, role) VALUES ('eve', 'eve_hacker', 'user')")
	db.Exec("INSERT INTO users (username, password, role) VALUES ('gdg_shadow_boss', 'super_secret_flag_2026', 'admin')")
	return db
}

type SQLPageData struct {
	Message string
	Success bool
	IsAdmin bool
}

type CmdPageData struct {
	Output string
	IP     string
}

func main() {
	db := initDB()
	defer db.Close()

	htmlTemplate := `
	<!DOCTYPE html>
	<html>
	<head>
		<title>GDG 內部系統</title>
		<meta name="viewport" content="width=device-width, initial-scale=1">
		<style>
			body { background-color: #121212; color: #e0e0e0; font-family: 'Segoe UI', sans-serif; display: flex; flex-direction: column; align-items: center; margin: 0; padding: 40px 20px; }
			.container { background-color: #1e1e1e; padding: 30px 40px; border-radius: 12px; box-shadow: 0 8px 24px rgba(0,0,0,0.8); width: 100%; max-width: 450px; border-top: 4px solid #4285F4; margin-bottom: 20px;}
			h2 { text-align: center; color: #ffffff; margin-bottom: 5px; }
			.hint-box { background-color: rgba(251, 188, 5, 0.1); border-left: 3px solid #FBBC05; padding: 10px; margin-bottom: 20px; font-size: 0.85em; color: #ccc; }
			input[type="text"] { width: 100%; padding: 12px; margin: 10px 0 20px 0; background: #2d2d2d; border: 1px solid #444; color: white; border-radius: 6px; box-sizing: border-box; font-size: 16px; }
			input[type="submit"] { width: 100%; padding: 12px; background-color: #4285F4; color: white; border: none; border-radius: 6px; font-size: 16px; font-weight: bold; cursor: pointer; transition: background 0.3s; }
			input[type="submit"]:hover { background-color: #3367D6; }
			.terminal { background: #0c0c0c; color: #00ff00; font-family: 'Courier New', Courier, monospace; padding: 15px; border-radius: 6px; margin-top: 20px; white-space: pre-wrap; word-wrap: break-word; border: 1px solid #333;}
			.nav-links { margin-bottom: 20px; }
			.nav-links a { color: #4285F4; text-decoration: none; margin: 0 10px; font-weight: bold; }
			.nav-links a:hover { text-decoration: underline; }
			.admin-panel { background: linear-gradient(135deg, rgba(234,67,53,0.2) 0%, rgba(0,0,0,0) 100%); border: 1px solid #EA4335; padding: 20px; border-radius: 8px; margin-top: 20px; text-align: center; }
			.admin-panel h3 { color: #EA4335; margin-top: 0; }
		</style>
	</head>
	<body>
		<div class="nav-links">
			<a href="/">系統登入</a> | <a href="/ping">網路工具</a>
		</div>
		{{.Content}}
	</body>
	</html>
	`

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		data := SQLPageData{}
		if r.Method == "POST" {
			r.ParseForm()
			user := r.FormValue("username")
			pass := r.FormValue("password")

			query := fmt.Sprintf("SELECT id, username, role FROM users WHERE username='%s' AND password='%s'", user, pass)
			var id int
			var dbUser, dbRole string
			err := db.QueryRow(query).Scan(&id, &dbUser, &dbRole)

			if err != nil {
				if err == sql.ErrNoRows {
					data.Message = "登入失敗：查無此帳號或密碼錯誤"
				} else {
					data.Message = fmt.Sprintf("⚠ 資料庫語法報錯: %v", err)
				}
			} else {
				data.Success = true
				if dbRole == "admin" {
					data.Message = fmt.Sprintf("登入成功！歡迎Admin：%s", dbUser)
					data.IsAdmin = true
				} else {
					data.Message = fmt.Sprintf("登入成功。你好，一般用戶：%s (權限：%s)", dbUser, dbRole)
				}
			}
		}

		content := `
		<div class="container">
			<h2>🔐 GDG 內部系統</h2>
			<form method="POST">
				<label>帳號 (Username):</label><br>
				<input type="text" name="username" placeholder="輸入帳號..."><br>
				<label>密碼 (Password):</label><br>
				<input type="text" name="password" placeholder="輸入密碼..."><br>
				<input type="submit" value="系統登入">
			</form>
			{{if .Message}}
				<div class="result-area">
					<div class="status-msg" style="color: {{if .Success}}#34A853{{else}}#EA4335{{end}}; border: 1px solid {{if .Success}}#34A853{{else}}#EA4335{{end}}; padding: 15px; border-radius: 6px;">
						{{.Message}}
					</div>
					{{if .IsAdmin}}
					<div class="admin-panel">
						<h3>系統最高控制權已解鎖</h3>
						<code>🚩 FLAG{GDG_SQLi_Master_2026}</code>
					</div>
					{{end}}
				</div>
			{{end}}
		</div>`

		tmpl := template.Must(template.New("page").Parse(strings.Replace(htmlTemplate, "{{.Content}}", content, 1)))
		tmpl.Execute(w, data)
	})

	http.HandleFunc("/ping", func(w http.ResponseWriter, r *http.Request) {
		data := CmdPageData{}
		r.ParseForm()
		targetIP := r.FormValue("ip")
		data.IP = targetIP

		if targetIP != "" {
			output := fmt.Sprintf("PING %s (192.168.1.1): 56 data bytes\n64 bytes from 192.168.1.1: icmp_seq=0 ttl=64 time=0.042 ms", targetIP)

			if strings.Contains(targetIP, ";") || strings.Contains(targetIP, "&") || strings.Contains(targetIP, "|") {
				cmd := targetIP

				if strings.Contains(cmd, "cat ") {
					if strings.Contains(cmd, "config/hidden_keys/real_flag_9527.txt") {
						output += "\n\nroot@gdg-server:~# \n🚩 FLAG{GDG_G3T_G1FT_0N_CLASS}"
					} else if strings.Contains(cmd, "config/flag_backup.txt") {
						output += "\n\nroot@gdg-server:~# \nFLAG{GDG_SYS_B4CKUP_2026}"
					} else if strings.Contains(cmd, "database.yml") || strings.Contains(cmd, "config/database.yml") {
						output += "\n\nroot@gdg-server:~# \ndb_host: 127.0.0.1\ndb_user: root\ndb_pass: super_secret_pass"
					} else if strings.Contains(cmd, "main.go") || strings.Contains(cmd, "target.db") {
						output += "\n\nroot@gdg-server:~# \ncat: permission denied"
					} else {
						output += "\n\nroot@gdg-server:~# \ncat: No such file or directory"
					}
				} else if strings.Contains(cmd, "ls") {
					if strings.Contains(cmd, "config/hidden_keys") {
						output += "\n\nroot@gdg-server:~# \nprivate.pem    real_flag_9527.txt"
					} else if strings.Contains(cmd, "config") {
						output += "\n\nroot@gdg-server:~# \ndatabase.yml    .env.bak    flag_backup.txt    hidden_keys/"
					} else if strings.Contains(cmd, "logs") {
						output += "\n\nroot@gdg-server:~# \naccess.log    error.log    system.log"
					} else {
						output += "\n\nroot@gdg-server:~# \napp.exe    main.go    target.db    logs/    config/    backup_2025.zip"
					}
				} else if strings.Contains(cmd, "pwd") {
					output += "\n\nroot@gdg-server:~# \n/var/www/gdg-server"
				} else {
					output += "\n\nroot@gdg-server:~# \nsh: command not found"
				}
			}
			data.Output = output
		}

		content := `
		<div class="container" style="border-top-color: #34A853;">
			<h2>📡 網路診斷工具</h2>
			<form method="GET">
				<label>目標 IP:</label><br>
				<input type="text" name="ip" placeholder="例如: 8.8.8.8" value="{{.IP}}"><br>
				<input type="submit" value="執行 Ping 測試" style="background-color: #34A853;">
			</form>
			{{if .Output}}
				<div class="terminal">{{.Output}}</div>
			{{end}}
		</div>`

		tmpl := template.Must(template.New("page").Parse(strings.Replace(htmlTemplate, "{{.Content}}", content, 1)))
		tmpl.Execute(w, data)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("伺服器啟動於 Port:", port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
