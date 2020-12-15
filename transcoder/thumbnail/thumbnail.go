package thumbnail

import (
	"fmt"
	"os/exec"
	"strconv"
	"strings"

	"github.com/new_transcoder/ftp"
	"github.com/new_transcoder/transcoder/preset"
)

// MakeThumbnail by using ffmpeg
func PresetMakeDefaultThumbnail(Thumbnail preset.DefaultThumbnail, input string, Prefix string) error {
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	findduration, err := exec.Command("mediainfo", `--Inform=General;%Duration%`, input).Output()
	if err != nil {
		return err
	}
	var Duration int
	if findduration == nil {
		Duration, err = strconv.Atoi(strings.ReplaceAll(string(findduration), "\n", ""))
	} else {
		Duration = 0
	}
	Duration = Duration - Thumbnail.StartTime*1000

	ThumbNum := Thumbnail.ThumbnailNum
	for i := 0; i < ThumbNum; i++ {
		output := Thumbnail.Output
		output = strings.ReplaceAll(output, "$prefix", Prefix)
		output = strings.ReplaceAll(output, "$count", strconv.Itoa(i+1))
		ThumbnailTime := strconv.Itoa(Thumbnail.StartTime + ((Duration/ThumbNum)*i)/1000)
		output = strings.ReplaceAll(output, "$second", ThumbnailTime)
		splitq := strings.Split(output, "/")
		thbName := splitq[len(splitq)-1]
		var width, height int
		fmt.Println(Thumbnail.Width, Thumbnail.Height)
		width = Thumbnail.Width
		height = Thumbnail.Height
		var cmd *exec.Cmd
		if width == 0 || height == 0 {
			cmd = exec.Command("/usr/local/bin/ffmpeg", "-i", input, "-ss", ThumbnailTime, "-vframes", "1", "-y", "/home"+data.HomePath+"/thb/"+thbName)
		} else {
			cmd = exec.Command("/usr/local/bin/ffmpeg", "-i", input, "-ss", ThumbnailTime, "-vframes", "1", "-y", "-s", strconv.Itoa(width)+"*"+strconv.Itoa(height), "/home"+data.HomePath+"/thb/"+thbName)
		}
		fmt.Println(cmd)
		err = cmd.Run()
		if err != nil {
			return err
		}
		for k := 0; k < len(data.ResultIP); k++ {
			err = ftp.CurlVideoUpFtp("/home"+data.HomePath+"/thb/"+thbName, output, data.ResultIP[k])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func OrgMakeDefaultThumbnail(Thumbnail preset.DefaultThumbnail, input string) error {
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	findduration, err := exec.Command("mediainfo", `--Inform=General;%Duration%`, input).Output()
	if err != nil {
		return err
	}
	fmt.Println(input)
	var Duration int
	if findduration == nil {
		Duration, err = strconv.Atoi(strings.ReplaceAll(string(findduration), "\n", ""))
	} else {
		Duration = 0
	}
	Duration = Duration - Thumbnail.StartTime*1000

	ThumbNum := Thumbnail.ThumbnailNum
	for i := 0; i < ThumbNum; i++ {
		fmt.Println(i)
		output := Thumbnail.Output
		output = strings.ReplaceAll(output, "$count", strconv.Itoa(i+1))
		ThumbnailTime := strconv.Itoa(Thumbnail.StartTime + ((Duration/ThumbNum)*i)/1000)
		output = strings.ReplaceAll(output, "$second", ThumbnailTime)
		splitq := strings.Split(output, "/")
		thbName := splitq[len(splitq)-1]
		var width, height int
		fmt.Println(Thumbnail.Width, Thumbnail.Height)
		width = Thumbnail.Width
		height = Thumbnail.Height
		var cmd *exec.Cmd
		if width == 0 || height == 0 {
			cmd = exec.Command("/usr/local/bin/ffmpeg", "-i", input, "-ss", ThumbnailTime, "-vframes", "1", "-y", "/home"+data.HomePath+"/thb/"+thbName)
		} else {
			cmd = exec.Command("/usr/local/bin/ffmpeg", "-i", input, "-ss", ThumbnailTime, "-vframes", "1", "-y", "-s", strconv.Itoa(width)+"*"+strconv.Itoa(height), "/home"+data.HomePath+"/thb/"+thbName)
		}
		fmt.Println(cmd)
		err = cmd.Run()
		if err != nil {
			return err
		}
		for k := 0; k < len(data.ResultIP); k++ {
			err = ftp.CurlVideoUpFtp("/home"+data.HomePath+"/thb/"+thbName, output, data.ResultIP[k])
			if err != nil {
				return err
			}
		}
	}
	return nil
}
func PresetMakeGridThumbnail(Thumbnail preset.GridThumbnail, input string, Prefix string) error {
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	size := fmt.Sprintf("%d", Thumbnail.Width) + "x" + fmt.Sprintf("%d", Thumbnail.Height)
	gCmd := exec.Command("ffmpeg", "-i", input, "-r", "1/"+fmt.Sprintf("%d", Thumbnail.ThumbnailInterval), "-s", size, "/home"+data.HomePath+"/thb/g-%0006d.png")
	fmt.Println(gCmd)
	err := gCmd.Run()
	if err != nil {
		return err
	}
	split := strings.Split(Thumbnail.Output, ".")
	extimg := split[len(split)-1]
	mCmd := exec.Command("montage", "-tile", fmt.Sprintf("%d", Thumbnail.Column), "-geometry", size+"+0+0", "/home"+data.HomePath+"/thb/g-*.png", "/home"+data.HomePath+"/thb/out."+extimg)
	fmt.Println(mCmd)
	err = mCmd.Run()
	if err != nil {
		return err
	}
	rCmd := exec.Command("rm", "-f", "/home"+data.HomePath+"/thb/g-*.png")
	fmt.Println(mCmd)
	err = rCmd.Run()
	if err != nil {
		return err
	}
	output := strings.ReplaceAll(Thumbnail.Output, "$prefix", Prefix)
	for i := 0; i < len(data.ResultIP); i++ {
		err = ftp.CurlVideoUpFtp("/home"+data.HomePath+"/thb/out."+extimg, output, data.ResultIP[i])
		if err != nil {
			return err
		}
	}
	return nil
}
func OrgMakeGridThumbnail(Thumbnail preset.GridThumbnail, input string) error {
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	size := fmt.Sprintf("%d", Thumbnail.Width) + "x" + fmt.Sprintf("%d", Thumbnail.Height)
	gCmd := exec.Command("ffmpeg", "-i", input, "-r", "1/"+fmt.Sprintf("%d", Thumbnail.ThumbnailInterval), "-s", size, "/home"+data.HomePath+"/thb/g-%0006d.png")
	fmt.Println(gCmd)
	err := gCmd.Run()
	if err != nil {
		return err
	}
	split := strings.Split(Thumbnail.Output, ".")
	extimg := split[len(split)-1]
	mCmd := exec.Command("montage", "-tile", fmt.Sprintf("%d", Thumbnail.Column), "-geometry", size+"+0+0", "/home"+data.HomePath+"/thb/g-*.png", "/home"+data.HomePath+"/thb/out."+extimg)
	fmt.Println(mCmd)
	err = mCmd.Run()
	if err != nil {
		return err
	}
	rCmd := exec.Command("rm", "-f", "/home"+data.HomePath+"/thb/g-*.png")
	fmt.Println(rCmd)
	err = rCmd.Run()
	if err != nil {
		return err
	}
	for i := 0; i < len(data.ResultIP); i++ {
		err = ftp.CurlVideoUpFtp("/home"+data.HomePath+"/thb/out."+extimg, Thumbnail.Output, data.ResultIP[i])
		if err != nil {
			return err
		}
	}
	return nil
}

// func MakeThumbnail() error {

// 	jobTemplate := "transcode.json"
// 	data := preset.Loadjson(jobTemplate)

// 	splitimg := strings.Split(data.DefaultThumbnail.Output, ".")[1]
// 	// orgFile := "/usr/src/tr/video/" + customerName + "/" + data.FileName
// 	// lenName := strings.SplitAfter(data.FilePath, "/")
// 	// fileName := strings.Split(lenName, "/")[len(lenName)-1]
// 	// customerUpPath := strings.ReplaceAll(replaceHome, splitMp4, "")                 // test2/

// 	fmt.Println("Duration", data.Duration)
// 	Duration2, _ := strconv.ParseFloat(data.Duration, 64)
// 	Duration3 := int(Duration2)
// 	ThumbNum := data.DefaultThumbnail.ThumbNum
// 	//Duration, err := exec.Command("mediainfo", " --Inform=\"Video;%Duration%\"", orgFile).Output()
// 	//Duration2, _ := strconv.Atoi(string(Duration))
// 	fmt.Println("Duration3", Duration3)
// 	ThumbNum2, _ := strconv.Atoi(ThumbNum)
// 	for j := 0; j < len(data.Transcodings); j++ {
// 		for i := 1; i <= ThumbNum2; i++ {
// 			splitout := strings.Split(data.Transcodings[j].Output, "/")
// 			resultname := splitout[len(splitout)-1]
// 			resultFile := "/home" + data.HomePath + "/" + resultname //result File의 위치
// 			output := data.DefaultThumbnail.Output
// 			output = strings.ReplaceAll(output, "$prefix", data.Presets[j].Prefix)
// 			output = strings.ReplaceAll(output, "$count", strconv.Itoa(i))
// 			output = strings.ReplaceAll(output, "$second", strconv.Itoa((Duration3/(ThumbNum2+1)*i)/1000))
// 			fmt.Println("thumbnail Path" + output)
// 			splitq := strings.Split(output, "/")
// 			thbName := splitq[len(splitq)-1]
// 			input := resultFile
// 			cmd := exec.Command("ffmpeg", "-i", input, "-ss", strconv.Itoa((Duration3 / (ThumbNum2 + 1) * i / 1000)), "-vcodec", splitimg, "-vframes", "1", "-y", "/home"+data.HomePath+"/thb/"+thbName)
// 			fmt.Println(cmd)
// 			var out bytes.Buffer
// 			var stderr bytes.Buffer
// 			cmd.Stdout = &out
// 			cmd.Stderr = &stderr
// 			err := cmd.Run()
// 			if err != nil {
// 				fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
// 				fmt.Println("test error", err)
// 				return err
// 			} else {
// 				log.Println(thbName + " thumbnail is created")
// 			}
// 			for k := 0; k < len(data.ResultIP); k++ {
// 				err = ftp.CurlVideoUpFtp("/home"+data.HomePath+"/thb/"+thbName, output, data.ResultIP[k])
// 				if err != nil {
// 					return err
// 				}
// 			}
// 			customlogger.InfoLogger.Println("UPLOAD thumbnail END " + output)
// 		}
// 	}
// 	return nil
// }
