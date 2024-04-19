package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
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

var logFileInfos []logFileInfo

func main() {
	loc, err := time.LoadLocation("Asia/Seoul")
	if err != nil {
		log.Printf("location error")
		return
	}

	startTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	currentTime = time.Now().Format("20060102")

	threshold := flag.String("threshold", "50", "disk usage threshold")
	bucketName := flag.String("bucketname", "empty", "s3 bucketname")
	flag.Parse()
	thresholdVal, _ := strconv.Atoi(*threshold)

	if thresholdVal < 0 || thresholdVal > 99 {
		log.Print("threshold should be between 0 and 99")
		return
	}
	if *bucketName == "empty" {
		log.Printf("bucketName: %s", *bucketName)
		return
	}

	//Usage Check
	usage, err := checkDiskUsage()
	if err != nil {
		fmt.Printf("Error checking disk usage: %s\n", err)
		os.Exit(1)
	}
	if int(usage) < thresholdVal {
		fmt.Printf("Disk usage is %d, below threshold %d\n", int(usage), thresholdVal)
		return
	}
	fmt.Printf("Disk usage is %d, above threshold %d\n", int(usage), thresholdVal)

	//add files to slice
	err = filepath.Walk(localLogPath, fileCheck)

	//file check logic and Send old files in go routine
	var wg sync.WaitGroup
	for _, el := range logFileInfos {
		if currentTime == el.ModTime {
			log.Printf("Today's log excluded")
		}

		wg.Add(1)

		go uploadFileToS3(el.Path, el.Name, *bucketName, &wg)
	}

	wg.Wait()
	endTime := time.Now().In(loc).Format("2006-01-02 15:04:05")
	log.Printf("Eviction Job started at %s and done at %s! bye~", startTime, endTime)
	return
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

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(fileName),
		Body:   file,
	})
	return err
}
