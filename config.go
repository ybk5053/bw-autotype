package main

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/ncruces/zenity"
	"github.com/spf13/viper"
	"github.com/zalando/go-keyring"
	"golang.design/x/hotkey"
)

const (
	bwofficialurl      = "https://vault.bitwarden.com"
	service            = "bw-autotype"
	user               = "bw-autotype"
	noconferr          = "config not set"
	serverconf         = "serverurl"
	userconf           = "username"
	passconf           = "password"
	sslconf            = "sslpath"
	filemode           = 0660
	hotkey1conf        = "hotkey1"
	hotkey2conf        = "hotkey2"
	hotkey1patternconf = "hotkey1pattern"
	hotkey2patternconf = "hotkey2pattern"
)

var (
	configpath string
	bw         string
)

func init() {
	path, err := os.Executable()
	if err != nil {
		log.Fatal(err)
	}
	if runtime.GOOS == "windows" {
		bw = filepath.Dir(path) + "/bw.exe"
		configpath = os.Getenv("HOMEDRIVE") + os.Getenv("HOMEPATH") + "/AppData/Roaming/Bitwarden CLI/"
	}
	if runtime.GOOS == "linux" {
		bw = filepath.Dir(path) + "/bw"
		configpath = os.Getenv("HOME") + "/.config/Bitwarden CLI/"
	}
	viper.SetConfigName("config")
	viper.SetConfigType("yaml")
	viper.AddConfigPath(configpath)
	log.Println(configpath)
	viper.SetDefault(hotkey1conf, "{Ctrl}{Alt}{A}")
	viper.SetDefault(hotkey2conf, "{Ctrl}{Alt}{S}")
	viper.SetDefault(hotkey1patternconf, "{Username}")
	viper.SetDefault(hotkey2patternconf, "{Password}")
	err = viper.ReadInConfig()
	if err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			log.Fatal(err)
		}
		viper.WriteConfigAs(configpath + "config.yml")
	}
	viper.WriteConfig()
	log.Println(viper.ConfigFileUsed())
}

func initapp() []byte {
	var bkey []byte
	key, err := keyring.Get(service, user)
	if err != nil {
		log.Println(err)
		bkey = make([]byte, 32)
		_, err = rand.Read(bkey)
		if err != nil {
			log.Fatal(err)
		}
		key = base64.StdEncoding.EncodeToString(bkey)
		err = keyring.Set(service, user, key)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		bkey, err = base64.StdEncoding.DecodeString(key)
		if err != nil {
			log.Fatal(err)
		}
	}

	return bkey
}

func readconf(c string) (string, error) {
	res := viper.GetString(c)
	if res == "" {
		return "", errors.New(noconferr)
	}
	return res, nil
}

func writeconf(c, s string) error {
	viper.Set(c, s)
	err := viper.WriteConfig()
	log.Println("write")

	return err
}

func rmvconf() error {
	viper.Set(serverconf, "")
	viper.Set(userconf, "")
	viper.Set(passconf, "")
	viper.Set(sslconf, "")

	return viper.WriteConfig()
}

func get_server() (string, error) {
	server, err := readconf(serverconf)
	if err == nil {
		return server, nil
	}

	if err.Error() == noconferr {
		server, err = zenity.Entry("Server Url:", zenity.EntryText("https://debian.test/vault"))
		if err != nil {
			return "", err
		}
		if server == "" {
			server = bwofficialurl
		}
		err = writeconf(serverconf, server)
		if err != nil {
			return "", err
		}
		return server, nil
	}

	return "", err
}

func get_user_pass() (string, string, error) {
	user, err := readconf(userconf)
	if err == nil {
		epass, err := readconf(passconf)
		if err == nil {
			pass, err := decrypt(epass)
			if err != nil {
				return "", "", err
			}
			return user, pass, nil
		}
		if err.Error() == noconferr {
			user, pass, err := zenity.Password(zenity.Username(), zenity.Title("BW-Autotype"), zenity.DisallowEmpty())
			if err != nil {
				return "", "", err
			}
			err = writeconf(userconf, user)
			if err != nil {
				return "", "", err
			}
			epass, err := encrypt(pass)
			if err != nil {
				return "", "", err
			}
			err = writeconf(passconf, epass)
			if err != nil {
				return "", "", err
			}
			return user, pass, nil
		}
	}

	if err.Error() == noconferr {
		user, pass, err := zenity.Password(zenity.Username(), zenity.Title("BW-Autotype"), zenity.DisallowEmpty())
		if err != nil {
			return "", "", err
		}
		err = writeconf(userconf, user)
		if err != nil {
			return "", "", err
		}
		epass, err := encrypt(pass)
		if err != nil {
			return "", "", err
		}
		err = writeconf(passconf, epass)
		if err != nil {
			return "", "", err
		}
		return user, pass, nil
	}

	return "", "", err
}

func get_pass() (string, error) {
	epass, err := readconf(passconf)
	if err == nil {
		pass, err := decrypt(epass)
		if err != nil {
			return "", err
		}
		return pass, nil
	}
	if err.Error() == noconferr {
		_, pass, err := zenity.Password(zenity.Title("BW-Autotype"), zenity.DisallowEmpty())
		epass, err := encrypt(pass)
		if err != nil {
			return "", err
		}
		err = writeconf(passconf, epass)
		if err != nil {
			return "", err
		}
		return pass, nil
	}

	return "", err
}

func encrypt(p string) (string, error) {
	plain := []byte(p)
	block, err := aes.NewCipher(keycode)
	if err != nil {
		return "", err
	}
	cipherText := make([]byte, len(plain)+aes.BlockSize)
	iv := cipherText[:aes.BlockSize]
	if _, err = io.ReadFull(rand.Reader, iv); err != nil {
		return "", err
	}
	cfb := cipher.NewCFBEncrypter(block, iv)
	cfb.XORKeyStream(cipherText[aes.BlockSize:], plain)
	return base64.RawStdEncoding.EncodeToString(cipherText), nil
}

func decrypt(p string) (string, error) {
	block, err := aes.NewCipher(keycode)
	if err != nil {
		return "", err
	}
	cipherText, err := base64.RawStdEncoding.DecodeString(p)
	if err != nil {
		return "", err
	}
	if len(cipherText) < aes.BlockSize {
		err = errors.New("Ciphertext block size is too short!")
		return "", err
	}
	iv := cipherText[:aes.BlockSize]
	cipherText = cipherText[aes.BlockSize:]
	cfb := cipher.NewCFBDecrypter(block, iv)
	cfb.XORKeyStream(cipherText, cipherText)

	return string(cipherText), nil
}

func get_sslpath() (string, error) {

	path, err := readconf(sslconf)
	if err != nil {
		if err.Error() == noconferr {
			path, err = zenity.SelectFile(zenity.Title("BW-Autotype"), zenity.DisallowEmpty())
			if err == nil {
				writeconf(sslconf, path)
				return path, nil
			}
		}
		return "", err
	}
	return path, nil
}

func readhotkey1() (hotkey.Key, []hotkey.Modifier, []string, error) {
	return readhotkey(hotkey1conf, hotkey1patternconf)
}

func readhotkey2() (hotkey.Key, []hotkey.Modifier, []string, error) {
	return readhotkey(hotkey2conf, hotkey2patternconf)
}

func readhotkey(x, y string) (hotkey.Key, []hotkey.Modifier, []string, error) {
	key, err := readconf(x)
	if err != nil {
		return 0x00, nil, nil, err
	}
	keys := strings.FieldsFunc(key, splitkeymods)
	for i, v := range keys {
		keys[i] = strings.ToLower(v)
	}
	lastkey := keymap[keys[len(keys)-1]]
	modkeys := make([]hotkey.Modifier, 0)
	for i, v := range keys {
		if i == len(keys)-1 {
			break
		}
		modkeys = append(modkeys, modmap[v])
	}

	pattern, err := readconf(y)
	if err != nil {
		return 0x00, nil, nil, err
	}
	patterns := strings.FieldsFunc(pattern, splitkeymods)
	for i, v := range patterns {
		patterns[i] = strings.ToLower(v)
	}

	return lastkey, modkeys, patterns, nil
}

func splitkeymods(r rune) bool {
	return r == '{' || r == '}'
}

func edithotkey(c, p string) error {
	s, _ := readconf(c)
	s, err := zenity.Entry("Hotkey", zenity.EntryText(s))
	if err != nil {
		if err.Error() != "dialog canceled" {
			return err
		}
	}

	ss, _ := readconf(p)
	ss, err = zenity.Entry("Input", zenity.EntryText(ss))
	if err != nil {
		return err
	}
	err = writeconf(c, s)
	if err != nil {
		return err
	}
	err = writeconf(p, ss)
	if err != nil {
		return err
	}

	return nil
}

func edithotkey1() error {
	return edithotkey(hotkey1conf, hotkey1patternconf)
}

func edithotkey2() error {
	return edithotkey(hotkey2conf, hotkey2patternconf)
}
