package main

import (
	"eternal-infer-worker/apis"
	"eternal-infer-worker/config"
	"eternal-infer-worker/libs/dockercmd"
	"eternal-infer-worker/libs/logger"
	_ "eternal-infer-worker/libs/logger"
	"eternal-infer-worker/manager"
	watcher "eternal-infer-worker/task-watcher"
	"fmt"
	"log"
	"os"
	"runtime/debug"
	"time"

	"eternal-infer-worker/tui"

	tea "github.com/charmbracelet/bubbletea"
)

var VersionTag string

func main() {
	defer func() {
		if r := recover(); r != nil {
			if tui.UI != nil {
				tui.UI.UpdateSectionText(tui.UIMessageData{
					Section: tui.UISectionStatusText,
					Color:   "danger",
					Text:    "Panic attack! 💀 ",
				})
			}
			log.Println("Panic attack", r)
			log.Println("stacktrace from panic: \n" + string(debug.Stack()))
		}
	}()

	var err error
	err = checkRequirement()
	if err != nil {
		panic(err)
	}

	cfg, err := config.ReadConfig()
	if err != nil {
		fmt.Println("Error reading config file: ", err)
		panic(err)
	}

	modelManager := manager.NewModelManager(cfg.ModelsDir, cfg.RPC, cfg.NodeMode, cfg.WorkerHub)

	newTaskWatcher, err := watcher.NewTaskWatcher(watcher.NetworkConfig{
		RPC: cfg.RPC,
		// WS:  *ws,
	}, cfg.WorkerHub, cfg.Account, cfg.ModelsDir, cfg.LighthouseAPI, cfg.NodeMode, 1, 1, modelManager, nil)
	if err != nil {
		panic(err)
	}

	stopChn := make(chan struct{}, 1)

	ui := tui.InitialModel(VersionTag, cfg.NodeMode, stopChn, newTaskWatcher, modelManager)
	tui.UI = &ui

	logger.DefaultLogger.SetTermPrinter(tui.UI.Print)
	go func() {
		p := tea.NewProgram(ui, tea.WithAltScreen())
		if _, err := p.Run(); err != nil {
			fmt.Printf("Alas, there's been an error: %v", err)
			os.Exit(1)
		}
	}()
	ui.UpdateSectionText(tui.UIMessageData{
		Section: tui.UISectionStatusText,
		Color:   "waiting",
		Text:    "Starting server...",
	})
	time.Sleep(1 * time.Second)

	shutdownEmitted := false
	go func() {
		for {
			select {
			case <-stopChn:
				shutdownEmitted = true
				tui.UI.UpdateSectionText(tui.UIMessageData{
					Section: tui.UISectionStatusText,
					Color:   "danger",
					Text:    "Shutting down..."})
				err = modelManager.RemoveAllInstanceDocker()
				if err != nil {
					panic(err)
				}
				time.Sleep(1 * time.Second)
				os.Exit(0)
			case <-time.After(1 * time.Second):
			}
		}
	}()

	go func() {
		for {
			select {
			case <-stopChn:
				if shutdownEmitted {
					tui.UI.UpdateSectionText(tui.UIMessageData{
						Section: tui.UISectionStatusText,
						Color:   "danger",
						Text:    "Force shutting down..."})
					time.Sleep(1 * time.Second)
					os.Exit(0)
				}
			case <-time.After(1 * time.Second):
			}
		}
	}()

	// go modelManager.WatchAndPreloadModels()

	go newTaskWatcher.Start()

	go func() {
		err = apis.InitRouter(cfg.Port, newTaskWatcher).StartRouter()
		if err != nil {
			panic(err)
		}
	}()
	select {}
}

func checkRequirement() error {
	err := checkDockerExist()
	if err != nil {
		return err
	}

	return nil
}

func checkDockerExist() error {
	ok := dockercmd.CheckDockerExist()
	if !ok {
		return fmt.Errorf("docker not found")
	}
	return nil
}

// func checkCondaExist() error {

// 	cmd := exec.Command("conda", "-V")

// 	out, err := cmd.Output()
// 	if err != nil {
// 		return err
// 	}

// 	versionParts := strings.Split(string(out), " ")
// 	if len(versionParts) < 2 {
// 		return fmt.Errorf("conda not found")
// 	}

// 	versions := strings.Split(versionParts[1], ".")
// 	if len(versions) < 2 {
// 		return fmt.Errorf("conda not found")
// 	}

// 	majorVersion, err := strconv.ParseInt(versions[0], 10, 64)
// 	if err != nil {
// 		return err
// 	}

// 	minorVersion, err := strconv.ParseInt(versions[1], 10, 64)
// 	if err != nil {
// 		return err
// 	}

// 	if majorVersion < 4 || (majorVersion == 4 && minorVersion < 8) {
// 		return fmt.Errorf("conda version is too low")
// 	}

// 	return nil
// }
