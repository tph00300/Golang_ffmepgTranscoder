package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/new_transcoder/customlogger"
	"github.com/new_transcoder/database"
	"github.com/new_transcoder/ftp"
	"github.com/new_transcoder/response"
	"github.com/new_transcoder/transcoder/ffmpeg"
	"github.com/new_transcoder/transcoder/preset"
	"github.com/new_transcoder/transcoder/thumbnail"
)

var Progress1 response.PresetProgress
var JsonRt string

func main() {
	// fiber는 go언어 기반 웹프레임워크
	// fiber 생성
	/*
		원래는 app := fiber.New() 로 사용하고 있었는데,
		panic이 발생했을 때에는 critical log를 날리기 위해서 custom ErrorHandler를 추가 해주는 코드를 추가해줌
	*/

	app := fiber.New(fiber.Config{
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			return c.Status(500).SendString(err.Error())
		},
	})
	app.Use(logger.New())
	app.Use(recover.New())
	customlogger.Init()
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, ATLAS Transcoder 👋!")
	})
	app.Get("/ping", func(c *fiber.Ctx) error {
		return c.SendString("OK")
	})
	app.Post("/tr1/job", func(c *fiber.Ctx) error {

		result := string(c.Body())
		f2, _ := os.Create("transcode.json")
		defer f2.Close()
		n, _ := f2.WriteString(result)
		fmt.Println("file size : ", n)
		// fmt.Println(result)

		return c.SendString("Job Received\n" + result)
		// c.Redirect("/tr1/transcode")
	})
	app.Get("/tr1/2020-10-15/transcode", func(c *fiber.Ctx) error {
		// Step 1. make Transcode.json
		jobTemplate := "transcode.json"
		data := preset.Loadjson(jobTemplate)
		DownPath := "/home" + data.HomePath + "/" + data.FileName
		// Step 2. Download OrgVideo
		err := ftp.CurlDownFtp(data.UploadIP, data.FileName, strings.ReplaceAll(data.FilePath, "/home"+data.HomePath, ""), DownPath)
		if err != nil {
			customlogger.CritLogger.Println(err.Error())
			return getError(err)
		}
		// Step 3. Upload OrgVideo
		JsonRt = "OrgUploading"
		for i := 0; i < len(data.ResultIP); i++ {
			UploadPath := strings.ReplaceAll(data.ResultIP[i].RemainingPath, "/home"+data.HomePath, "")
			if UploadPath != "" {
				err = ftp.CurlVideoUpFtp(DownPath, UploadPath, data.ResultIP[i])
				if err != nil {
					customlogger.CritLogger.Println(err)
					return getError(err)
				}
			}
		}
		// Create : Progressing
		JsonRt = "Transcoding"
		Progress1.PP = make(map[string]float64)
		for i := 0; i < len(data.Presets); i++ {
			Progress1.PP[data.Presets[i].PresetName] = 0.0
		}
		dbms := database.NewDBMSWithHost("root", "Qwpo1209", "211.110.226.42", "SKB_TRANSCODING")

		res, _ := json.Marshal(Progress1.TranscodingstartResponse())
		sql := `UPDATE
					DR_JOB
				SET
					transcoding_json = JSON_MERGE_PATCH(transcoding_json,'` + string(res) + `')
				WHERE 
					job_id = ` + strconv.Itoa(data.JobID)
		err = dbms.MySQLExec(sql)
		// Step 4. Transcoding JSON
		res, _ = json.Marshal(Progress1.TranscodingResponse())
		sql = `UPDATE
					DR_JOB
				SET
					transcoding_json = JSON_MERGE_PATCH(transcoding_json,'` + string(res) + `')
				WHERE 
					job_id = ` + strconv.Itoa(data.JobID)
		err = dbms.MySQLExec(sql)
		if err != nil {
			customlogger.CritLogger.Println("Error Transcoding JSON Upload")
			return getError(err)
		}
		// Step 5. Transcoding
		if len(data.Transcodings) > 0 {
			err = transPass3()
			if err != nil {
				fmt.Println("Transcoding FAIL")
				customlogger.CritLogger.Println("Transcoding FAIL", DownPath, len(data.Transcodings), "file total")
				return getError(err)
			} else {
				fmt.Println("Transcoding complete")
				customlogger.InfoLogger.Println("Transcoding Complete", DownPath, len(data.Transcodings), "file total")
			}
			for i := 0; i < len(data.Transcodings); i++ {
				fmt.Println("this is " + data.FilePath)
				// customeName := data.CustomerName
				// transcoding result file
				splitout := strings.Split(data.Transcodings[i].Output, "/")
				resultname := splitout[len(splitout)-1]
				resultFile := "/home" + data.HomePath + "/" + resultname //result File의 위치
				// full path : /home/woori/sample2.mp4
				fmt.Println(resultFile)
				if data.Presets[i].Video["OverlayW"] != "" {
					fmt.Println("########")
					OrgWidth := data.Presets[i].Video["Width"].(string)
					OrgHeight := data.Presets[i].Video["Height"].(string)
					OverlayW := data.Presets[i].Video["OverlayW"].(string)
					OverlayH := data.Presets[i].Video["OverlayH"].(string)
					GeometryX := data.Presets[i].Video["OverlayGeometryX"].(string)
					GeometryY := data.Presets[i].Video["OverlayGeometryX"].(string)
					Transparency := data.Presets[i].Video["OverlayTransparency"].(int)
					err = DoOverlay("Figure-1_large.png", resultFile, OrgWidth, OrgHeight, OverlayW, OverlayH, data.Presets[i].Video["OverlayGravity"].(string), GeometryX, GeometryY, Transparency)
					if err != nil {
						return getError(err)
					}
					cmd := exec.Command("mv", "-f", strings.ReplaceAll(resultFile, ".mp4", "_ovr.mp4"), resultFile)
					err := cmd.Run()
					if err != nil {
						customlogger.CritLogger.Println(err)
						return getError(err)
					}
					customlogger.InfoLogger.Println("Overlay file :" + resultFile)
				}
			}
		}
		// JsonRt = "Size"
		// for i:=0; i<len(data.Transcodings); i++{
		// 	size,err := exec.Command("mediainfo",`--Inform="General"`)
		// }
		fmt.Println("12321321321#")
		fmt.Println("Thumbnail Num :", data.DefaultThumbnail)
		// Step 6. Make Thumbnail
		cmd2 := exec.Command("mkdir", "-p", "/home"+data.HomePath+"/thb")
		fmt.Println("mkdir")
		err = cmd2.Run()
		if err != nil {
			return getError(err)
		}
		fmt.Println(len(data.DefaultThumbnail))
		for i := 0; i < len(data.DefaultThumbnail); i++ {
			fmt.Println("Default Thumbnail")
			if data.DefaultThumbnail[i].ThumbnailSource == "original" {
				fmt.Println("org default Thumbnail")
				OrgPath := "/home" + data.HomePath + "/" + data.FileName
				err := thumbnail.OrgMakeDefaultThumbnail(data.DefaultThumbnail[i], OrgPath)
				if err != nil {
					return getError(err)
				}
			} else if data.DefaultThumbnail[i].ThumbnailSource != "" {
				for j := 0; j < len(data.Transcodings); j++ {
					if data.Transcodings[j].PresetName == data.DefaultThumbnail[i].ThumbnailSource {
						splitout := strings.Split(data.Transcodings[j].Output, "/")
						resultname := splitout[len(splitout)-1]
						resultFile := "/home" + data.HomePath + "/" + resultname //result File의 위치
						err := thumbnail.PresetMakeDefaultThumbnail(data.DefaultThumbnail[i], resultFile, data.Presets[j].Prefix)
						if err != nil {
							return getError(err)
						}
						break
					}
				}
			}
		}
		for i := 0; i < len(data.GridThumbnail); i++ {
			fmt.Println("Grid Thumbnail")
			if data.GridThumbnail[i].ThumbnailSource == "original" {
				fmt.Println("org grid Thumbnail")
				OrgPath := "/home" + data.HomePath + "/" + data.FileName
				err := thumbnail.OrgMakeGridThumbnail(data.GridThumbnail[i], OrgPath)
				if err != nil {
					return getError(err)
				}
			} else if data.GridThumbnail[i].ThumbnailSource != "" {
				for j := 0; j < len(data.Transcodings); j++ {
					if data.Transcodings[j].PresetName == data.GridThumbnail[i].ThumbnailSource {
						fmt.Println("grid thumbnail", data.Transcodings[j].PresetName)
						splitout := strings.Split(data.Transcodings[j].Output, "/")
						resultname := splitout[len(splitout)-1]
						resultFile := "/home" + data.HomePath + "/" + resultname //result File의 위치
						err := thumbnail.PresetMakeGridThumbnail(data.GridThumbnail[i], resultFile, data.Presets[j].Prefix)
						if err != nil {
							return getError(err)
						}
						break
					}
				}
			}
		}
		// Step 7. Upload Transcoding Video,Thumbnail
		JsonRt = "Uploading"
		sql = `SELECT
					transcoding_json
				FROM 
					DR_JOB
				WHERE 
					job_id = ` + strconv.Itoa(data.JobID)

		jobinfo, err := dbms.MySQLMultirowQuery(sql)
		if err != nil {
			return getError(err)
		}
		res, err = json.Marshal(response.UploadingResponse((*jobinfo)[0]["transcoding_json"].(string)))
		if err != nil {
			return getError(err)
		}

		sql = `UPDATE
					DR_JOB
				SET
					transcoding_json = '` + string(res) + `'
				WHERE 
					job_id = ` + strconv.Itoa(data.JobID)
		err = dbms.MySQLExec(sql)
		if err != nil {
			return getError(err)
		}
		for i := 0; i < len(data.Transcodings); i++ {
			fmt.Println("this is " + data.FilePath)
			// customeName := data.CustomerName
			// transcoding result file
			splitout := strings.Split(data.Transcodings[i].Output, "/")
			resultname := splitout[len(splitout)-1]
			resultFile := "/home" + data.HomePath + "/" + resultname //result File의 위치
			// full path : /home/woori/sample2.mp4
			uploadPath := data.Transcodings[i].Output // /test2/sample_1080p.mp4 // 업로드할 위치
			fmt.Println(resultFile)
			fmt.Println(uploadPath)
			for j := 0; j < len(data.ResultIP); j++ {
				err = ftp.CurlVideoUpFtp(resultFile, uploadPath, data.ResultIP[j])
				if err != nil {
					return getError(err)
				}
			}
		}
		JsonRt = "Done"
		// Step 8. Upload End JSON
		sql = `SELECT
					transcoding_json
				FROM 
					DR_JOB
				WHERE 
					job_id = ` + strconv.Itoa(data.JobID)

		jobinfo, err = dbms.MySQLMultirowQuery(sql)
		res, err = json.Marshal(response.DoneResponse((*jobinfo)[0]["transcoding_json"].(string)))
		if err != nil {
			return getError(err)
		}
		sql = `UPDATE
					DR_JOB
				SET
					transcoding_json = '` + string(res) + `'
				WHERE 
					job_id = ` + strconv.Itoa(data.JobID)
		err = dbms.MySQLExec(sql)
		if err != nil {
			return getError(err)
		}
		return c.SendString("Job Complete")
	})
	app.Listen(":3333")
}
func getError(err error) error {
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	customlogger.CritLogger.Println(err.Error())
	dbms := database.NewDBMSWithHost("root", "Qwpo1209", "211.110.226.42", "SKB_TRANSCODING")
	res, _ := json.Marshal(response.FailResponse(err, JsonRt))
	sql := `UPDATE
				DR_JOB
			SET
				transcoding_json = JSON_MERGE_PATCH(transcoding_json,'` + string(res) + `'),
				status = 'FAIL'
			WHERE 
				job_id = ` + strconv.Itoa(data.JobID)
	err = dbms.MySQLExec(sql)
	if err != nil {
		customlogger.CritLogger.Println("Error Transcoding JSON Upload")
		return getError(err)
	}
	return fiber.NewError(fiber.StatusInternalServerError, "{\"err\": \""+err.Error()+"\"}")
}
func transPass3() error {

	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	// ************ transcoding file Begin ************
	var wg sync.WaitGroup
	orgfile := "/home" + data.HomePath + "/" + data.FileName
	chkPreset := make([]string, 0)
	findwidth, err := exec.Command("mediainfo", `--Inform=Video;%Width%`, orgfile).Output()
	if err != nil {
		return err
	}
	widthorg2, err := strconv.Atoi(strings.ReplaceAll(string(findwidth), "\n", ""))
	if err != nil {
		return err
	}
	findheight, err := exec.Command("mediainfo", `--Inform=Video;%Height%`, orgfile).Output()
	if err != nil {
		return err
	}
	heightorg2, err := strconv.Atoi(strings.ReplaceAll(string(findheight), "\n", ""))
	if err != nil {
		return err
	}
	// if data.Presets.Video != null {
	for i := 0; i < len(data.Presets); i++ {
		chkPreset = append(chkPreset, data.Presets[i].PresetName)
	}
	customlogger.InfoLogger.Println(data.FileName)
	wg.Add(len(data.Transcodings))
	for i := 0; i < len(data.Transcodings); i++ {
		customlogger.DebugLogger.Println("Preset : " + data.Transcodings[i].PresetName + " start")
		presetName := data.Transcodings[i].PresetName
		fmt.Println("data.Transcodings[i].Output=", data.Transcodings[i].Output)
		fmt.Println()
		splitout := strings.Split(data.Transcodings[i].Output, "/")
		outputname := splitout[len(splitout)-1]
		customerName := strings.Split(data.FilePath, "/")[2]
		output := "/home/" + customerName + "/" + outputname
		var func_error error
		go func() {
			defer wg.Done()
			for i := 0; i < len(chkPreset); i++ {

				if presetName == chkPreset[i] {

					var videoOpt map[string]interface{}
					videoOpt = (data.Presets[i].Video)

					var audioOpt map[string]interface{}
					audioOpt = (data.Presets[i].Audio)

					var opts ffmpeg.Options

					width := data.Presets[i].Video["Width"]
					height := data.Presets[i].Video["Height"]

					resolution := width.(string) + ":" + height.(string)

					// imgLocation := data.Presets[i].Video["OverlayXY"].(string)
					// imgSize := data.Presets[i].Video["OverlaySize"]

					// imgOverlay := imgLocation.(string) + imgSize.(string)

					if videoOpt["Codec"] != nil {
						optValue := videoOpt["Codec"].(string)
						opts.VideoCodec = &(optValue)
					}
					for chk := range videoOpt {
						// fmt.Println(i)
						if videoOpt[chk] == nil {
							continue
						}
						switch chk {
						case "Codec":
							optValue := videoOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.VideoCodec = &(optValue)
						case "Bitrate":
							optValue := videoOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.VideoBitRate = &(optValue)
						case "MaxBitrate":
							optValue := videoOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.VideoMaxBitRate = &(optValue)
						case "Width":
							if videoOpt["Mode"] == "0" {
								optValue := strconv.Itoa(widthorg2) + "x" + strconv.Itoa(heightorg2)
								if optValue == "" {
									continue
								}
								opts.Resolution = &(optValue)
							} else {
								optValue := "scale=" + resolution
								if optValue == "" {
									continue
								}
								if videoOpt["Mode"] == "3" {
									optValue = optValue + ":force_original_aspect_ratio=decrease,pad=" + resolution + ":(ow-iw)/2:(oh-ih)/2"
								} else if videoOpt["Mode"] == "4" {
									width4, _ := strconv.Atoi(width.(string))
									height4, _ := strconv.Atoi(height.(string))
									fmt.Println("width : ", width4, "height : ", height4)
									width4 = width4 - widthorg2
									height4 = height4 - heightorg2
									if width4 < 0 {
										width4 = width4 * -1
									}
									if height4 < 0 {
										height4 = height4 * -1
									}
									optValue = "crop=iw-" + strconv.Itoa(width4) + ":ih-" + strconv.Itoa(height4) + "," + optValue
								}
								opts.VideoFilter = &(optValue)
							}
						case "FrameRate":
							optValue := videoOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.FrameRate = &(optValue)
						case "Profile":
							optValue := videoOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.VideoProfile = &(optValue)
						case "Level":
							optValue := videoOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.ProfileLevel = &(optValue)
						case "Crf":
							optValue := videoOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.Crf = &(optValue)

						// case "OverlayH":
						// 	if videoOpt[chk].(string)!=""{
						// 		optValue := "[1:v]scale="+videoOpt["OverlayW"].(string)+":"+videoOpt["OverlayH"].(string)+"[ol],[0:v][ol]overlay=W-w-"+videoOpt["OverlayX"].(string)+":H-h-"+videoOpt["OverlayY"].(string)
						// 		opts.Overlay = &(optValue)
						// 	}
						// case "PixelAspect":
						// 	optValue := videoOpt[chk].(string)
						// 	opts.Aspect = &(optValue)
						// case "RateControl":
						// 	optValue := videoOpt[chk].(string)
						// 	if optValue == "CBR" {
						// 		opts.VideoMaxBitRate = &(optValue)
						// 	}
						// 	opts.VideoProfile = &(optValue)
						// case "Pass":
						// 	optValue := videoOpt[chk].(string)
						// 	opts.Pass = &(optValue)
						case "GOP":
							optValue, _ := strconv.Atoi(videoOpt[chk].(string))
							opts.KeyframeInterval = &(optValue)
						case "Bframe":
							optValue, _ := strconv.Atoi(videoOpt[chk].(string))
							opts.Bframe = &(optValue)
						// case "Iframe":
						// 	optValue := videoOpt[chk].(string)
						// 	opts.VideoProfile = &(optValue)
						// case "Interlacing":
						// 	optValue := videoOpt[chk].(string)
						// 	opts.VideoProfile = &(optValue)
						// case "Deinterlacing":
						// 	optValue := videoOpt[chk].(string)
						// 	opts.VideoProfile = &(optValue)
						case "Format":
							optValue := videoOpt[chk].(string)
							opts.OutputFormat = &(optValue)
						}
					}

					for chk := range audioOpt {
						switch chk {
						case "Codec":
							optValue := audioOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.AudioCodec = &(optValue)
						case "Bitrate":
							optValue := audioOpt[chk].(string)
							if optValue == "" {
								continue
							}
							opts.AudioBitrate = &(optValue)
						}
					}

					overwrite := true
					opts.Overwrite = &overwrite

					inputPath := "/home/" + customerName + "/" + data.FileName
					ffmpegConf := &ffmpeg.Config{
						FfmpegBinPath:   "/usr/local/bin/ffmpeg",
						FfprobeBinPath:  "/usr/local/bin/ffprobe",
						ProgressEnabled: true,
					}
					progress, err := ffmpeg.
						New(ffmpegConf).
						Input(inputPath).
						Output(output).
						WithOptions(opts).
						Start(opts)

					if err != nil {
						func_error = err
						return
					}
					for msg := range progress {
						log.Printf(output+"%+v", msg)
						if msg.(ffmpeg.Progress).Progress >= 98 {
							Progress1.PP[data.Presets[i].PresetName] = 100
							dbms := database.NewDBMSWithHost("root", "Qwpo1209", "211.110.226.42", "SKB_TRANSCODING")
							res, _ := json.Marshal(Progress1.TranscodingResponse())
							sql := `UPDATE
										DR_JOB
									SET
										transcoding_json = JSON_MERGE_PATCH(transcoding_json,'` + string(res) + `')
									WHERE 
										job_id = ` + strconv.Itoa(data.JobID)
							err = dbms.MySQLExec(sql)
							if err != nil {
								log.Fatal(err)
							}
						} else if Progress1.PP[data.Presets[i].PresetName]+5 < msg.(ffmpeg.Progress).Progress {
							Progress1.PP[data.Presets[i].PresetName] = msg.(ffmpeg.Progress).Progress
							dbms := database.NewDBMSWithHost("root", "Qwpo1209", "211.110.226.42", "SKB_TRANSCODING")
							res, _ := json.Marshal(Progress1.TranscodingResponse())
							sql := `UPDATE
										DR_JOB
									SET
										transcoding_json = JSON_MERGE_PATCH(transcoding_json,'` + string(res) + `')
									WHERE 
										job_id = ` + strconv.Itoa(data.JobID)
							err = dbms.MySQLExec(sql)
							if err != nil {
								log.Fatal(err)
							}
						}
					}
				}
			}
		}()
		if func_error != nil {
			fmt.Println("Transcoding Error")
			fmt.Println(func_error)
			return func_error
		}
		customlogger.DebugLogger.Println("Preset : " + data.Transcodings[i].PresetName + " END")
	}

	wg.Wait()
	// ************ transcoding file End ************
	return nil
}
func DoOverlay(OverlayImg string, Orgfile string, OrgW string, OrgH string, OverW string, OverH string, Gravity string, GeometryX string, GeometryY string, Transparency int) error {
	overcmd := ":"
	switch Gravity {
	case "nw":
		overcmd = overcmd + GeometryX + ":" + GeometryY
	case "n":
		overcmd = overcmd + "(ow/2)-(iw/2)+" + GeometryX + ":" + GeometryY
	case "ne":
		overcmd = overcmd + "ow-iw-" + GeometryX + ":" + GeometryY
	case "w":
		overcmd = overcmd + GeometryX + ":(oh/2)-(ih/2)+" + GeometryY
	case "c":
		overcmd = overcmd + "(ow/2)-(iw/2)+" + GeometryX + ":" + "(oh/2)-(ih/2)+" + GeometryY
	case "e":
		overcmd = overcmd + "ow-iw-" + GeometryX + ":(oh/2)-(ih/2)+" + GeometryY
	case "sw":
		overcmd = overcmd + GeometryX + ":oh-ih-" + GeometryY
	case "s":
		overcmd = overcmd + "(ow/2)-(iw/2)+" + GeometryX + ":oh-ih-" + GeometryY
	case "se":
		overcmd = overcmd + "ow-iw-" + GeometryX + ",oh-ih-" + GeometryY
	}
	var floatTransparency float64
	floatTransparency = float64(Transparency) / 100
	a := fmt.Sprintf("%.1f", floatTransparency)
	fmt.Println(a)
	cmd := exec.Command("/usr/local/bin/ffmpeg", "-i", Orgfile, "-i", OverlayImg, "-filter_complex \"[1:v] scale="+OverW+":"+OverH+", pad="+OrgW+":"+OrgH+overcmd+", setsar=sar=1, format=rgba [lo];[0:v]setsar=sar=1, format=rgba [bg]; [bg][lo] blend=all_mode=addition:all_opacity="+fmt.Sprintf("%.1f", floatTransparency)+"\"", "-codec:a", strings.ReplaceAll(Orgfile, ".mp4", "")+"_ovr.mp4")
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err))
		log.Println(cmd)
		log.Println(err)
		return err
	} else {
		log.Println(Orgfile + " is overlay Transcoding Complete")
	}
	return nil
}
