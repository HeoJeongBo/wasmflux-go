//go:build js && wasm

package jsutil

import (
	"fmt"
	"syscall/js"
)

// QuerySelector returns the first element matching the CSS selector.
func QuerySelector(selector string) (js.Value, error) {
	el := Document.Call("querySelector", selector)
	if el.IsNull() {
		return js.Undefined(), fmt.Errorf("element not found: %s", selector)
	}
	return el, nil
}

// QuerySelectorAll returns all elements matching the CSS selector.
func QuerySelectorAll(selector string) []js.Value {
	nodeList := Document.Call("querySelectorAll", selector)
	n := nodeList.Length()
	result := make([]js.Value, n)
	for i := 0; i < n; i++ {
		result[i] = nodeList.Index(i)
	}
	return result
}

// GetElementByID returns the element with the given ID.
func GetElementByID(id string) (js.Value, error) {
	el := Document.Call("getElementById", id)
	if el.IsNull() {
		return js.Undefined(), fmt.Errorf("element not found: #%s", id)
	}
	return el, nil
}

// CreateElement creates a new DOM element with the given tag.
func CreateElement(tag string) js.Value {
	return Document.Call("createElement", tag)
}

// SetText sets the textContent of a DOM element.
func SetText(el js.Value, text string) {
	el.Set("textContent", text)
}

// SetHTML sets the innerHTML of a DOM element.
func SetHTML(el js.Value, html string) {
	el.Set("innerHTML", html)
}

// AddClass adds a CSS class to a DOM element.
func AddClass(el js.Value, class string) {
	el.Get("classList").Call("add", class)
}

// RemoveClass removes a CSS class from a DOM element.
func RemoveClass(el js.Value, class string) {
	el.Get("classList").Call("remove", class)
}

// ToggleClass toggles a CSS class on a DOM element. Returns the new state.
func ToggleClass(el js.Value, class string) bool {
	return el.Get("classList").Call("toggle", class).Bool()
}

// HasClass reports whether a DOM element has the given CSS class.
func HasClass(el js.Value, class string) bool {
	return el.Get("classList").Call("contains", class).Bool()
}

// SetAttr sets an attribute on a DOM element.
func SetAttr(el js.Value, name, value string) {
	el.Call("setAttribute", name, value)
}

// GetAttr returns the value of an attribute on a DOM element.
func GetAttr(el js.Value, name string) string {
	return el.Call("getAttribute", name).String()
}

// RemoveAttr removes an attribute from a DOM element.
func RemoveAttr(el js.Value, name string) {
	el.Call("removeAttribute", name)
}

// SetStyle sets a CSS style property on a DOM element.
func SetStyle(el js.Value, prop, value string) {
	el.Get("style").Set(prop, value)
}

// AppendChild appends a child element to a parent element.
func AppendChild(parent, child js.Value) {
	parent.Call("appendChild", child)
}

// RemoveElement removes an element from the DOM.
func RemoveElement(el js.Value) {
	el.Call("remove")
}

// Listen adds an event listener to a DOM element and returns a cleanup function.
// The cleanup function removes the listener and releases the js.Func.
func Listen(el js.Value, eventType string, handler func(event js.Value)) func() {
	fn := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			handler(args[0])
		}
		return nil
	})
	el.Call("addEventListener", eventType, fn)
	return func() {
		el.Call("removeEventListener", eventType, fn)
		fn.Release()
	}
}

// ListenOnce adds a one-shot event listener. Automatically removed after first call.
func ListenOnce(el js.Value, eventType string, handler func(event js.Value)) js.Func {
	opts := js.Global().Get("Object").New()
	opts.Set("once", true)
	fn := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) > 0 {
			handler(args[0])
		}
		return nil
	})
	el.Call("addEventListener", eventType, fn, opts)
	return fn
}

// DelegateListener adds a delegated event listener to a parent element.
// Only triggers when the event target matches the CSS selector.
func DelegateListener(parent js.Value, selector, eventType string, handler func(event js.Value)) func() {
	fn := js.FuncOf(func(_ js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		evt := args[0]
		target := evt.Get("target")
		if target.IsNull() || target.IsUndefined() {
			return nil
		}
		el := target.Call("closest", selector)
		if !el.IsNull() && !el.IsUndefined() {
			handler(evt)
		}
		return nil
	})
	parent.Call("addEventListener", eventType, fn)
	return func() {
		parent.Call("removeEventListener", eventType, fn)
		fn.Release()
	}
}
