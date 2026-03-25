package assets

import (
	"embed"
	"io/fs"

	"gabe565.com/utils/must"
)

//go:embed static/dist static/dist/.vite
var static embed.FS

//go:embed altcha_gatekeeper.js
var gatekeeper []byte

//go:embed altcha.min.js
var altchaJS []byte

func Static() fs.FS {
	return must.Must2(fs.Sub(static, "static/dist"))
}

func Gatekeeper() []byte {
	return gatekeeper
}

func AltchaJS() []byte {
	return altchaJS
}
