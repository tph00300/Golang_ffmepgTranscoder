package preset

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
)

// Audio is audio encoding info
type Audio struct {
	Codec          string `json:"codec"`
	Profile        string `json:"profile"`
	BitrateControl string `json:"BitrateControl"`
	Coding         string `json:"Coding"`
	Bitrate        string `json:"Bitrate"`
	Sample         string `json:"Sample"`
}

// Video is video encoding info
type Video struct {
	Codec               string `json:"Codec"`
	Bitrate             string `json:"Bitrate"`
	MaxBitrate          string `json:"MaxBitrate"`
	Width               string `json:"Width"`
	Height              string `json:"Height"`
	FrameRate           string `json:"FrameRate"`
	Profile             string `json:"Profile"`
	Level               string `json:"Level"`
	PixelAspect         string `json:"PixelAspect"`
	RateControl         string `json:"RateControl"`
	Pass                string `json:"Pass"`
	GOP                 string `json:"GOP"`
	Bframe              string `json:"Bframe"`
	Iframe              string `json:"Iframe"`
	Interlacing         string `json:"Interlacing"`
	Deinterlacing       string `json:"Deinterlacing"`
	Format              string `json:"Format"`
	OverlayGeometryX    int    `json:"OverlayGeometryX"`
	OverlayGeometryy    int    `json:"OverlayGeometryY"`
	OverlayGravity      string `json:"OverlayGravity"`
	OverlayW            string `json:"OverlayW"`
	OverlayH            string `json:"OverlayH"`
	OverlayImage        string `json:"OverlayImage"`
	OverlayTransparency int    `json:"OverlayTransparency"`
	Mode                string `json:"Mode"`
}

// Presets is preset in subtask
type Presets struct {
	PresetID   int                    `json:"PresetID"`
	PresetName string                 `json:"PresetName"`
	Prefix     string                 `json:"Prefix"`
	Uploader   string                 `json:"Uploader"`
	Creater    string                 `json:"Creater"`
	Video      map[string]interface{} `json:"Video"`
	Audio      map[string]interface{} `json:"Audio"`
}

// Transcodings is subtask in Info struct
type Transcodings struct {
	PresetName string `json:"PresetName"`
	Output     string `json:"Output"`
}

// DefaultThumbnail is extracting thumbnail in Info struct
type DefaultThumbnail struct {
	ThumbnailSource string `json:"ThumbnailSource"`
	ThumbnailNum    int    `json:"ThumbnailNum"`
	Width           int    `json:"ThumbnailWidth"`
	Height          int    `json:"ThumbnailHeight"`
	StartTime       int    `json:"StartTime"`
	Output          string `json:"Output"`
}
type GridThumbnail struct {
	ThumbnailSource   string `json:"ThumbnailSource"`
	ThumbnailInterval int    `json:"ThumbnailInterval"`
	Width             int    `json:"ThumbnailWidth"`
	Height            int    `json:"ThumbnailHeight"`
	Column            int    `json:"ThumbnailColumn"`
	Output            string `json:"Output"`
}
type ResultIP struct {
	Host          string `json:"Host"`
	RemainingPath string `json:"RemainingPath"`
	HomePath      string `json:"HomePath"`
}

// Info is json file struct
type Info struct {
	JobID                 int                    `json:"JobID"`
	FileName              string                 `json:"FileName"`
	CustomerName          string                 `json:"CustomerName"`
	CustomerID            string                 `json:"CustomerID"`
	FilePath              string                 `json:"FilePath"`
	UploadIP              string                 `json:"UploadIP"`
	Uploader              string                 `json:"Uploader"`
	HomePath              string                 `json:"HomePath"`
	ResultIP              []ResultIP             `json:"ResultIP"`
	TemplateName          string                 `json:"TemplateName"`
	DefaultThumbnail      []DefaultThumbnail     `json:"DefaultThumbnail"`
	GridThumbnail         []GridThumbnail        `json:"GridThumbnail"`
	Callback              string                 `json:"Callback"`
	Transcodings          []Transcodings         `json:"Transcodings"`
	Presets               []Presets              `json:"Presets"`
	OriginHosting         string                 `json:"OriginHosting"`
	CDNHosting            string                 `json:"CDNHosting"`
	SKBStorageOutput      string                 `json:"SKBStorageOutput"`
	CustomerStorageOutput string                 `json:"CustomerStorageOutput"`
	Media                 map[string]interface{} `json:"media"`
}

// Loadjson is Transcoding json file
func Loadjson(jobTemplate string) Info {

	defer func() {
		s := recover()
		fmt.Println(s)
	}()

	// jsonFile, err := ioutil.ReadFile("/tmp/transcode.json")
	jsonFile, err := ioutil.ReadFile(jobTemplate)
	if err != nil {
		fmt.Println(err)
	}

	var info Info
	// save json content in data
	json.Unmarshal(jsonFile, &info)
	return info
}
