package image

import (
	imageTypes "miniK8s/pkg/minik8sTypes"
	"testing"
)

var TestImageURLs = []string{
	"docker.io/library/ubuntu:latest",
	"docker.io/library/ubuntu:20.04",
	// "docker.io/library/busybox:latest",
	// "docker.io/library/hello-world:latest",
	// "docker.io/library/nginx:latest",
	// "docker.io/library/redis:latest",
}

var TestImageURLs_Correct = []string{
	"ubuntu:latest",
	"ubuntu:20.04",
	// "busybox:latest",
	// "hello-world:latest",
	// "nginx:latest",
	// "redis:latest",
}

// 测试parseImageRef函数
func TestParseImageRef(t *testing.T) {
	for id, url := range TestImageURLs {
		result := parseImageRef(url)
		if result != TestImageURLs_Correct[id] {
			t.Errorf("parseImageRef error")
		}
	}
}

func TestPullImageWithPolicy(t *testing.T) {
	im := &ImageManager{}
	for _, url := range TestImageURLs {
		imageID, err := im.PullImageWithPolicy(url, imageTypes.PullAlways)
		if err != nil {
			t.Error(err)
		}
		t.Log(imageID)
		// println(imageID)
	}
}

func TestFindLocalImageIDsByImageRef(t *testing.T) {
	im := &ImageManager{}
	for _, url := range TestImageURLs {
		images, err := im.findLocalImageIDsByImageRef(url)
		if err != nil {
			t.Error(err)
		}

		if len(images) != 1 {
			t.Errorf("image not found or found more than one")
			// 输出images
			println(len(images))
		}
	}
}

// // 测试RemoveImage
// func TestRemoveImage(t *testing.T) {
// 	im := &ImageManager{}
// 	for _, url := range TestImageURLs {
// 		err := im.RemoveImage(url)
// 		if err != nil {
// 			t.Error(err)
// 		}
// 	}
// }
