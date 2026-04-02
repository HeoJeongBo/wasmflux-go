// wasmflux-go JS glue layer
// Provides helpers for the Go WASM bridge.

(function () {
    "use strict";

    const wasmflux = {
        _handlers: {},

        // Register a JS handler callable from Go.
        register(name, fn) {
            this._handlers[name] = fn;
        },

        // Call a registered handler.
        call(name, ...args) {
            const fn = this._handlers[name];
            if (!fn) {
                console.warn(`[wasmflux] handler not found: ${name}`);
                return undefined;
            }
            return fn(...args);
        },

        // Log to the HTML log panel.
        appendLog(level, msg) {
            const el = document.getElementById("log");
            if (!el) return;
            const line = document.createElement("div");
            line.className = `log-${level}`;
            const ts = new Date().toISOString().slice(11, 23);
            line.textContent = `[${ts}] [${level.toUpperCase()}] ${msg}`;
            el.appendChild(line);
            el.scrollTop = el.scrollHeight;
        },

        // Batch transfer: convert JS array to Float64Array for efficient Go transfer.
        toFloat64Array(arr) {
            return new Float64Array(arr);
        },

        // Performance helpers.
        perf: {
            mark(name) {
                performance.mark(name);
            },
            measure(name, start, end) {
                performance.measure(name, start, end);
                const entries = performance.getEntriesByName(name, "measure");
                return entries.length > 0 ? entries[entries.length - 1].duration : 0;
            },
        },
    };

    window.wasmflux = wasmflux;
})();
