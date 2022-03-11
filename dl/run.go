// Copyright © 2018 ilove91
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package dl

import (
	"fmt"
	"os"
	"os/exec"
	"sync"
	"path/filepath"
	"time"

	"github.com/cavaliergopher/grab/v3"
	"github.com/ilove91/91dl/m3u8"
	"github.com/schollz/progressbar/v3"
	log "github.com/sirupsen/logrus"
)

var baseURL = "http://91porn.com"

// LinksDl download by links
func LinksDl(vlinks []string) {
	var vs []*video
	for _, u := range vlinks {
		v, err := parseVideo(u)
		if err != nil {
			log.Errorf("video parse err: %v, url: %v", err, u)
			continue
		}
		vs = append(vs, v)
	}

	log.Infof("Total videos: %d", len(vs))

	for i, v := range vs {
		if _, err := os.Stat(filepath.Join(destDir, v.title+".mp4")); os.IsNotExist(err) {
			download(i, v)
		} else {
			log.Infof("Exists %3d  %s ...", i+1, v.title)
		}
	}
}

// PagesDl download by pages
// category: new hot rp long md tf mf rf top top-1 hd
func PagesDl(p1 int, p2 int, t string) {
	if p1 > p2 {
		p1 = p2
	}
	log.Infof("Download category %s from page %d to %d", t, p1, p2)
	log.Info("======================================")

	var url string
	for i := p1; i <= p2; i++ {
		switch t {
		case "new":
			url = fmt.Sprintf("%s/v.php?next=watch&page=%d", baseURL, i)
		case "lasttop":
			url = fmt.Sprintf("%s/v.php?category=top&m=-1&viewtype=basic&page=%d", baseURL, i)
		default:
			url = fmt.Sprintf("%s/v.php?category=%s&viewtype=basic&page=%d", baseURL, t, i)
		}
		log.Infof("Category %s page %d url: %v", t, i, url)
		vl := parsePage(url)
		log.Infof("Downloading category %s page %d ...", t, i)
		LinksDl(vl)
		log.Info("======================================")
	}
}

func download(i int, v *video) {
	toFile := filepath.Join(destDir, v.title+".mp4")
	log.Infof("Downloading %3d  %s ...", i+1, v.title)

	if v.mediaType == "m3u8" {
		err := m3u8.Download(v.videoSrc, v.title, destDir, 25)
		if err != nil {
			log.Error(err)
			return
		}
	}

	if v.mediaType == "mp4" {
		tmpFile := toFile + ".tmp"
		req, _ := grab.NewRequest(tmpFile, v.videoSrc)
		resp := grabClient.Do(req)

		// start UI loop
		t := time.NewTicker(1 * time.Second)
		defer t.Stop()

		bar := progressbar.DefaultBytes(resp.Size())

	Loop:
		for {
			select {
			case <-t.C:
				bar.Set64(resp.BytesComplete())
			case <-resp.Done:
				bar.Set64(resp.BytesComplete())
				if resp.BytesComplete() == resp.Size() {
					bar.Finish()
				}
				break Loop
			}
		}

		// check for errors
		if err := resp.Err(); err != nil {
			log.Errorf("Download failed: %v, v: %v", err, v.title)
			return
		}
		os.Rename(tmpFile, toFile)
	}
	wg := sync.WaitGroup{}
    wg.Add(1)
	go func (f string) {
		log.Infof("Uploading %s ...", f)
		c := "/root/work/faker115uploader/upload.sh"
		cmd := exec.Command("sh", "-c", c, f)
		cmd.Output()
		wg.Done()
		log.Infof("Upload %s success...", f)
	}(toFile)
	wg.Wait()
	log.Info("[Done] ", toFile)
}
