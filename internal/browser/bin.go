package browser

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/go-rod/rod/lib/launcher"
)

type resolvedBin struct {
	Path   string
	Source string
}

func resolveBrowserBin(cfg Config) (resolvedBin, error) {
	if p := strings.TrimSpace(cfg.BinPath); p != "" {
		if _, err := os.Stat(p); err != nil {
			return resolvedBin{}, fmt.Errorf("browser: configured bin_path does not exist: %q: %w", p, err)
		}
		return resolvedBin{Path: p, Source: "config"}, nil
	}

	if runtime.GOOS == "windows" {
		if r, ok := resolveWindowsKnownBinaries(); ok {
			return r, nil
		}
		if p, ok := launcher.LookPath(); ok {
			p = strings.TrimSpace(p)
			if p != "" {
				if _, err := os.Stat(p); err == nil {
					return resolvedBin{Path: p, Source: "launcher.LookPath"}, nil
				}
			}
		}
	} else {
		if p, ok := launcher.LookPath(); ok {
			p = strings.TrimSpace(p)
			if p != "" {
				if _, err := os.Stat(p); err == nil {
					return resolvedBin{Path: p, Source: "launcher.LookPath"}, nil
				}
			}
		}
	}

	return resolvedBin{}, nil
}

func resolveWindowsKnownBinaries() (resolvedBin, bool) {
	// Prefer Edge first (installed by default on most Windows machines), then Chrome.
	programFiles := strings.TrimSpace(os.Getenv("ProgramFiles"))
	programFilesX86 := strings.TrimSpace(os.Getenv("ProgramFiles(x86)"))
	localAppData := strings.TrimSpace(os.Getenv("LocalAppData"))

	candidates := []struct {
		path   string
		source string
	}{
		{filepath.Join(programFilesX86, "Microsoft", "Edge", "Application", "msedge.exe"), "ProgramFiles(x86)"},
		{filepath.Join(programFiles, "Microsoft", "Edge", "Application", "msedge.exe"), "ProgramFiles"},
		{filepath.Join(localAppData, "Microsoft", "Edge", "Application", "msedge.exe"), "LocalAppData"},

		{filepath.Join(programFilesX86, "Google", "Chrome", "Application", "chrome.exe"), "ProgramFiles(x86)"},
		{filepath.Join(programFiles, "Google", "Chrome", "Application", "chrome.exe"), "ProgramFiles"},
		{filepath.Join(localAppData, "Google", "Chrome", "Application", "chrome.exe"), "LocalAppData"},
	}

	for _, c := range candidates {
		p := strings.TrimSpace(c.path)
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return resolvedBin{Path: p, Source: "windows:" + c.source}, true
		}
	}
	return resolvedBin{}, false
}
