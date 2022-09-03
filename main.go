package main

import (
	"context"
	"log"
	"os"

	"git.tcp.direct/kayos/sendkeys"
	"github.com/getlantern/systray"
)

var (
	keycode []byte
	cancel  context.CancelFunc
	kb      *sendkeys.KBWrap
)

func init() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	var err error
	kb, err = sendkeys.NewKBWrapWithOptions(sendkeys.Noisy)
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	f, err := os.OpenFile(configpath+"log.log", os.O_RDWR|os.O_CREATE|os.O_TRUNC, filemode)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	log.SetOutput(f)
	keycode = initapp()
	hideConsole()
	defer bw_lock()
	systray.Run(onReady, onExit)
}

func onExit() {
	cancel()
	log.Println("Exit")
}

func onReady() {
	var ctx context.Context
	ctx, cancel = context.WithCancel(context.Background())
	go registerHotKeys(ctx, kb)
	systray.SetTemplateIcon(iconb, iconb)
	systray.SetTitle("Bw")
	systray.SetTooltip("Bw-Autotype")

	mKey1 := systray.AddMenuItem("Hotkey 1", "Change hotkey 1")
	mKey2 := systray.AddMenuItem("Hotkey 2", "Change hotkey 2")
	systray.AddSeparator()
	mSync := systray.AddMenuItem("Sync", "Sync BW")
	mLock := systray.AddMenuItem("Lock", "Lock BW")
	mLogin := systray.AddMenuItem("Login", "Login to BW")
	mLogout := systray.AddMenuItem("Logout", "Logout of BW")
	mQuit := systray.AddMenuItem("Quit", "Quit the app")
	err := bw_login()
	mSync.Enable()
	mLock.Enable()
	mLogin.Disable()
	mLogout.Enable()
	if err != nil {
		log.Println(err)
		mSync.Disable()
		mLock.Disable()
		mLogin.Enable()
		mLogout.Disable()
	}
	go func() {
		for {
			select {
			case <-mKey1.ClickedCh:
				err := edithotkey1()
				if err != nil {
					log.Println(err)
					break
				}
				cancel()
				var ctx context.Context
				ctx, cancel = context.WithCancel(context.Background())
				registerHotKeys(ctx, kb)
			case <-mKey2.ClickedCh:
				err := edithotkey2()
				if err != nil {
					log.Println(err)
					break
				}
				cancel()
				var ctx context.Context
				ctx, cancel = context.WithCancel(context.Background())
				registerHotKeys(ctx, kb)
			case <-mSync.ClickedCh:
				bw_lock()
			case <-mLock.ClickedCh:
				bw_lock()
				mSync.Enable()
				mLock.Disable()
				mLogin.Enable()
				mLogout.Enable()
			case <-mLogin.ClickedCh:
				err := bw_login()
				if err != nil {
					log.Println(err)
					break
				}
				mSync.Enable()
				mLock.Enable()
				mLogin.Disable()
				mLogout.Enable()
			case <-mLogout.ClickedCh:
				bw_logout()
				mSync.Disable()
				mLock.Disable()
				mLogin.Enable()
				mLogout.Disable()
				err := rmvconf()
				if err != nil {
					log.Println(err)
				}
			case <-mQuit.ClickedCh:
				systray.Quit()
			}
		}
	}()
}
