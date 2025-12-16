package file_utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"go.viam.com/rdk/pointcloud"
	rutils "go.viam.com/rdk/utils"
)

var (
	CaptureDirectory = filepath.Join(rutils.ViamDotDir, "capture")
)

func CaptureDir(passID string) string {
	return filepath.Join(CaptureDirectory, passID)
}

// EnsureDirExists creates the target directory path if it does not exist
// using secure default permissions. It is safe to call multiple times.
func EnsureDirExists(dirPath string) error {
	if dirPath == "" {
		return nil
	}

	if err := os.MkdirAll(dirPath, 0o750); err != nil {
		return fmt.Errorf("failed to create directory %q: %w", dirPath, err)
	}

	return nil
}

// EnsureDir checks if a directory exists and creates it if it doesn't.
// This function provides additional validation compared to EnsureDirExists.
func EnsureDir(path string) error {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		// directory doesn't exist, create it
		return EnsureDirExists(path)
	}
	if err != nil {
		// some other error accessing the path
		return fmt.Errorf("error checking directory: %w", err)
	}
	if !info.IsDir() {
		// exists but is not a directory
		return fmt.Errorf("%s exists but is not a directory", path)
	}
	// the directory in question already exists
	return nil
}

func SaveFileToSync(b []byte, filename, passID string, t time.Time) error {
	// if passID is nil then there is no passID and we can't capture
	// this may happen b/c the code is being executed outside of the context of a production
	// robot (such as a unit test) and in such a case we don't want to save these files to disk
	if passID == "" {
		return nil
	}
	timestamp := t.Format("January_02_2006_15_04_05")
	capturePassDir := CaptureDir(passID)
	// create destination directory if it doesn't exist
	if err := EnsureDir(capturePassDir); err != nil {
		return fmt.Errorf("failed to create destination directory: %w", err)
	}

	dst := filepath.Join(capturePassDir, fmt.Sprintf("%s_%s", timestamp, filename))
	if err := os.WriteFile(dst, b, 0o600); err != nil {
		return err
	}
	return nil
}

func SaveJsonToSync(data any, filename, passID string, t time.Time) error {
	bytes, err := json.Marshal(data)
	if err != nil {
		return err
	}
	return SaveFileToSync(bytes, filename, passID, t)
}

func SavePCToSync(data pointcloud.PointCloud, filename, passID string, t time.Time) error {
	bytes, err := pointcloud.ToBytes(data)
	if err != nil {
		return err
	}
	return SaveFileToSync(bytes, filename, passID, t)
}
