package main

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"time"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/tidwall/gjson"
)


import "os"

const BindAddr = "127.0.0.1:8888"
const NodeAddr = "127.0.0.1:26657"


var (
	promLatestBlockHeight = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "latest_block_height",
		Help: "The Latest Block Height",
	})

	promLatestBlockDeltaMilliseconds = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "latest_block_delta_milliseconds",
		Help: "The Latest Block Delta Time between latest block time and current time",
	})

	promActivePeersCount = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "active_peers_count",
		Help: "Count of active peers",
	})
)

func promLatestBlockUpdater(nodeAddr string) {
	for {
		resp, err := http.Get(fmt.Sprintf("http://%v/status", nodeAddr))
		if err != nil {
			log.Fatalln(err)
		}

		body, err := io.ReadAll(resp.Body)
		latestBlockHeight, err := strconv.Atoi(gjson.Get(string(body), "result.sync_info.latest_block_height").String())
		if err != nil {
			fmt.Println("latest_block_height is not valid integer!")
		}
		log.Printf("Got latest_block_height: %v", latestBlockHeight)
		promLatestBlockHeight.Set(float64(latestBlockHeight))

		latestBlockTime := gjson.Get(string(body), "result.sync_info.latest_block_time").String()
		log.Printf("Got latest_block_time: %v", latestBlockTime)

		layout := "2006-01-02T15:04:05.000000000Z"
		t, err := time.Parse(layout, latestBlockTime)
		if err != nil {
			log.Printf("Got wrong latest_block_time: %v", latestBlockTime)
			continue
		}
		now := time.Now()
		delta := now.Sub(t).Milliseconds()
		log.Printf("Got latest_block_delta_time: %v", delta)
		promLatestBlockDeltaMilliseconds.Set(float64(delta))

		time.Sleep(2 * time.Second)
	}
}

func promActivePeersCountUpdater(nodeAddr string){
	for {
		resp, err := http.Get(fmt.Sprintf("http://%v/net_info", nodeAddr))
		if err != nil {
			log.Fatalln(err)
		}

		body, err := io.ReadAll(resp.Body)
		activePeersCount, err := strconv.Atoi(gjson.Get(string(body), "result.n_peers").String())
		if err != nil {
			log.Println("n_peers is not valid integer!")
		}
		log.Printf("Got active_peers_count: %v", activePeersCount)
		promActivePeersCount.Set(float64(activePeersCount))

		time.Sleep(2 * time.Second)
	}
}

func main(){
	bindAddr := os.Getenv("COSMOEXP_ADDR")
	nodeAddr := os.Getenv("COSMONODE_ADDR")

	if len(bindAddr) < 1 {
		bindAddr = BindAddr
	}
	if len(nodeAddr) < 1 {
		nodeAddr = NodeAddr
	}

	fmt.Printf("COSMOEXP_ADDR: %v\n", bindAddr)
	fmt.Printf("COSMONODE_ADDR: %v\n", nodeAddr)

	go promLatestBlockUpdater(nodeAddr)
	go promActivePeersCountUpdater(nodeAddr)

	http.Handle("/metrics", promhttp.Handler())
	http.ListenAndServe(bindAddr, nil)
}