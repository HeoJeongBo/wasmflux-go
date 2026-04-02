//go:build js && wasm

package wasmflux

import (
	"syscall/js"

	"github.com/heojeongbo/wasmflux-go/bridge"
	"github.com/heojeongbo/wasmflux-go/event"
	werrors "github.com/heojeongbo/wasmflux-go/errors"
	"github.com/heojeongbo/wasmflux-go/log"
)

// App is the top-level lifecycle orchestrator.
// It manages the bridge, event bus, logger, and module lifecycle.
type App struct {
	bridge   *bridge.Bridge
	bus      *event.Bus
	logger   *log.Logger
	logLevel log.Level
	modules  []Module
	registry *Registry
	done     chan struct{}
}

// New creates an App with the given options.
func New(opts ...Option) *App {
	a := &App{
		logLevel: log.LevelInfo,
		registry: NewRegistry(),
		done:     make(chan struct{}),
	}
	for _, opt := range opts {
		opt(a)
	}

	if a.bridge == nil {
		a.bridge = bridge.New()
	}
	if a.bus == nil {
		a.bus = event.NewBus()
	}
	if a.logger == nil {
		a.logger = log.NewLogger(log.NewConsoleWriter(), a.logLevel)
	}

	return a
}

// Register adds a module to the app. Chainable.
func (a *App) Register(m Module) *App {
	a.modules = append(a.modules, m)
	return a
}

// Bridge returns the bridge instance.
func (a *App) Bridge() *bridge.Bridge { return a.bridge }

// Bus returns the event bus instance.
func (a *App) Bus() *event.Bus { return a.bus }

// Logger returns the logger instance.
func (a *App) Logger() *log.Logger { return a.logger }

// Get returns a registered module by name, or nil if not found.
func (a *App) Get(name string) Module {
	for _, m := range a.modules {
		if m.Name() == name {
			return m
		}
	}
	return nil
}

// ModuleInfo holds metadata about a registered module.
type ModuleInfo struct {
	Name  string
	Index int
}

// Modules returns info about all registered modules.
func (a *App) Modules() []ModuleInfo {
	infos := make([]ModuleInfo, len(a.modules))
	for i, m := range a.modules {
		infos[i] = ModuleInfo{Name: m.Name(), Index: i}
	}
	return infos
}

// Run initializes and starts all modules, then blocks until Shutdown is called.
// This is the standard Go WASM keep-alive pattern.
func (a *App) Run() {
	a.logger.Info("wasmflux starting",
		log.Int("modules", int64(len(a.modules))),
	)

	// Expose shutdown function to JS.
	a.bridge.ExposeSimple("shutdown", func(_ []js.Value) {
		a.Shutdown()
	})

	ctx := ModuleContext{
		Bridge:   a.bridge,
		Bus:      a.bus,
		Logger:   a.logger,
		registry: a.registry,
	}

	// Init phase.
	for _, m := range a.modules {
		modLogger := a.logger.With(log.String("module", m.Name()))
		modLogger.Debug("initializing")
		if err := m.Init(ctx); err != nil {
			modLogger.Error("init failed", log.Err(err))
			return
		}
	}

	// Start phase.
	for _, m := range a.modules {
		modLogger := a.logger.With(log.String("module", m.Name()))
		modLogger.Debug("starting")
		if err := m.Start(); err != nil {
			modLogger.Error("start failed", log.Err(err))
			a.stopModules(a.modules)
			return
		}
	}

	a.logger.Info("wasmflux ready")
	a.bus.Emit("app:ready", nil)

	// Block forever until shutdown.
	<-a.done
}

// Shutdown gracefully stops all modules in reverse order and unblocks Run.
func (a *App) Shutdown() {
	a.logger.Info("wasmflux shutting down")
	a.bus.Emit("app:shutdown", nil)
	a.stopModules(a.modules)
	a.bridge.Release()
	close(a.done)
}

func (a *App) stopModules(modules []Module) {
	// Stop in reverse order.
	for i := len(modules) - 1; i >= 0; i-- {
		m := modules[i]
		modLogger := a.logger.With(log.String("module", m.Name()))
		modLogger.Debug("stopping")
		if err := werrors.Recover(func() {
			if stopErr := m.Stop(); stopErr != nil {
				modLogger.Error("stop error", log.Err(stopErr))
			}
		}); err != nil {
			modLogger.Error("stop panic", log.Err(err))
		}
	}
}
