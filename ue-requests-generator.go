package main

import (
	"fmt"
	"time"
	"os"
	"os/exec"
	"strings"
	"strconv"

	"gonum.org/v1/gonum/stat/distuv"
	"golang.org/x/exp/rand"
	"github.com/p76081158/ue-requests-generator/module/net"
)

var Cmd = "curl google.com"
var Resource_pattern = "500:50"
var Interval_delay = 2
var TimeWindow = 5
var Request_ratio = 10
var TotalRequestSend = 0
var TotalDelaySum = 0

// convert string to int
func StringToInt(s string) int {
    i, err := strconv.Atoi(s)
    if err != nil {
        // handle error
        fmt.Println(err)
        os.Exit(2)
    }
	return i
}

// sending requests to server
func SendRequest() {
	input_cmd := Cmd
	cmd       := exec.Command("/bin/sh", "-c", input_cmd)
	cmd.Stdin  = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		fmt.Printf("Got error: %s\n", err.Error())
	}
	TotalRequestSend++
	return
}

// sending requests with poisson distribution during the timewindow
func RequestPoisson(lambda float64, request_num int) {
	r := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	p := distuv.Poisson{lambda, r}
	for i := 0; i < request_num; i++ {
		request_delay := int(p.Rand())
		go SendRequest()
		time.Sleep(time.Duration(request_delay) * time.Millisecond)
		TotalDelaySum += request_delay
	}
}

// generate lambda of requests number and requests delay within a timewindow
// number of request is generate by poisson distribution
// e.g. num_lambda   = 250  (average number requests between timewindow is 250)
//      delay_lambda = 20   (average delay between requests is 20ms)
func RequestPatternGenerator(resource int, duration int) {
	timeWindowNum := duration / TimeWindow
	num_lambda    := float64((float64(resource) / float64(Request_ratio)) * float64(TimeWindow))
	r             := rand.New(rand.NewSource(uint64(time.Now().UnixNano())))
	num_poisson   := distuv.Poisson{num_lambda, r}
	fmt.Println("Timewindow length: ", TimeWindow, "s")
	fmt.Println("Number of timewindows: ", timeWindowNum)
	fmt.Println("Lambda of request number every timewindow: ", num_lambda)
	fmt.Println("")
	for i := 0; i < timeWindowNum; i++ {
		request_nums := int(num_poisson.Rand())
		fmt.Println("Timewindow id: ", i+1)
		fmt.Println("Number of requests: ", request_nums)
		if request_nums == 0 {
			continue
		}
		delay_lambda := float64(1000.0 * float64(TimeWindow) / float64(request_nums))
		fmt.Println("Lambda of request delay: ", delay_lambda, "ms")
		fmt.Println("")
		go RequestPoisson(delay_lambda ,request_nums)
		time.Sleep(time.Duration(TimeWindow) * time.Second)
		fmt.Println("")
	}
}

// schedule between request interval
func RequestScheduler(pattern string) {
	interval := strings.Split(pattern, ",")
	fmt.Println("Total numbers of Interval: ", len(interval))
	fmt.Println("")
	for i := 0; i < len(interval); i++ {
		temp := strings.Split(interval[i], ":")
		resource := StringToInt(temp[0])
		duration := StringToInt(temp[1])
		fmt.Println("Interval id: ", i+1)
		fmt.Println("")
		go RequestPatternGenerator(resource, duration)
		if i == 0 {
			time.Sleep(time.Duration(duration) * time.Second)
		} else {
			time.Sleep(time.Duration(duration + Interval_delay) * time.Second)
		}
		fmt.Println("")
	}
}

// example cmd :                           ./ue-requests-generator "curl google.com" none 500:10,400:15 500
// example cmd (select network interface): ./ue-requests-generator "curl google.com" eth3 500:10,400:15 500

func main() {
	if len(os.Args) != 5 {
		fmt.Printf("Usage : %s <curl cmd> <specify interface name> <resource pattern> <request ratio>\n", os.Args[0])
		os.Exit(0)
	}
	if (os.Args[1]!="") {
		Cmd = string(os.Args[1])
	}
	if (os.Args[2]!="") {
		if string(os.Args[2]) != "none" {
			net.CheckInterface(string(os.Args[2]))
			Cmd += " --interface " + string(os.Args[2])
			fmt.Println(Cmd)
			fmt.Println("")
		}
	}
	if (os.Args[3]!="") {
		Resource_pattern = string(os.Args[3])
	}
	if (os.Args[4]!="") {
		Request_ratio = StringToInt(string(os.Args[4]))
	}

	start := time.Now()
	RequestScheduler(Resource_pattern)

	fmt.Println("")
	fmt.Println("Duration of execution time: ", time.Since(start))
	fmt.Println("Delay time between intervals: ", Interval_delay, "s")
	fmt.Println("Total number of requests sended: ", TotalRequestSend)
	fmt.Println("Average sending delay between requests: ", float64(TotalDelaySum / TotalRequestSend), "ms")
}