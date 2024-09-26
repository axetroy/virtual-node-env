package extractor

import (
	"archive/tar"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/ulikunitz/xz"
)

func extractTarXz(tarXzFilePath, distFolder string) error {
	// 打开 .tar.xz 文件
	file, err := os.Open(tarXzFilePath)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %w", tarXzFilePath, err)
	}
	defer file.Close()

	// 创建 xz.Reader 来解压缩 .xz 数据
	xzReader, err := xz.NewReader(file)
	if err != nil {
		return fmt.Errorf("failed to create xz reader: %w", err)
	}

	// 创建 tar.Reader 来读取解压缩后的 .tar 数据
	tarReader := tar.NewReader(xzReader)

	// 迭代 .tar 文件中的每个文件和目录
	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break // 读取完毕
		}
		if err != nil {
			return fmt.Errorf("failed to read tar header: %w", err)
		}

		// 目标文件路径
		target := filepath.Join(distFolder, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			// 如果是目录，创建目录
			if err := os.MkdirAll(target, os.FileMode(header.Mode)); err != nil {
				return errors.WithStack(err)
			}
		case tar.TypeReg:
			// 如果是普通文件，创建并写入文件内容
			file, err := os.OpenFile(target, os.O_CREATE|os.O_RDWR, os.FileMode(header.Mode))
			if err != nil {
				return errors.WithStack(err)
			}
			defer file.Close()

			if _, err := io.Copy(file, tarReader); err != nil {
				return errors.WithStack(err)
			}
		case tar.TypeSymlink:
			// 如果是软链接，创建软链接
			if err := os.Symlink(header.Linkname, target); err != nil {
				return errors.WithStack(err)
			}
		case tar.TypeLink:
			// 如果是硬链接，创建硬链接
			if err := os.Link(filepath.Join(distFolder, header.Linkname), target); err != nil {
				return errors.WithStack(err)
			}
		default:
			return fmt.Errorf("Unsupported type: %v in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}
