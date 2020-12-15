package ftp

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"

	"github.com/new_transcoder/customlogger"
	"github.com/new_transcoder/transcoder/preset"
	// "github.com/new_transcoder/response"
	// "github.com/new_transcoder/transcoder/preset"
)

func CurlDownFtp(UploadIP string, FileName string, inputfile string, downPath string) error {

	// uploadIP := data.ResultIP
	// password := "!Qwpo1209"
	// curl --ftp-create-dirs -u "atlas:Qwpo1209" "ftp://trupload.myskcdn.net:2100/test2/sample3.mp4" -O
	cmd := exec.Command("curl", "--create-dirs", UploadIP+inputfile, "-o", downPath)
	fmt.Println(cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	} else {
		log.Println(FileName + " is downloaded")
	}
	return nil
}
func CurlRemoveFtp(UploadIP string, FileName string, inputfile string) error {

	// uploadIP := data.ResultIP
	// password := "!Qwpo1209"
	// curl --ftp-create-dirs -u "atlas:Qwpo1209" "ftp://trupload.myskcdn.net:2100/test2/sample3.mp4" -O
	cmd := exec.Command("curl", UploadIP+inputfile)
	fmt.Println(cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		return err
	} else {
		log.Println(FileName + " is downloaded")
	}
	return nil
}
func CurlJsonUpFtp(res interface{}) error {

	jobTemplate := "transcode.json"

	data := preset.Loadjson(jobTemplate)
	res_json, err := json.Marshal(res)
	if err != nil {
		return err
	}
	filePath := data.FilePath // /home/atlas/test2/sample3.mp4
	customerName := strings.Split(filePath, "/")[2]
	// ftpUser := customerName + ":Qwpo1209"

	f2, _ := os.Create("/home/" + customerName + "/json/" + data.FileName + ".json")
	defer f2.Close()
	n, err := f2.WriteString(string(res_json))
	// fmt.Println(strings.Split(data.FileName, ".")[0] + ".json" + " is created")
	log.Println(strings.Split(data.FileName, ".")[0] + ".json" + " is created")
	fmt.Println("file size : ", n)
	// fmt.Println(string(res))

	// transcoding result file
	jasonFile := "/home/" + customerName + "/json/" + data.FileName + ".json"
	// minio upload file name
	objectFile := "json" + strings.ReplaceAll(data.FilePath, "/home/"+customerName, "") + ".json"

	// lenName := strings.SplitAfter(filePath, "/")
	// cutName := strings.Split(filePath, "/")[len(lenName)-1]
	// cutHome := strings.ReplaceAll(filePath, "/home/"+customerName, "")
	// upPath := strings.ReplaceAll(cutHome, cutName, "")

	// curl --ftp-create-dirs -u "atlas:Qwpo1209" "ftp://trupload.myskcdn.net:2100/test2/sample3.mp4" -O
	for i := 0; i < len(data.ResultIP); i++ {
		Wftp := data.ResultIP[i]
		cmd := exec.Command("curl", "-T", jasonFile, Wftp.Host+Wftp.HomePath+"/"+objectFile)
		fmt.Println(cmd)
		var out bytes.Buffer
		var stderr bytes.Buffer
		cmd.Stdout = &out
		cmd.Stderr = &stderr
		err = cmd.Run()
		if err != nil {
			customlogger.CritLogger.Println("Upload FAIL Response JSON", jasonFile, Wftp.Host+Wftp.HomePath+"/"+objectFile)
			return err
		} else {
			log.Println(jasonFile + " is uploaded")
			customlogger.DebugLogger.Println("Upload Success Response JSON", Wftp.Host+Wftp.HomePath+"/"+objectFile)
		}
	}
	return nil
}
func CurlVideoUpFtp(resultFile string, uploadPath string, Wftp preset.ResultIP) error { // ip, name, resultfile의 위치, 업로드할 위치

	// ftpUser := customerName + ":Qwpo1209"
	// curl --ftp-create-dirs -u "atlas:Qwpo1209" "ftp://trupload.myskcdn.net:2100/test2/sample3.mp4" -O
	cmd := exec.Command("curl", "-T", resultFile, "--ftp-create-dirs", Wftp.Host+Wftp.HomePath+uploadPath)
	fmt.Println(cmd)
	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr
	err := cmd.Run()
	if err != nil {
		fmt.Println(fmt.Sprint(err) + ": " + stderr.String())
		log.Println(cmd)
		log.Println(err)
		customlogger.CritLogger.Println("UPLOAD FAIL", "resultFile : "+resultFile, "uploadPath : "+uploadPath)
		return err
	} else {
		log.Println(resultFile + " is uploaded")
		customlogger.DebugLogger.Println("UPLOAD Success", "resultFile : "+resultFile, "uploadPath : "+uploadPath)
	}

	return nil
}
