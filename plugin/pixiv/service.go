// Package pixiv ...
package pixiv

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/FloatTech/AnimeAPI/pixiv"
	"github.com/FloatTech/AnimeAPI/pixiv/model"
	"github.com/FloatTech/ZeroBot-Plugin/plugin/pixiv/cache"
	"github.com/FloatTech/floatbox/file"
	log "github.com/sirupsen/logrus"
	zero "github.com/wdvxdr1123/ZeroBot"
	"github.com/wdvxdr1123/ZeroBot/message"
)

var cacheFilling sync.Map

// Service 用于封装整个 Pixiv 模块的依赖与接口
type Service struct {
	DB  *cache.DB
	API *pixiv.PixivAPI
	// 内部任务锁：限制每个人同一时间只能执行一个请求
	taskMu sync.Mutex
	tasks  map[int64]*taskState

	// 并发控制
	DownloadWorkers int
	SendWorkers     int
}

type taskState struct {
	Running bool
}

const pixivTempDir = "data/pixiv/temp"

// NewService ...
func NewService(db *cache.DB, pixAPI *pixiv.PixivAPI) *Service {
	return &Service{
		DB:              db,
		API:             pixAPI,
		tasks:           make(map[int64]*taskState),
		DownloadWorkers: 4,
		SendWorkers:     2,
	}
}

// Acquire ...
func (s *Service) Acquire(userID int64) bool {
	s.taskMu.Lock()
	defer s.taskMu.Unlock()

	t, ok := s.tasks[userID]
	if ok && t.Running {
		return false
	}

	if !ok {
		t = &taskState{}
		s.tasks[userID] = t
	}
	t.Running = true
	return true
}

// Release ...
func (s *Service) Release(userID int64) {
	s.taskMu.Lock()
	defer s.taskMu.Unlock()

	if t, ok := s.tasks[userID]; ok {
		t.Running = false
	}
	delete(s.tasks, userID)
}

func removeTempImages(paths []string) {
	for _, p := range paths {
		if err := os.Remove(p); err != nil && !os.IsNotExist(err) {
			log.Warnf("删除 pixiv 临时文件失败: %s: %v", p, err)
		}
	}
}

func scheduleCleanupPixivImages(paths []string, delay time.Duration) {
	if len(paths) == 0 {
		return
	}
	local := append([]string(nil), paths...)
	go func() {
		time.Sleep(delay)
		removeTempImages(local)
	}()
}

func pixivImageExt(rawURL string) string {
	parsed, err := url.Parse(rawURL)
	if err != nil {
		return ".jpg"
	}
	ext := path.Ext(parsed.Path)
	if ext == "" {
		return ".jpg"
	}
	return ext
}

func createPixivTempImagePath(pid int64, index int, rawURL string) (string, error) {
	baseDir := filepath.Join(file.BOTPATH, pixivTempDir)
	if err := os.MkdirAll(baseDir, 0o775); err != nil {
		return "", err
	}
	pattern := fmt.Sprintf("%d-", pid)
	if index > 0 {
		pattern = fmt.Sprintf("%d-%d-", pid, index)
	}
	tmp, err := os.CreateTemp(baseDir, pattern+"*"+pixivImageExt(rawURL))
	if err != nil {
		return "", err
	}
	name := tmp.Name()
	if err = tmp.Close(); err != nil {
		return "", err
	}
	return name, nil
}

func writeTempImages(pid int64, images [][]byte, urls []string) ([]string, error) {
	paths := make([]string, 0, len(images))
	cleanup := func() {
		removeTempImages(paths)
	}

	for i, img := range images {
		if len(img) == 0 {
			continue
		}
		rawURL := ""
		if i < len(urls) {
			rawURL = urls[i]
		}
		imagePath, err := createPixivTempImagePath(pid, i+1, rawURL)
		if err != nil {
			cleanup()
			return nil, err
		}
		if err = os.WriteFile(imagePath, img, 0o644); err != nil {
			cleanup()
			return nil, err
		}
		paths = append(paths, imagePath)
	}

	return paths, nil
}

func toFileImage(path string) message.Segment {
	normalized := strings.ReplaceAll(path, "\\", "/")
	if strings.HasPrefix(normalized, "/") {
		return message.Image("file://" + normalized)
	}
	return message.Image("file:///" + normalized)
}

// SendIllusts ...
func (s *Service) SendIllusts(ctx *zero.Ctx, illusts []model.IllustCache) {
	downloadSem := make(chan struct{}, s.DownloadWorkers)
	type DLResult struct {
		Ill    model.IllustCache
		Images [][]byte
		Err    error
	}

	gid := ctx.Event.GroupID
	if gid == 0 {
		gid = -ctx.Event.UserID
	}

	results := make(chan DLResult, len(illusts))

	// 并发下载
	for _, ill := range illusts {
		ill1 := ill
		downloadSem <- struct{}{}

		go func() {
			defer func() { <-downloadSem }()

			d := DLResult{
				Ill: ill1,
			}

			if ill1.PageCount > 1 {

				type pageResult struct {
					index int
					data  []byte
					err   error
				}

				pageCh := make(chan pageResult, ill1.PageCount)
				var wg sync.WaitGroup

				pageSem := make(chan struct{}, s.DownloadWorkers)

				for i := 0; int64(i) < ill1.PageCount; i++ {
					wg.Add(1)
					pageSem <- struct{}{}

					go func(page int) {
						defer wg.Done()
						defer func() { <-pageSem }()

						u := pixiv.ModifyPageGeneric(ill1.OriginalURL, page)
						img, err := s.API.Client.FetchPixivImage(ill1, u)
						if err != nil {
							pageCh <- pageResult{
								index: page,
								err:   err,
							}
							return
						}
						pageCh <- pageResult{
							index: page,
							data:  img,
							err:   err,
						}
					}(i)
				}

				// 关闭 channel
				go func() {
					wg.Wait()
					close(pageCh)
				}()

				// 预分配，保证顺序
				d.Images = make([][]byte, ill1.PageCount)

				for r := range pageCh {
					if r.err != nil && d.Err == nil {
						d.Err = r.err
					}
					if len(r.data) == 0 {
						continue
					}
					d.Images[r.index] = r.data
				}

			} else {
				img, err := s.API.Client.FetchPixivImage(ill1, ill1.OriginalURL)
				if err == nil {
					d.Images = append(d.Images, img)
				} else {
					d.Err = err
				}
			}

			log.Print("下载图片完成：", ill1.PID)
			results <- d
		}()

	}

	// 发送（顺序）
	for range illusts {
		res := <-results

		if res.Err != nil {
			if httpErr, ok := errors.AsType[*pixiv.HTTPStatusError](res.Err); ok && httpErr.StatusCode == http.StatusNotFound {
				if err := s.DB.DeleteIllustByPID(res.Ill.PID); err != nil {
					ctx.SendChain(message.Text("清理已失效图片失败: ", err))
				} else {
					ctx.SendChain(message.Text("图片已被删除，已移除缓存，PID: ", res.Ill.PID))
				}
			} else {
				ctx.SendChain(message.Text("下载失败: ", res.Err))
			}

			continue
		}

		var msg message.Message
		msg = append(msg, message.Text(
			"PID:", res.Ill.PID,
			"\n标题:", res.Ill.Title,
			"\n画师:", res.Ill.AuthorName,
			"\ntag:", res.Ill.Tags,
			"\n收藏数:", res.Ill.Bookmarks,
			"\n浏览数:", res.Ill.TotalView,
			"\n发布时间:", res.Ill.CreateDate,
		))

		imageURLs := make([]string, len(res.Images))
		for i := range imageURLs {
			if i == 0 {
				imageURLs[i] = res.Ill.OriginalURL
				continue
			}
			imageURLs[i] = pixiv.ModifyPageGeneric(res.Ill.OriginalURL, i)
		}

		tempPaths, err := writeTempImages(res.Ill.PID, res.Images, imageURLs)
		if err != nil {
			ctx.SendChain(message.Text("暂存图片失败: ", err))
			continue
		}
		for _, tpath := range tempPaths {
			msg = append(msg, toFileImage(tpath))
		}

		ctx.Send(msg)
		scheduleCleanupPixivImages(tempPaths, 15*time.Second)

		s.DB.Create(&model.SentImage{
			GroupID: gid,
			PID:     res.Ill.PID,
		})
	}
}

// BackgroundCacheFiller ...
func (s *Service) BackgroundCacheFiller(keyword string, minCache int, r18Req bool, fetchCount int, gid int64) {
	if _, loaded := cacheFilling.LoadOrStore(keyword, struct{}{}); loaded {
		log.Print("已有后台任务在补缓存: ", keyword)
		return
	}

	go func() {
		defer cacheFilling.Delete(keyword)

		count, err := s.DB.CountIllustsSmart(gid, keyword, r18Req)
		if err != nil {
			log.Print("查询数据库发生错误: ", err)
			return
		}

		if count >= int64(minCache) {
			log.Print("缓存足够，无需补充: ", keyword)
			return
		}

		log.Printf("后台补充关键词 %s, 数量 %d\n", keyword, fetchCount)

		sendedcache, err := s.DB.GetSentPictureIDs(gid)
		if err != nil {
			log.Print("后台补充缓存失败: ", err)
			return
		}
		s1, err := s.DB.GetIllustIDsByKeyword(keyword)
		if err != nil {
			log.Print("后台补充缓存失败: ", err)
			return
		}
		sendedcache = append(sendedcache, s1...)
		newIllusts, err := s.API.FetchPixivIllusts(keyword, r18Req, fetchCount, sendedcache)
		if err != nil {
			log.Print("后台补充缓存失败: ", err)
			return
		}

		if len(newIllusts) == 0 {
			log.Print("后台补充缓存：没有新图")
			return
		}

		for _, illust := range newIllusts {
			s.DB.Create(&illust)
			log.Print("后台补充缓存：", illust.PID)
		}

		log.Printf("后台成功补充 %d 张图片到关键词 %s 缓存\n", len(newIllusts), keyword)
	}()
}
