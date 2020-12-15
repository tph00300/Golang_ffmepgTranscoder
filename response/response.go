package response

import (
	"encoding/json"
	"fmt"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/new_transcoder/transcoder/preset"
)

type PresetProgress struct {
	PP map[string]float64 `json:"ProgressPercentage"`
}
type ThumbnailList struct {
	ThumbnailPath string                 `json:"ThumbnailPath"`
	MediaInfo     map[string]interface{} `json:"ThumbnailMeta"`
}

// ConverDataList is json file struct
type ConverDataList struct {
	FileSize      int                    `json:"FileSize"`
	FilePath      string                 `json:"FilePath"`
	ThumbnailList []ThumbnailList        `json:"ThumbnailList"`
	Md5           string                 `json:"MD5"`
	PresetName    string                 `json:"PresetName"`
	MediaInfo     map[string]interface{} `json:"MediaInfo"`
}

type OrgMetadata struct {
}
type Publishing struct {
	Origin map[string]string `json:"Origin"`
	CDN    CDN               `json:"CDN"`
}
type CDN struct {
	ProgressiveDownload map[string]interface{} `json:"Progressive Download"`
	HLS                 map[string]string      `json:"HLS"`
}

// Response is json file struct
type Response struct {
	ConvertCount    int                    `json:"ConvertCount"`
	FileName        string                 `json:"FileName"`
	FilePath        string                 `json:"FilePath"`
	FileSize        int                    `json:"FileSize"`
	OrgMetadata     map[string]interface{} `json:"OrgMetadata"`
	Md5             string                 `json:"MD5"`
	ConvertDataList []ConverDataList       `json:"ConvertDataList"`
	RtMsgDetail     string                 `json:"RtMsgDetail"`
	RtMsg           string                 `json:"RtMsg"`
	Rt              string                 `json:"Rt"`
	Progressing     PresetProgress         `json:"Progressing"`
	Publishing      Publishing             `json:"Publishing"`
}
type ResponseTranscoding struct {
	Progressing PresetProgress `json:"Progressing"`
	RtMsgDetail string         `json:"RtMsgDetail"`
	RtMsg       string         `json:"RtMsg"`
	Rt          string         `json:"Rt"`
}
type ResponseUploading struct {
	OrgFilePath     string           `json:"OrgFilePath"`
	ConvertDataList []ConverDataList `json:"ConvertDataList"`
	Progressing     PresetProgress   `json:"Progressing"`
	RtMsgDetail     string           `json:"RtMsgDetail"`
	RtMsg           string           `json:"RtMsg"`
	Rt              string           `json:"Rt"`
}
type ResponseDone struct {
	RtMsgDetail string     `json:"RtMsgDetail"`
	RtMsg       string     `json:"RtMsg"`
	Rt          string     `json:"Rt"`
	Publishing  Publishing `json:"Publishing"`
}
type ResponseFail struct {
	RtMsgDetail string `json:"RtMsgDetail"`
	RtMsg       string `json:"RtMsg"`
	Rt          string `json:"Rt"`
}
type ResponseOutput struct {
	Output   map[string]interface{} `json:"Output"`
	Response map[string]interface{} `json:"Response"`
}
type Responsecode struct {
	Status    string `json:"Status"`
	RtCode    int    `json:"RtCode"`
	Message   string `json:"Message"`
	StartTime string `json:"StartTime"`
	EndTime   string `json:"EndTime"`
}

func (pp PresetProgress) TranscodingstartResponse() interface{} {
	res := make(map[string]interface{})
	res["status"] = "Transcoding"
	res["rtcode"] = 200
	res["starttime"] = time.Now().Format(time.RFC3339)
	Output := make(map[string]interface{})
	Preset := make([]interface{}, 0)

	Output["output"] = Preset

	Output["response"] = res
	return Output
}
func (pp PresetProgress) TranscodingResponse() interface{} {
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	res := make(map[string]interface{})
	res["status"] = "Transcoding"
	res["rtcode"] = 200

	Output := make(map[string]interface{})
	Preset := make([]interface{}, 0)

	for i := 0; i < len(data.Transcodings); i++ {
		presetinfo := make(map[string]interface{})
		presetinfo["progress"] = int(pp.PP[data.Transcodings[i].PresetName])
		presetinfo["presetname"] = data.Transcodings[i].PresetName
		Preset = append(Preset, presetinfo)
	}
	Output["output"] = Preset
	Output["response"] = res
	return Output
}
func UploadingResponse(jobinfo string) interface{} {
	Jobinfo := make(map[string]interface{})
	err := json.Unmarshal([]byte(jobinfo), &Jobinfo)
	if err != nil {
		fmt.Println(err)
	}
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	res := Jobinfo["response"].(map[string]interface{})
	res["status"] = "Uploading"
	res["rtcode"] = 200
	Preset := Jobinfo["output"].([]interface{})
	for i := 0; i < len(data.Transcodings); i++ {
		fmt.Println("progres1...")
		presetinfo := Preset[i].(map[string]interface{})
		splitout := strings.Split(data.Transcodings[i].Output, "/")
		resultname := splitout[len(splitout)-1]
		output := "/home" + data.HomePath + "/" + resultname
		cmdOutput, err := exec.Command("mediainfo", `--Inform=file://custom_media_json.txt1`, output).Output()
		cmdOutputStr := string(cmdOutput)
		err = json.Unmarshal([]byte(cmdOutputStr), &presetinfo)
		if err != nil {
			fmt.Println(err)
		}
		presetinfo["presetname"] = data.Transcodings[i].PresetName
		presetinfo["progress"] = 100
		if presetinfo["general"] == nil {
			presetinfo["general"] = make(map[string]interface{})
		}
		presetGeneral := presetinfo["general"].(map[string]interface{})
		presetGeneral["filepath"] = data.Transcodings[i].Output
		md5, err := exec.Command("md5sum", output).Output()
		if err != nil {
			fmt.Println(err)
		}
		presetGeneral["md5"] = strings.Split(strings.TrimRight(string(md5), "\n"), " ")[0]
		Thumbnail1 := make([]map[string]interface{}, 0)
		for j := 0; j < len(data.DefaultThumbnail); j++ {
			if data.DefaultThumbnail[j].ThumbnailSource == data.Transcodings[i].PresetName {
				input := "/home" + data.HomePath + "/" + resultname
				findduration, err := exec.Command("mediainfo", `--Inform=General;%Duration%`, input).Output()
				if err != nil {
					return err
				}
				var Duration int
				if findduration != nil {
					Duration, err = strconv.Atoi(strings.ReplaceAll(string(findduration), "\n", ""))
				} else {
					Duration = 0
				}
				Duration = Duration - data.DefaultThumbnail[j].StartTime*1000
				ThumbNum := data.DefaultThumbnail[j].ThumbnailNum
				for k := 0; k < ThumbNum; k++ {
					Thumbnailinfo := make(map[string]interface{})
					output := data.DefaultThumbnail[j].Output
					output = strings.ReplaceAll(output, "$prefix", data.Presets[i].Prefix)
					output = strings.ReplaceAll(output, "$count", strconv.Itoa(k+1))
					ThumbnailTime := strconv.Itoa(data.DefaultThumbnail[j].StartTime + ((Duration/ThumbNum)*k)/1000)
					output = strings.ReplaceAll(output, "$second", ThumbnailTime)
					fmt.Println(output, "thumbnail", j)
					Thumbnailinfo["filepath"] = output
					splitq := strings.Split(output, "/")
					thbName := splitq[len(splitq)-1]
					thbName = "/home" + data.HomePath + "/thb/" + thbName
					md5, err = exec.Command("md5sum", thbName).Output()
					if err != nil {
						fmt.Println(err)
					}
					Thumbnailinfo["md5"] = strings.Split(strings.TrimRight(string(md5), "\n"), " ")[0]
					Thumbnail1 = append(Thumbnail1, Thumbnailinfo)
				}
			}
		}
		if len(Thumbnail1) > 0 {
			presetinfo["thumbnail"] = Thumbnail1
		}
		fmt.Println(presetinfo["thumbnail"])
		GridThumbnail := make([]map[string]interface{}, 0)
		for j := 0; j < len(data.GridThumbnail); j++ {
			if data.GridThumbnail[j].ThumbnailSource == data.Transcodings[i].PresetName {
				Thumbnail := make(map[string]interface{})
				Thumbnail["filepath"] = data.GridThumbnail[j].Output
				GridThumbnail = append(GridThumbnail, Thumbnail)
			}
		}
		if len(GridThumbnail) > 0 {
			presetinfo["gridthumbnail"] = GridThumbnail
		}
		presetinfo["general"] = presetGeneral
		Preset[i] = presetinfo
	}
	for i := 0; i < len(data.ResultIP); i++ {
		UploadPath := strings.ReplaceAll(data.ResultIP[i].RemainingPath, "/home"+data.HomePath, "")
		presetinfo := make(map[string]interface{})
		if UploadPath != "" {
			splitout := strings.Split(UploadPath, "/")
			resultname := splitout[len(splitout)-1]
			output := "/home" + data.HomePath + "/" + data.FileName
			cmdOutput, err := exec.Command("mediainfo", `--Inform=file://custom_media_json.txt1`, output).Output()
			if err != nil {
				fmt.Println(err)
			}
			cmdOutputStr := string(cmdOutput)
			json.Unmarshal([]byte(cmdOutputStr), &presetinfo)
			presetGeneral := presetinfo["general"].(map[string]interface{})
			presetGeneral["filepath"] = UploadPath
			splitq := strings.Split(output, "/")
			thbName := splitq[len(splitq)-1]
			thbName = "/home" + data.HomePath + "/thb/" + thbName
			md5, err := exec.Command("md5sum", thbName).Output()
			if err != nil {
				fmt.Println(err)
			}
			presetGeneral["md5"] = strings.Split(strings.TrimRight(string(md5), "\n"), " ")[0]
			presetinfo["general"] = presetGeneral
			presetinfo["filename"] = resultname
			presetinfo["presetname"] = "original"
		}
		fmt.Println("progres2...")
		var flag bool
		flag = false
		Thumbnail := make([]map[string]interface{}, 0)
		for j := 0; j < len(data.DefaultThumbnail); j++ {
			if data.DefaultThumbnail[j].ThumbnailSource == "original" {
				flag = true
				input := "/home" + data.HomePath + "/" + data.FileName
				findduration, err := exec.Command("mediainfo", `--Inform=General;%Duration%`, input).Output()
				if err != nil {
					fmt.Println(err)
				}
				var Duration int
				if findduration != nil {
					Duration, err = strconv.Atoi(strings.ReplaceAll(string(findduration), "\n", ""))
				} else {
					Duration = 0
				}
				Duration = Duration - data.DefaultThumbnail[j].StartTime*1000
				ThumbNum := data.DefaultThumbnail[j].ThumbnailNum
				for k := 0; k < ThumbNum; k++ {
					Thumbnailinfo := make(map[string]interface{})
					output := data.DefaultThumbnail[j].Output
					output = strings.ReplaceAll(output, "$count", strconv.Itoa(k+1))
					ThumbnailTime := strconv.Itoa(data.DefaultThumbnail[j].StartTime + ((Duration/ThumbNum)*k)/1000)
					output = strings.ReplaceAll(output, "$second", ThumbnailTime)
					Thumbnailinfo["filepath"] = output
					splitq := strings.Split(output, "/")
					thbName := splitq[len(splitq)-1]
					thbName = "/home" + data.HomePath + "/thb/" + thbName
					md5, err := exec.Command("md5sum", thbName).Output()
					if err != nil {
						fmt.Println(err)
					}
					Thumbnailinfo["md5"] = strings.Split(strings.TrimRight(string(md5), "\n"), " ")[0]
					Thumbnail = append(Thumbnail, Thumbnailinfo)
				}
			}
		}
		var flag2 bool
		flag2 = false
		GridThumbnail := make([]map[string]interface{}, 0)
		for j := 0; j < len(data.GridThumbnail); j++ {
			if data.GridThumbnail[j].ThumbnailSource == "original" {
				flag2 = true
				Thumbnail := make(map[string]interface{})
				Thumbnail["filepath"] = data.GridThumbnail[j].Output
				GridThumbnail = append(GridThumbnail, Thumbnail)
			}
		}
		if len(GridThumbnail) > 0 {
			presetinfo["gridthumbnail"] = GridThumbnail
		}
		if len(Thumbnail) > 0 {
			presetinfo["thumbnail"] = Thumbnail
		}
		fmt.Println(presetinfo["thumbnail"])
		if UploadPath != "" || flag == true || flag2 == true {
			presetinfo["presetname"] = "original"
			Preset = append(Preset, presetinfo)
		}
	}
	Jobinfo["output"] = Preset
	Jobinfo["response"] = res

	return Jobinfo
}
func DoneResponse(jobinfo string) interface{} {
	Jobinfo := make(map[string]interface{})
	json.Unmarshal([]byte(jobinfo), &Jobinfo)
	jobTemplate := "transcode.json"
	data := preset.Loadjson(jobTemplate)
	res := Jobinfo["response"].(map[string]interface{})
	res["status"] = "Done"
	res["rtcode"] = 200
	res["endtime"] = time.Now().Format(time.RFC3339)
	Preset := Jobinfo["output"].([]interface{})
	for i := 0; i < len(Preset); i++ {
		presetinfo := Preset[i].(map[string]interface{})
		if data.CDNHosting != "" {
			if presetinfo["general"] != nil {
				presetgeneral := presetinfo["general"].(map[string]interface{})
				presetinfo["cdnpath"] = data.CDNHosting + presetgeneral["filepath"].(string)
			}
		}
		if data.OriginHosting != "" {
			if presetinfo["general"] != nil {
				presetgeneral := presetinfo["general"].(map[string]interface{})
				presetinfo["orgpath"] = data.OriginHosting + presetgeneral["filepath"].(string)
			}
		}
		if presetinfo["thumbnail"] != nil {
			Thumbnail := presetinfo["thumbnail"].([]interface{})
			for j := 0; j < len(Thumbnail); j++ {
				thumbnailinfo := Thumbnail[j].(map[string]interface{})
				if data.CDNHosting != "" {
					thumbnailinfo["cdnpath"] = data.CDNHosting + thumbnailinfo["filepath"].(string)
				}
				if data.OriginHosting != "" {
					thumbnailinfo["orgpath"] = data.OriginHosting + thumbnailinfo["filepath"].(string)
				}
			}
			presetinfo["thumbnail"] = Thumbnail
		}
		if presetinfo["gridthumbnail"] != nil {
			Thumbnail := presetinfo["gridthumbnail"].([]interface{})
			for j := 0; j < len(Thumbnail); j++ {
				thumbnailinfo := Thumbnail[j].(map[string]interface{})
				if data.CDNHosting != "" {
					thumbnailinfo["cdnpath"] = data.CDNHosting + thumbnailinfo["filepath"].(string)
				}
				if data.OriginHosting != "" {
					thumbnailinfo["orgpath"] = data.OriginHosting + thumbnailinfo["filepath"].(string)
				}
			}
			presetinfo["gridthumbnail"] = Thumbnail
		}
		Preset[i] = presetinfo
	}
	Jobinfo["output"] = Preset
	Jobinfo["response"] = res
	return Jobinfo
}

// // ResponseJson is Transcoding result json file\
// func UploadingResponse() ResponseUploading { //UploadingResponse
// 	jobTemplate := "transcode.json"
// 	data := preset.Loadjson(jobTemplate)
// 	res := ResponseUploading{}

// 	rtMsg := "Uploading"
// 	res.RtMsg = rtMsg
// 	rt := "200"
// 	res.Rt = rt
// 	res.RtMsgDetail = ""
// 	list := make([]ConverDataList, 0)
// 	for i := 0; i < len(data.Transcodings); i++ {
// 		convertlist := ConverDataList{}
// 		splitout := strings.Split(data.Transcodings[i].Output, "/")
// 		resultname := splitout[len(splitout)-1]
// 		output := "/home" + data.HomePath + "/" + resultname
// 		filesize, _ := exec.Command("ffprobe", "-v", "error", "-show_entries", "format=size", "-of", "default=noprint_wrappers=1:nokey=1", output).Output()
// 		convertlist.FileSize, _ = strconv.Atoi(strings.TrimRight(string(filesize), "\n"))

// 		replaceHome := strings.ReplaceAll(data.Transcodings[i].Output, "/home"+data.HomePath, "") // /home/woori/test/sample2.mp4 -> /test/sample2.mp4
// 		// splitMp4 := data.FileName                                                                 // sample2.mp4
// 		filepath := replaceHome
// 		convertlist.FilePath = filepath
// 		// if data.DefaultThumbnail
// 		// replaceHome = strings.ReplaceAll(data.FilePath, "/home/"+customerName+"/", "")
// 		// customerUpPath = strings.ReplaceAll(replaceHome, data.FileName, "")
// 		// thbName := strings.Split(data.FileName, ".")[0] + ".png"
// 		// // if data.DefaultThumbnail
// 		// thumbnailPath := "/thb/" + customerUpPath + thbName
// 		// convertlist.ThumbnailPath = thumbnailPath
// 		ThumbNum := data.DefaultThumbnail.ThumbNum
// 		Duration2, _ := strconv.ParseFloat(data.Duration, 64)
// 		Duration3 := int(Duration2)
// 		ThumbNum2, _ := strconv.Atoi(ThumbNum)
// 		for j := 0; j < ThumbNum2; j++ {
// 			thumbnail := ThumbnailList{}
// 			output := data.DefaultThumbnail.Output
// 			output = strings.ReplaceAll(output, "$prefix", data.Presets[i].Prefix)
// 			output = strings.ReplaceAll(output, "$count", strconv.Itoa(j+1))
// 			output = strings.ReplaceAll(output, "$second", strconv.Itoa(Duration3/(ThumbNum2+1)*i/1000))
// 			thumbnail.ThumbnailPath = output
// 			splitthb := strings.Split(output, "/")
// 			thumbname := splitthb[len(splitthb)-1]
// 			fmt.Println("thumbname : ", thumbname)
// 			cmd2 := exec.Command("mediainfo", "--Output=JSON",`--Inform="file:///home/atlas/new_transcoder/main/custom_media_json.txt1"`, "/home"+data.HomePath+"/thb/"+thumbname)
// 			fmt.Println(cmd2)
// 			cmd, _ := exec.Command("mediainfo", "--Output=JSON",`--Inform="file:///home/atlas/new_transcoder/main/custom_media_json.txt1"`, "/home"+data.HomePath+"/thb/"+thumbname).Output()
// 			var thumbnailinfo map[string]interface{}
// 			fmt.Println(cmd)
// 			json.Unmarshal(cmd, &thumbnailinfo)
// 			fmt.Println(thumbnailinfo)
// 			thumbnail.MediaInfo = thumbnailinfo
// 			convertlist.ThumbnailList = append(convertlist.ThumbnailList, thumbnail)
// 		}

// 		md5, _ := exec.Command("md5sum", output).Output()
// 		convertlist.Md5 = strings.Split(strings.TrimRight(string(md5), "\n"), " ")[0]

// 		presetname := data.Presets[i].PresetName
// 		convertlist.PresetName = presetname

// 		var videoinfo map[string]interface{}
// 		videoinfo=make(map[string]interface{})
// 		cmd, _ := exec.Command("mediainfo",  "--output=JSON",`--Inform="file:///home/atlas/new_transcoder/main/custom_media_json.txt1"`, output).Output()
// 		json.Unmarshal(cmd, &videoinfo)
// 		fmt.Println(videoinfo)
// 		convertlist.MediaInfo = videoinfo
// 		list = append(list, convertlist)
// 	}
// 	res.ConvertDataList = list
// 	var progress PresetProgress
// 	progress.PP = make(map[string]float64)
// 	for i := 0; i < len(data.Presets); i++ {
// 		progress.PP[data.Presets[i].PresetName] = 100.0
// 	}
// 	res.Progressing = progress
// 	return res
// }
func FailResponse(err error, rtMsg string) interface{} {
	res := make(map[string]interface{})
	res["rtcode"] = 400
	res["errortime"] = time.Now().Format(time.RFC3339)
	if rtMsg == "Transcoding" {
		res["status"] = "Transcoding"
		res["Message"] = "Transcoding Error or invalid transcoding option"
	} else if rtMsg == "OrgUploading" {
		res["status"] = "Uploading"
		res["Message"] = "Upload Error while uploading $Output to storage($storage)"
	} else if rtMsg == "Size" {
		res["status"] = "Transoding"
		res["Message"] = "Transcoding Error (0 size file is created )"
	} else if rtMsg == "Thumbnail" {
		res["status"] = "Transcoding"
		res["Message"] = "Thumbnail error or Invalid image extraction parameters"
	} else if rtMsg == "Media" {
		res["status"] = "Uploading"
		res["Message"] = "Media info error (problems creating media info)"
	} else if rtMsg == "lastUp" {
		res["status"] = "Uploading"
		res["Message"] = "Upload Error  while uploading $Output to storage($storage)"
	}

	Output := make(map[string]interface{})
	Output["response"] = res
	return Output
}

// func DoneResponse() ResponseDone { // Done, 200 JSON

// 	jobTemplate := "transcode.json"
// 	data := preset.Loadjson(jobTemplate)
// 	fmt.Println("Done ResponsJson start")
// 	res := ResponseDone{}
// 	rtMsg := "Done"
// 	res.RtMsg = rtMsg
// 	rt := "200"
// 	res.Rt = rt
// 	res.RtMsgDetail = ""
// 	var ThumbnailList map[string]interface{}
// 	ThumbnailList = make(map[string]interface{})
// 	res.Publishing.Origin = make(map[string]string)
// 	res.Publishing.CDN.ProgressiveDownload = make(map[string]interface{})
// 	if data.OriginHosting != "" && data.SKBStorageOutput != "" {
// 		res.Publishing.Origin["Org"] = data.OriginHosting + data.SKBStorageOutput
// 		res.Publishing.CDN.ProgressiveDownload["Org"] = data.CDNHosting + data.SKBStorageOutput
// 	}
// 	for i := 0; i < len(data.Transcodings); i++ {
// 		if data.OriginHosting != "" {
// 			res.Publishing.Origin[data.Transcodings[i].PresetName] = data.OriginHosting + data.Transcodings[i].Output
// 		}
// 		if data.CDNHosting != "" {
// 			res.Publishing.CDN.ProgressiveDownload[data.Transcodings[i].PresetName] = data.CDNHosting + data.Transcodings[i].Output
// 			if data.DefaultThumbnail.ThumbNum != "" {
// 				Duration2, _ := strconv.ParseFloat(data.Duration, 64)
// 				Duration3 := int(Duration2)
// 				Thumbnum, _ := strconv.Atoi(data.DefaultThumbnail.ThumbNum)
// 				var thumbnail []string
// 				for j := 0; j < Thumbnum; j++ {
// 					output := data.CDNHosting + data.DefaultThumbnail.Output
// 					output = strings.ReplaceAll(output, "$prefix", data.Presets[i].Prefix)
// 					output = strings.ReplaceAll(output, "$count", strconv.Itoa(j+1))
// 					output = strings.ReplaceAll(output, "$second", strconv.Itoa(Duration3/(Thumbnum+1)*i/1000))
// 					thumbnail = append(thumbnail, output)
// 				}
// 				ThumbnailList[data.Transcodings[i].PresetName] = thumbnail
// 			}
// 		}
// 	}
// 	if data.DefaultThumbnail.ThumbNum != "" && data.CDNHosting != "" {
// 		res.Publishing.CDN.ProgressiveDownload["ThumbnailList"] = ThumbnailList
// 	}
// 	// fmt.Println(list)
// 	// ************ Make ConverDataList ************

// 	return res
// }

// }
// func (pp PresetProgress) TranscodingResponse() ResponseTranscoding { // Transcoding 중 200, Progressing Update

// 	// fmt.Println("chk debug : response.go")
// 	res := ResponseTranscoding{}

// 	rtMsg := "Transcoding"
// 	res.RtMsg = rtMsg
// 	rt := "200"
// 	res.Rt = rt
// 	res.RtMsgDetail = ""

// 	res.Progressing = pp
// 	// ************ Make ConverDataList ************
// 	// fmt.Println(list)
// 	// ************ Make ConverDataList ************

// 	return res

// }
