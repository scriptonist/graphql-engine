package seed

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"
)

// CreateSeedOptions has the list of options required
// to create a seed file
type CreateSeedOptions struct {
	UserProvidedSeedName string
	// DirectoryPath in which seed file should be created
	DirectoryPath string
}

// CreateSeed creates a .sql file according to the arguments
// provided and opens it for writing
func CreateSeed(opts CreateSeedOptions) (*string, error) {
	const fileExtension = "sql"

	timestamp := strconv.FormatInt(time.Now().UnixNano()/int64(time.Millisecond), 10)
	// filename will be in format <timestamp>_<userProvidedSeedName>.sql
	filenameWithTimeStamp := fmt.Sprintf("%s_%s.%s", timestamp, opts.UserProvidedSeedName, fileExtension)
	fullFilePath := filepath.Join(opts.DirectoryPath, filenameWithTimeStamp)

	// Create file
	file, err := os.Create(fullFilePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	return &fullFilePath, nil
}
