package config

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// ConfigWatcher watches for configuration file changes and reloads automatically
type ConfigWatcher struct {
	configFile  string
	config      *Config
	configMutex sync.RWMutex
	watcher     *fsnotify.Watcher
	callbacks   []func(*Config)
	ctx         context.Context
	cancel      context.CancelFunc
}

// NewConfigWatcher creates a new configuration watcher
func NewConfigWatcher(configFile string, initialConfig *Config) (*ConfigWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	cw := &ConfigWatcher{
		configFile: configFile,
		config:     initialConfig,
		watcher:    watcher,
		ctx:        ctx,
		cancel:     cancel,
		callbacks:  make([]func(*Config), 0),
	}

	return cw, nil
}

// Start begins watching for configuration changes
func (cw *ConfigWatcher) Start() error {
	if err := cw.watcher.Add(cw.configFile); err != nil {
		return err
	}

	go cw.watch()
	return nil
}

// Stop stops watching for configuration changes
func (cw *ConfigWatcher) Stop() {
	cw.cancel()
	cw.watcher.Close()
}

// GetConfig returns the current configuration (thread-safe)
func (cw *ConfigWatcher) GetConfig() *Config {
	cw.configMutex.RLock()
	defer cw.configMutex.RUnlock()
	return cw.config
}

// OnConfigChange registers a callback to be called when configuration changes
func (cw *ConfigWatcher) OnConfigChange(callback func(*Config)) {
	cw.callbacks = append(cw.callbacks, callback)
}

func (cw *ConfigWatcher) watch() {
	// Use a timer to debounce multiple rapid file changes
	timer := time.NewTimer(0)
	timer.Stop()

	for {
		select {
		case <-cw.ctx.Done():
			return
		case event, ok := <-cw.watcher.Events:
			if !ok {
				return
			}

			// Only handle write and create events
			if event.Op&fsnotify.Write == fsnotify.Write || event.Op&fsnotify.Create == fsnotify.Create {
				// Reset the timer to debounce rapid changes
				timer.Reset(100 * time.Millisecond)
			}

		case <-timer.C:
			cw.reloadConfig()

		case err, ok := <-cw.watcher.Errors:
			if !ok {
				return
			}
			log.Printf("Configuration watcher error: %v", err)
		}
	}
}

func (cw *ConfigWatcher) reloadConfig() {
	newConfig, err := LoadConfig(cw.configFile)
	if err != nil {
		log.Printf("Failed to reload configuration: %v", err)
		return
	}

	cw.configMutex.Lock()
	cw.config = newConfig
	cw.configMutex.Unlock()

	log.Printf("Configuration reloaded from %s", cw.configFile)

	// Notify all callbacks
	for _, callback := range cw.callbacks {
		go callback(newConfig)
	}
}

// ConfigManager manages configuration with hot reload support
type ConfigManager struct {
	watcher *ConfigWatcher
	config  *Config
}

// NewConfigManager creates a new configuration manager with hot reload
func NewConfigManager(configFile string) (*ConfigManager, error) {
	config, err := LoadConfig(configFile)
	if err != nil {
		return nil, err
	}

	watcher, err := NewConfigWatcher(configFile, config)
	if err != nil {
		return nil, err
	}

	cm := &ConfigManager{
		watcher: watcher,
		config:  config,
	}

	// Update config reference when configuration changes
	watcher.OnConfigChange(func(newConfig *Config) {
		cm.config = newConfig
	})

	return cm, nil
}

// Start begins configuration hot reload
func (cm *ConfigManager) Start() error {
	return cm.watcher.Start()
}

// Stop stops configuration hot reload
func (cm *ConfigManager) Stop() {
	cm.watcher.Stop()
}

// GetConfig returns the current configuration
func (cm *ConfigManager) GetConfig() *Config {
	return cm.watcher.GetConfig()
}

// OnConfigChange registers a callback for configuration changes
func (cm *ConfigManager) OnConfigChange(callback func(*Config)) {
	cm.watcher.OnConfigChange(callback)
}
