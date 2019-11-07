// Compile with: go build -ldflags -H=windowsgui FolderOpener.go

package main

import ("strings"
		"os"
		"os/exec"
		"fmt"
		"bytes"
		"path/filepath"
		"syscall"
		b64 "encoding/base64"
)
const (
	ATTACH_PARENT_PROCESS = ^uint32(0) // (DWORD)-1
)

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	procAttachConsole = modkernel32.NewProc("AttachConsole")
)

func AttachConsole(dwParentProcess uint32) (ok bool){
	r0, _, _ := syscall.Syscall(procAttachConsole.Addr(), 1, uintptr(dwParentProcess), 0, 0)
	ok = bool(r0 != 0)
	return
}

func run(name string, arg ...string) {
	cmd := exec.Command(name, arg...)
	var outb, errb bytes.Buffer
	cmd.Stdout = &outb
	cmd.Stderr = &errb
	cmd.Run()

	err := errb.String()
	if strings.TrimSpace(err) != "" {
		fmt.Print(err)
	}

	out := outb.String()
	if strings.TrimSpace(out) != "" {
		fmt.Print(out)
	}
}

func main() {
	AttachConsole(ATTACH_PARENT_PROCESS)

	args := os.Args
	argsCount := len(args)

	switch argsCount {
	case 1: { // Install protocol handler in registry
		run("reg", `delete`, `HKEY_CLASSES_ROOT\FolderOpener`, `/f`) // Uninstall first, otherwise it gets stuck
		run("reg", `add`, `HKEY_CLASSES_ROOT\FolderOpener`)
		run("reg", `add`, `HKEY_CLASSES_ROOT\FolderOpener`, `/v`, `URL Protocol`)
		run("reg", `add`, `HKEY_CLASSES_ROOT\FolderOpener`, `/ve`, `/d`, `URL:FolderOpener Protocol`, `/f`)
		
		run("reg", `add`, `HKEY_CLASSES_ROOT\FolderOpener\shell`)
		run("reg", `add`, `HKEY_CLASSES_ROOT\FolderOpener\shell\open`)
		run("reg", `add`, `HKEY_CLASSES_ROOT\FolderOpener\shell\open\command`)
		
		exePath, _ := filepath.Abs(args[0])
		run("reg", `add`, `HKEY_CLASSES_ROOT\FolderOpener\shell\open\command`, `/ve`, `/d`, fmt.Sprintf(`%v "%%1"`, exePath), `/f`)
	}
	case 2: { // Handle protocol
		loweredArg := strings.ToLower(args[1])

		switch loweredArg {
		case "uninstall":
			run("reg", `delete`, `HKEY_CLASSES_ROOT\FolderOpener`, `/f`)
		case "help", "--help", "-help", "-h", "--h", "/?", "?":
			fmt.Println(
`FolderOpener.exe
Version: 1.2
Created by: Yarin
FolderOpener is a protocol handler that helps open the Windows Explorer (NOT Internet Explorer) from web pages.
usages:

FolderOpener.exe [no arguments]
- Run as an administrator!
- Registers this executable as a protocol handler. THe executable's current location is registered.

FolderOpener.exe uninstall
- Run as an administrator!
- Uninstalls the FolderOpener protocol handler.

FolderOpener.exe folderopener:<folder path in base64>
- decodes the folder path and opens it in Windows Explorer.

FolderOpener.exe /?
- Displays this help dialog.`)
			fmt.Println()
			default:
				encodedPath := strings.Replace(args[1], "folderopener:", "", 1)
				decodedPathBytes, _ := b64.StdEncoding.DecodeString(encodedPath)
				decodedPath := string(decodedPathBytes)
				run("explorer.exe", decodedPath)
		}
	}
	default:
		fmt.Println("fuck you.")

	}
}