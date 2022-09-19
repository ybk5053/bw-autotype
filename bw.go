package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
)

// var commonwords = []string{"the", "where", "how", "what", "or", "if", "when", "who", "and", "on"}
var bw_session string
var bw_items []map[string]string

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
		} else {
			writeconf(sslconf, "")
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

	err = bw_sync()
	if err != nil {
		return err
	}

	log.Println("Login")
	return nil
}

func bw_logout() {
	exec.Command(bw, "logout").Run()
	bw_session = ""
	bw_items = make([]map[string]string, 0)
	log.Println("Logout")
}

func bw_lock() {
	exec.Command(bw, "lock").Run()
	bw_session = ""
	bw_items = make([]map[string]string, 0)
	log.Println("Lock")
}

func bw_search(id string) (map[string]string, error) {
	b, err := exec.Command(bw, "get", "item", id, "--session", bw_session).CombinedOutput()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	i := make(map[string]interface{}, 0)
	err = json.Unmarshal(b, &i)
	if err != nil {
		log.Println(err)
		return nil, err
	}

	name := i["name"].(string)
	item := i["login"].(map[string]interface{})
	user := item["username"].(string)
	pass := item["password"].(string)
	m := map[string]string{"name": name, "username": user, "password": pass}

	return m, nil
}

func findPass(name string) ([]map[string]string, error) {
	res := make([]map[string]string, 0)
	for _, v := range bw_items {
		if strings.Contains(strings.ToLower(name), strings.ToLower(v["name"])) {
			res = append(res, v)
		}
	}
	if len(res) == 0 {
		return nil, errors.New("login not found")
	}
	return res, nil
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

func bw_sync() error {
	cmd := exec.Command(bw, "sync", "--session", bw_session)
	path, err := get_sslpath()
	if err != nil {
		return err
	}
	if path != "" {
		cmd.Env = append(os.Environ(), fmt.Sprintf("NODE_EXTRA_CA_CERTS=%s", path))
	}
	r, err := cmd.CombinedOutput()
	if err != nil {
		log.Println(err, string(r))
		return err
	}

	b, err := exec.Command(bw, "list", "items", "--session", bw_session).CombinedOutput()
	if err != nil {
		log.Println(err)
		return err
	}
	i := make([]map[string]interface{}, 0)
	err = json.Unmarshal(b, &i)
	if err != nil {
		log.Println(string(b))
		log.Println(err)
		return err
	}
	bw_items = make([]map[string]string, 0)
	log.Println(len(i))
	for _, v := range i {
		id, ok := v["id"].(string)
		if !ok {
			continue
		}
		name, ok := v["name"].(string)
		if !ok {
			name = ""
		}
		item, ok := v["login"].(map[string]interface{})
		if !ok {
			continue
		}
		user, ok := item["username"].(string)
		if !ok {
			user = ""
		}
		m := map[string]string{"id": id, "name": name, "user": user}
		bw_items = append(bw_items, m)
	}

	return nil
}
