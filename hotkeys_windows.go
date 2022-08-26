package main

import (
	"context"
	"log"

	"git.tcp.direct/kayos/sendkeys"
	"golang.design/x/hotkey"
)

var (
	keymap = map[string]hotkey.Key{
		"1": hotkey.Key1,
		"2": hotkey.Key2,
		"3": hotkey.Key3,
		"4": hotkey.Key4,
		"5": hotkey.Key5,
		"6": hotkey.Key6,
		"7": hotkey.Key7,
		"8": hotkey.Key8,
		"9": hotkey.Key9,
		"0": hotkey.Key0,
		"a": hotkey.KeyA,
		"b": hotkey.KeyB,
		"c": hotkey.KeyC,
		"d": hotkey.KeyD,
		"e": hotkey.KeyE,
		"f": hotkey.KeyF,
		"g": hotkey.KeyG,
		"h": hotkey.KeyH,
		"i": hotkey.KeyI,
		"j": hotkey.KeyJ,
		"k": hotkey.KeyK,
		"l": hotkey.KeyL,
		"m": hotkey.KeyM,
		"n": hotkey.KeyN,
		"o": hotkey.KeyO,
		"p": hotkey.KeyP,
		"q": hotkey.KeyQ,
		"r": hotkey.KeyR,
		"s": hotkey.KeyS,
		"t": hotkey.KeyT,
		"u": hotkey.KeyU,
		"v": hotkey.KeyV,
		"w": hotkey.KeyW,
		"x": hotkey.KeyX,
		"y": hotkey.KeyY,
		"z": hotkey.KeyZ,
	}

	modmap = map[string]hotkey.Modifier{
		"ctrl":  hotkey.ModCtrl,
		"shift": hotkey.ModShift,
		"alt":   hotkey.ModAlt,
		"win":   hotkey.ModWin,
	}
)

func registerHotKeys(ctx context.Context, kb *sendkeys.KBWrap) {
	key, modkeys, pattern, err := readhotkey1()
	if err != nil {
		log.Println(err)
	} else {
		//register hotkey
		go listenHotkey(kb, ctx, key, copymod(modkeys), pattern)
	}

	key, modkeys, pattern, err = readhotkey2()
	if err != nil {
		log.Println(err)
	} else {
		//register hotkey
		go listenHotkey(kb, ctx, key, copymod(modkeys), pattern)
	}

	log.Println("Hotkey Registered")
}
