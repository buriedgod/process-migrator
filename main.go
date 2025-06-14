package main

import (
	"archive/tar"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
)

func createArchive(dir, archive string) error {
	f, err := os.Create(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	tw := tar.NewWriter(f)
	defer tw.Close()

	return filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}
		header, err := tar.FileInfoHeader(info, "")
		if err != nil {
			return err
		}
		header.Name = rel
		if err := tw.WriteHeader(header); err != nil {
			return err
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()
		if _, err := io.Copy(tw, f); err != nil {
			return err
		}
		return nil
	})
}

func extractArchive(archive, dir string) error {
	f, err := os.Open(archive)
	if err != nil {
		return err
	}
	defer f.Close()

	tr := tar.NewReader(f)
	for {
		hdr, err := tr.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		path := filepath.Join(dir, hdr.Name)
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			return err
		}
		out, err := os.Create(path)
		if err != nil {
			return err
		}
		if _, err := io.Copy(out, tr); err != nil {
			out.Close()
			return err
		}
		out.Close()
	}
	return nil
}

func dumpProcess(pid int, archive string) error {
	dir, err := os.MkdirTemp("", "criu-dump")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	cmd := exec.Command("criu", "dump", "-t", fmt.Sprint(pid), "-D", dir, "--shell-job", "--leave-running")
	if err := cmd.Run(); err != nil {
		return err
	}
	return createArchive(dir, archive)
}

func restoreProcess(archive string) error {
	dir, err := os.MkdirTemp("", "criu-restore")
	if err != nil {
		return err
	}
	defer os.RemoveAll(dir)

	if err := extractArchive(archive, dir); err != nil {
		return err
	}
	cmd := exec.Command("criu", "restore", "-D", dir, "--shell-job")
	return cmd.Run()
}

func main() {
	pid := flag.Int("pid", 0, "process id to dump")
	archive := flag.String("archive", "checkpoint.tar", "archive file")
	restore := flag.Bool("restore", false, "restore from archive")
	flag.Parse()

	if *restore {
		if err := restoreProcess(*archive); err != nil {
			fmt.Println("restore failed:", err)
		}
		return
	}

	if *pid == 0 {
		fmt.Println("pid required")
		return
	}

	if err := dumpProcess(*pid, *archive); err != nil {
		fmt.Println("dump failed:", err)
	}
}
