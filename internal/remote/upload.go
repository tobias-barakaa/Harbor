package remote

import (
	"io"
	"os"
	"path"
	"path/filepath"

	"github.com/pkg/sftp"
)

// UploadFile copies one local file to remotePath, creating the
// destination directory on the server first if it doesn't exist.
func (c *Client) UploadFile(localPath, remotePath string) error {
	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	if err := sftpClient.MkdirAll(path.Dir(remotePath)); err != nil {
		return err
	}

	local, err := os.Open(localPath)
	if err != nil {
		return err
	}
	defer local.Close()

	remote, err := sftpClient.Create(remotePath)
	if err != nil {
		return err
	}
	defer remote.Close()

	_, err = io.Copy(remote, local)
	return err
}

// UploadDir recursively copies a local directory tree to remoteRoot,
// preserving relative structure. This is what lets Harbor upload an
// already-extracted workspace directly, not just a single zip file.
func (c *Client) UploadDir(localRoot, remoteRoot string) error {
	sftpClient, err := sftp.NewClient(c.conn)
	if err != nil {
		return err
	}
	defer sftpClient.Close()

	return filepath.Walk(localRoot, func(localPath string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(localRoot, localPath)
		if err != nil {
			return err
		}
		remotePath := path.Join(remoteRoot, filepath.ToSlash(rel))

		if info.IsDir() {
			return sftpClient.MkdirAll(remotePath)
		}
		if err := sftpClient.MkdirAll(path.Dir(remotePath)); err != nil {
			return err
		}

		local, err := os.Open(localPath)
		if err != nil {
			return err
		}
		defer local.Close()

		remote, err := sftpClient.Create(remotePath)
		if err != nil {
			return err
		}
		defer remote.Close()

		_, err = io.Copy(remote, local)
		return err
	})
}
