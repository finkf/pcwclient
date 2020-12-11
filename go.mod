module github.com/finkf/pcwclient

require (
	github.com/UNO-SOFT/ulog v1.1.6
	github.com/fatih/color v1.7.0
	github.com/finkf/gofiler v0.3.0
	github.com/finkf/pcwgo v0.8.10
	github.com/mattn/go-colorable v0.1.4 // indirect
	github.com/mattn/go-isatty v0.0.10 // indirect
	github.com/spf13/cobra v1.1.1
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9 // indirect
	golang.org/x/sys v0.0.0-20200930185726-fdedc70b468f // indirect
)

replace github.com/finkf/pcwgo => ../pcwgo

go 1.13
