package main

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	"git.tcp.direct/kayos/sendkeys"
	"github.com/ncruces/zenity"
	"golang.design/x/hotkey"
)

func listenHotkey(kb *sendkeys.KBWrap, c context.Context, key hotkey.Key, mods []hotkey.Modifier, pattern []string) {
	hk := hotkey.New(mods, key)

	err := hk.Register()
	if err != nil {
		log.Fatal(err)
	}

	// Blocks until the hotkey is triggered.
	for {
		select {
		case <-hk.Keyup():
			if !checkbwlogin() {
				break
			}
			name := getActiveWindowName()
			items, err := findPass(name)
			if err != nil {
				log.Println(err)
			} else {
				if len(items) > 1 {
					items, err = selectitem(items)
					if err != nil {
						if strings.Contains(err.Error(), "canceled") {
							break
						}
					}
					time.Sleep(500 * time.Millisecond)
				}
				//setActiveWindow(hwnd)
				entry := items[0]
				for _, v := range pattern {
					switch v {
					case "tab":
						kb.Tab()
					case "enter":
						kb.Enter()
					case "space":
						kb.Type(" ")
					default:
						kb.Type(entry[v])
					}
				}
			}
		case <-c.Done():
			hk.Unregister()
			return
		}
	}
}

func selectitem(items []map[string]string) ([]map[string]string, error) {
	list := make([]string, 0)
	for i, v := range items {
		list = append(list, fmt.Sprintf("%d: %s : %s", i, v["name"], v["username"]))
	}
	s, err := zenity.List("Select:", list, zenity.Title("BW-Autotype"), zenity.DisallowEmpty())
	if err != nil {
		return nil, err
	}
	i, err := strconv.ParseInt(strings.Split(s, ":")[0], 0, 0)
	if err != nil {
		log.Println(err)
	}

	return []map[string]string{items[i]}, nil
}

func copymod(old []hotkey.Modifier) []hotkey.Modifier {
	new := make([]hotkey.Modifier, len(old))
	copy(new, old)
	return new
}
