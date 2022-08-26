package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

var bw_session string

func bw_login() error {
	var cmd *exec.Cmd
	if checkbwlogin() {
		pass, err := get_pass()
		if err != nil {
			return err
		}
		cmd = exec.Command(bw, "unlock", pass)
	} else {
		server, err := get_server()
		if err != nil {
			return err
		}
		exec.Command(bw, "config", "server", server).Run()
		user, pass, err := get_user_pass()
		if err != nil {
			return err
		}
		cmd = exec.Command(bw, "login", user, pass)
		if server != bwofficialurl {
			path, err := get_sslpath()
			if err != nil {
				return err
			}
			cmd.Env = append(os.Environ(), fmt.Sprintf("NODE_EXTRA_CA_CERTS=%s", path))
		}
	}

	b, err := cmd.CombinedOutput()
	if err != nil {
		log.Fatal(err, string(b))
	}

	temp := strings.Split(string(b), " ")
	if len(temp) < 20 {
		return errors.New("login error")
	}
	temp = strings.Split(temp[19], "\n")
	if len(temp) < 1 {
		return errors.New("login error")
	}
	temp = strings.Split(temp[0], `"`)
	if len(temp) < 2 {
		return errors.New("login error")
	}

	bw_session = temp[1]
	log.Println("Login")
	return nil
}

func bw_logout() {
	exec.Command(bw, "logout").Run()
	bw_session = ""
	log.Println("Logout")
}

func bw_lock() {
	exec.Command(bw, "lock").Run()
	bw_session = ""
	log.Println("Lock")
}

// func getPass(p, s string) string {
// 	//p = password or username
// 	//s = search string
// 	if bw_session == "" {
// 		bw_login()
// 	}
// 	b, err := exec.Command(bw, "get", p, s, "--session", bw_session).CombinedOutput()
// 	if err != nil {
// 		if !strings.Contains(string(b), "password") {
// 			//log.Println(string(b), err)
// 			return ""
// 		}
// 		bw_login()
// 		if err != nil {
// 			log.Fatal(err)
// 			return ""
// 		}
// 	}
// 	return string(b)
// }

// func bw_search(s string) []map[string]interface{} {
// 	b, err := exec.Command(bw, "list", "items", "--search", s, "--session", bw_session).CombinedOutput()
// 	if err != nil {
// 		log.Println(err)
// 	}
// 	i := make([]map[string]interface{}, 0)
// 	err = json.Unmarshal(b, &i)
// 	if err != nil {
// 		log.Println(err)
// 	}

// 	return i
// }

func bw_search_con(s string, c chan []map[string]interface{}, ctx context.Context) {
	b, err := exec.Command(bw, "list", "items", "--search", s, "--session", bw_session).CombinedOutput()
	if err != nil {
		log.Println(err)
		return
	}
	i := make([]map[string]interface{}, 0)
	err = json.Unmarshal(b, &i)
	if err != nil {
		log.Println(string(b))
		log.Println(err)
		return
	}

	if len(i) > 0 {
		select {
		case c <- i:
			return
		case <-ctx.Done():
			return
		}
	}
}

func findPass(name string) ([]map[string]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	windows := strings.FieldsFunc(name, splitmods)
	win := strings.TrimSpace(windows[len(windows)-1])
	find := make(chan []map[string]interface{})
	go bw_search_con(win, find, ctx)
	for i := len(windows) - 1; i >= 0; i-- {
		windows2 := strings.Split(win, " ")
		for i := len(windows2) - 1; i >= 0; i-- {
			go bw_search_con(windows2[i], find, ctx)
		}
	}
	items := <-find
	cancel()
	if len(items) == 0 {
		return nil, errors.New("login not found")
	}
	r := make([]map[string]string, 0)
	for _, item := range items {
		name := item["name"].(string)
		item = item["login"].(map[string]interface{})
		user := item["username"].(string)
		pass := item["password"].(string)
		m := map[string]string{"name": name, "username": user, "password": pass}
		r = append(r, m)
	}
	return r, nil
}

func splitmods(r rune) bool {
	return r == '-' || r == '—'
}

func checkbwlogin() bool {
	b, err := exec.Command(bw, "status").CombinedOutput()
	s := string(b)
	if err != nil {
		log.Fatal(err)
	}
	if strings.Contains(s, "unauthenticated") {
		return false
	}
	return true
}
