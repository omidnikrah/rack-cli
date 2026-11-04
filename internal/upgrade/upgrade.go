package upgrade

import (
	"archive/tar"
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

const (
	repo = "omidnikrah/rack-cli"
	bin  = "rack-cli"
)

var client = &http.Client{Timeout: 2 * time.Minute}

type Result struct {
	From, To string
	Path     string
	Upgraded bool
}

func Run(current string) (Result, error) {
	if current == "dev" {
		return Result{}, errors.New("this is a development build; install a release to use --upgrade")
	}

	latest, err := latestTag()
	if err != nil {
		return Result{}, fmt.Errorf("checking latest release: %w", err)
	}

	res := Result{From: current, To: strings.TrimPrefix(latest, "v")}
	if strings.TrimPrefix(current, "v") == res.To {
		return res, nil
	}

	exe, err := os.Executable()
	if err != nil {
		return res, err
	}
	if exe, err = filepath.EvalSymlinks(exe); err != nil {
		return res, err
	}
	res.Path = exe

	archive := fmt.Sprintf("%s_%s_%s.tar.gz", bin, runtime.GOOS, runtime.GOARCH)
	base := fmt.Sprintf("https://github.com/%s/releases/download/%s/", repo, latest)

	data, err := fetch(base + archive)
	if err != nil {
		return res, err
	}
	sums, err := fetch(base + "checksums.txt")
	if err != nil {
		return res, err
	}
	if err := verify(data, sums, archive); err != nil {
		return res, err
	}

	binary, err := extract(data)
	if err != nil {
		return res, err
	}

	if err := replace(exe, binary); err != nil {
		return res, err
	}

	res.Upgraded = true
	return res, nil
}

func latestTag() (string, error) {
	body, err := fetch("https://api.github.com/repos/" + repo + "/releases/latest")
	if err != nil {
		return "", err
	}

	var rel struct {
		TagName string `json:"tag_name"`
	}
	if err := json.Unmarshal(body, &rel); err != nil {
		return "", err
	}
	if rel.TagName == "" {
		return "", errors.New("no release found")
	}
	return rel.TagName, nil
}

func fetch(url string) ([]byte, error) {
	resp, err := client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("GET %s: %s", url, resp.Status)
	}
	return io.ReadAll(resp.Body)
}

func verify(data, sums []byte, archive string) error {
	sc := bufio.NewScanner(bytes.NewReader(sums))
	for sc.Scan() {
		fields := strings.Fields(sc.Text())
		if len(fields) != 2 || fields[1] != archive {
			continue
		}
		sum := sha256.Sum256(data)
		if hex.EncodeToString(sum[:]) != fields[0] {
			return fmt.Errorf("checksum mismatch for %s", archive)
		}
		return nil
	}
	return fmt.Errorf("no checksum listed for %s", archive)
}

func extract(data []byte) ([]byte, error) {
	gz, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	defer gz.Close()

	tr := tar.NewReader(gz)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			return nil, fmt.Errorf("%s not found in archive", bin)
		}
		if err != nil {
			return nil, err
		}
		if hdr.Typeflag == tar.TypeReg && filepath.Base(hdr.Name) == bin {
			return io.ReadAll(tr)
		}
	}
}

func replace(exe string, binary []byte) error {
	tmp, err := os.CreateTemp(filepath.Dir(exe), "."+bin+"-upgrade-*")
	if err != nil {
		if errors.Is(err, os.ErrPermission) {
			return fmt.Errorf("no permission to write to %s; rerun with sudo", filepath.Dir(exe))
		}
		return err
	}
	defer os.Remove(tmp.Name())

	if _, err := tmp.Write(binary); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Chmod(0o755); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), exe)
}
