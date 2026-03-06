package ctx

// GlobalCtx holds flags shared across all commands.
type GlobalCtx struct {
	Output        string // "auto" | "table" | "json" | "csv"
	Profile       string
	NoColor       bool
	Quiet         bool
	Verbose       bool
	NoInteractive bool
}
