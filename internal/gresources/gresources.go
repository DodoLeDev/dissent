package gresources

import (
	"log/slog"

	"github.com/diamondburned/gotk4/pkg/gio/v2"
	"github.com/diamondburned/gotk4/pkg/gtk/v4"
	"github.com/diamondburned/gotk4/pkg/glib/v2"
)

var Path string = "/app/share/glib-2.0/dissent/resources.gresource"
var resourcePath string = "/so/libdb/dissent"

// UI access facilities
func WindowUI() glib.Objector { return uiBuild("window.ui", "content") }


func Load() bool {
	var resource *gio.Resource
	var err error

	resource, err = gio.ResourceLoad(Path)
	if err != nil {
		slog.Error("Failed to open resource file at", "path", Path)
		return false
	}
	gio.ResourcesRegister(resource)
	return true
}

func uiBuild(filename string, rootID string) glib.Objector {
	var builder *gtk.Builder = gtk.NewBuilderFromResource(resourcePath + "/" + filename)
	return builder.GetObject(rootID).Cast()
}
