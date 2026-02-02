package gamestate

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"github.com/fsnotify/fsnotify"
	"github.com/redis/go-redis/v9"
)

var (
	RDB     *redis.Client
	scripts = make(map[string]*redis.Script)
	permissions = make(map[string]string)
	mu          sync.RWMutex
)

func Init(addr string) error {
	RDB = redis.NewClient(&redis.Options{
		Addr: addr,
	})

	if err := RDB.Ping(context.Background()).Err(); err != nil {
		return fmt.Errorf("failed to connect to redis: %w", err)
	}

	if err := LoadScripts(); err != nil {
		return err
	}
	
	// Start Watcher for any scripts changes
	go WatchScripts()
	
	return nil
}

func WatchScripts() {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		fmt.Printf("ERROR: Failed to create watcher: %v\n", err)
		return
	}
	defer watcher.Close()

	// Resolve scripts path
	root := "scripts"
	if _, err := os.Stat(root); os.IsNotExist(err) {
		if _, err := os.Stat("../" + root); err == nil {
			root = "../" + root
		} else if _, err := os.Stat("../../" + root); err == nil {
			root = "../../" + root
		}
	}

	err = watcher.Add(root)
	if err != nil {
		fmt.Printf("ERROR: Failed to watch scripts dir: %v\n", err)
		return
	}
	
	fmt.Printf("Watching %s for changes...\n", root)

	for {
		select {
		case event, ok := <-watcher.Events:
			if !ok {
				return
			}
			if event.Has(fsnotify.Write) || event.Has(fsnotify.Create) {
				fmt.Printf("Script modified: %s, reloading...\n", event.Name)
				// For simplicity/safety, just reload all. 
				if err := LoadScripts(); err != nil {
					fmt.Printf("ERROR reloading scripts: %v\n", err)
				}
			}
		case err, ok := <-watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("Watcher error: %v\n", err)
		}
	}
}

func LoadScripts() error {
	mu.Lock()
	defer mu.Unlock()

	scripts = make(map[string]*redis.Script)
	permissions = make(map[string]string)

	// Walk scripts directory
	root := "scripts"
	// Handle test environment where path might differ
	if _, err := os.Stat(root); os.IsNotExist(err) {
		// Try stepping up (mostly for tests)
		if _, err := os.Stat("../" + root); err == nil {
			root = "../" + root
		} else if _, err := os.Stat("../../" + root); err == nil {
			root = "../../" + root
		} else {
			// Create scripts folder if not found
			if err := os.Mkdir(root, 0755); err != nil {
				return fmt.Errorf("failed to create scripts dir: %w", err)
			}
			fmt.Println("Created missing scripts directory")
		}
	}

	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("failed to read scripts dir: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".lua") {
			name := strings.TrimSuffix(entry.Name(), ".lua")
			path := filepath.Join(root, entry.Name())
			
			content, err := os.ReadFile(path)
			if err != nil {
				return fmt.Errorf("failed to read script %s: %w", name, err)
			}
			
			// Parse Metadata
			// Look for line starting with "-- ROLE:"
			role := "player" // Default
			lines := strings.Split(string(content), "\n")
			for _, line := range lines {
				line = strings.TrimSpace(line)
				if strings.HasPrefix(line, "-- ROLE:") {
					role = strings.TrimSpace(strings.TrimPrefix(line, "-- ROLE:"))
					break
				}
			}
			
			permissions[name] = role
			scripts[name] = redis.NewScript(string(content))
			fmt.Printf("Loaded script: %s (Role: %s)\n", name, role)
		}
	}

	return nil
}

func GetScriptRole(scriptName string) string {
	mu.RLock()
	defer mu.RUnlock()
	if role, ok := permissions[scriptName]; ok {
		return role
	}
	return "player"
}

func ExecuteScript(ctx context.Context, scriptName string, keys []string, args ...interface{}) (interface{}, error) {
	mu.RLock()
	script, ok := scripts[scriptName]
	mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("script %s not found", scriptName)
	}

	return script.Run(ctx, RDB, keys, args...).Result()
}

func SubscribeToGame(ctx context.Context, gameID string) *redis.PubSub {
	return RDB.Subscribe(ctx, "game_updates:game:"+gameID)
}
