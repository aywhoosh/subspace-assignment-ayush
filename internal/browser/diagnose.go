package browser

import "fmt"

type Diagnosis struct {
	ResolvedBinPath   string
	ResolvedBinSource string
	Leakless          bool
	AllowDownload     bool
}

func Diagnose(cfg Config) (Diagnosis, error) {
	r, err := resolveBrowserBin(cfg)
	if err != nil {
		return Diagnosis{}, err
	}
	return Diagnosis{
		ResolvedBinPath:   r.Path,
		ResolvedBinSource: r.Source,
		Leakless:          cfg.Leakless,
		AllowDownload:     cfg.AllowDownload,
	}, nil
}

func (d Diagnosis) String() string {
	path := d.ResolvedBinPath
	if path == "" {
		path = "(none)"
	}
	src := d.ResolvedBinSource
	if src == "" {
		src = "(n/a)"
	}
	return fmt.Sprintf("browser: bin=%s (source=%s) leakless=%v allow_download=%v", path, src, d.Leakless, d.AllowDownload)
}
