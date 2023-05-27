package zip

import (
	"os"

	"github.com/mholt/archiver"
)

// 压缩文件或文件夹为ZIP格式
// 压缩文件夹 /path/to/your/folder， 建议不要写作 /path/to/your/folder/，否则会出问题
func CompressToZip(source, target string) error {
	return archiver.DefaultZip.Archive([]string{source}, target)
}

func DecompressZip(source, target string) error {
	return archiver.DefaultZip.Unarchive(source, target)
}

func ComvertZipToBytes(source string) ([]byte, error) {
	// TODO
	zipBytes, err := os.ReadFile(source)

	if err != nil {
		return nil, err
	}

	return zipBytes, nil
}

func ConvertBytesToZip(source []byte, target string) error {
	// TODO
	err := os.WriteFile(target, source, os.ModePerm)

	if err != nil {
		return err
	}

	return nil
}

func CompressToTar(source, target string) error {
	return archiver.DefaultTar.Archive([]string{source}, target)
}

func DecompressTar(source, target string) error {
	return archiver.DefaultTar.Unarchive(source, target)
}

func ComvertTarToBytes(source string) ([]byte, error) {
	// TODO
	tarBytes, err := os.ReadFile(source)

	if err != nil {
		return nil, err
	}

	return tarBytes, nil
}

func ConvertBytesToTar(source []byte, target string) error {
	// TODO
	err := os.WriteFile(target, source, os.ModePerm)

	if err != nil {
		return err
	}

	return nil
}
