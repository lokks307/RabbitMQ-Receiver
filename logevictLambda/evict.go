package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

const localLogPath string = "/home/ec2-user/logs/"

var currentTime string

type logFileInfo struct {
	Name    string
	Path    string
	ModTime string
}
type webHookMessage struct {
	Message string `json:"text"`
}

var logFileInfos []logFileInfo

func main() {
	ticker := time.NewTicker(30 * time.Minute) // 30분마다 실행
	defer ticker.Stop()

	currentTime = time.Now().Format("20060102")

	threshold := flag.String("threshold", "50", "disk usage threshold")
	bucketName := flag.String("bucketname", "empty", "s3 bucketname")
	flag.Parse()
	thresholdVal, _ := strconv.Atoi(*threshold)

	if thresholdVal < 0 || thresholdVal > 99 {
		log.Print("threshold should be between 0 and 99")
		sendWebHook("threshold should be between 0 and 99")
		return
	}
	if *bucketName == "empty" {
		log.Printf("bucketName: %s", *bucketName)
		sendWebHook("bucketempty")
		return
	}
	// 처음 실행을 위해 고루틴 호출
	go routine(thresholdVal, *bucketName)

	for {
		select {
		case <-ticker.C:
			go routine(thresholdVal, *bucketName) // 30분마다 고루틴 호출
		}
	}
}
func routine(thresholdVal int, bucketName string) {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("location error")
		sendWebHook("location error")
		return
	}

	startTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	//Usage Check
	usage, err := checkDiskUsage()
	if err != nil {
		log.Printf("Error checking disk usage: %s\n", err)
		sendWebHook("Error checking disk usage")
		os.Exit(1)
	}
	if int(usage) < thresholdVal {
		l := fmt.Sprintf("Disk usage is %d, below threshold %d\n", int(usage), thresholdVal)
		log.Printf(l)
		sendWebHook(l)
		return
	}

	l := fmt.Sprintf("Disk usage is %d, above threshold %d\n", int(usage), thresholdVal)
	log.Printf(l)
	sendWebHook(l)
	//add files to slice
	err = filepath.Walk(localLogPath, fileCheck)

	//file check logic and Send old files in go routine
	var wg sync.WaitGroup
	for _, el := range logFileInfos {
		if currentTime == el.ModTime {
			log.Printf("Today's log excluded")
		}

		wg.Add(1)

		go uploadFileToS3(el.Path, el.Name, bucketName, &wg)
	}

	wg.Wait()
	endTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	l = fmt.Sprintf("Eviction Job started at %s and done at %s! bye~", startTime, endTime)
	log.Println(l)
	sendWebHook(l)

}
func fileCheck(path string, info os.FileInfo, err error) error {
	if err != nil {
		fmt.Printf("Error accessing path %s: %v\n", path, err)
		return nil
	}

	if info.IsDir() {
		log.Printf("%s is dir, so continue", path)
		return nil
	}

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("Error:", err)
		return nil
	}

	//If not dir return file path and made name
	modTime := fileInfo.ModTime().Format("20060102")
	fileName := fileInfo.Name()

	logFileInfos = append(logFileInfos, logFileInfo{Name: fileName, Path: path, ModTime: modTime})

	return nil
}

func checkDiskUsage() (float64, error) {
	cmd := exec.Command("df", "--output=pcent", "/")
	output, err := cmd.Output()
	if err != nil {
		return 0, err
	}
	lines := strings.Fields(string(output))
	if len(lines) < 2 {
		return 0, fmt.Errorf("unexpected output: %s", output)
	}
	pctStr := strings.TrimRight(lines[1], "%")
	var pct float64
	fmt.Sscanf(pctStr, "%f", &pct)
	return pct, nil
}

func uploadFileToS3(filePath, fileName, bucket string, wg *sync.WaitGroup) error {
	defer wg.Done()

	sess := session.Must(session.NewSession(&aws.Config{
		Region: aws.String("ap-northeast-2"),
	}))
	uploader := s3manager.NewUploader(sess)

	log.Printf("Sending file %s to %s bucket", filePath, bucket)

	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	upperDir := filepath.Base(filepath.Dir(filePath))
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(path.Join(upperDir, fileName)),
		Body:   file,
	})

	// err = os.Remove(filePath)
	// if err != nil {
	// 	log.Printf("Failed to remove file %s: %v", filePath, err)
	// } else {
	// 	log.Printf("File %s removed successfully", filePath)
	// }
	return err
}

func sendWebHook(msg string) error {
	log.Println("sending webhook")
	webhookURL := "https://lokksio.webhook.office.com/webhookb2/a4303448-984f-4c61-9c6d-bf574051f4c7@813bddad-fdc2-434d-b0a0-9c831c139401/IncomingWebhook/553927d201904fabbb988ec59166db31/f149ba3b-8aa6-4944-92de-6b6471a157ae"
	message := webHookMessage{
		Message: msg,
	}
	messageBytes, err := json.Marshal(message)
	if err != nil {
		fmt.Println("Failed to serialize message:", err)
		return err
	}

	req, err := http.NewRequest("POST", webhookURL, bytes.NewBuffer(messageBytes))
	if err != nil {
		fmt.Println("Failed to create HTTP request:", err)
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Println("Failed to send HTTP request:", err)
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
