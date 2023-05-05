package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"time"

	log "github.com/sirupsen/logrus"
)

var (
	ListenAddress  = flag.String("listen", ":9017", "listen addresses")
	NimbleAddress  = flag.String("nimble", "http://127.0.0.1:8082", "nimble address")
	NimbleAuthSalt = flag.String("auth_salt", "", "auth salt (default empty)")
	NimbleAuthHash = flag.String("auth_hash", "", "auth hash (default empty)")
	LogLevel       = flag.String("loglevel", "info", "Log level")
	LogFmt         = flag.String("logfmt", "json", "Log formatter (normal, json)")
)

type ServerStatus struct {
	Connections int `json:"Connections"`
	OutRate     int `json:"OutRate"`
	SysInfo     struct {
		Ap   int    `json:"ap"`
		Scl  string `json:"scl"`
		Tpms int64  `json:"tpms"`
		Fpms int64  `json:"fpms"`
		Tsss int    `json:"tsss"`
		Fsss int    `json:"fsss"`
	} `json:"SysInfo"`
	RAMCacheSize     int `json:"RamCacheSize"`
	FileCacheSize    int `json:"FileCacheSize"`
	MaxRAMCacheSize  int `json:"MaxRamCacheSize"`
	MaxFileCacheSize int `json:"MaxFileCacheSize"`
}

type SrtReceivers struct {
	SrtReceivers []struct {
		ID       string `json:"id"`
		Streamid string `json:"streamid"`
		State    string `json:"state"`
		Stats    struct {
			Time   int `json:"time"`
			Window struct {
				Flow       int `json:"flow"`
				Congestion int `json:"congestion"`
				Flight     int `json:"flight"`
			} `json:"window"`
			Link struct {
				Rtt              float64 `json:"rtt"`
				MbpsBandwidth    float64 `json:"mbpsBandwidth"`
				MbpsMaxBandwidth int     `json:"mbpsMaxBandwidth"`
			} `json:"link"`
			Recv struct {
				PacketsReceived              int     `json:"packetsReceived"`
				PacketsReceivedRetransmitted int     `json:"packetsReceivedRetransmitted"`
				PacketsLost                  int     `json:"packetsLost"`
				PacketsDropped               int     `json:"packetsDropped"`
				PacketsBelated               int     `json:"packetsBelated"`
				NAKsSent                     int     `json:"NAKsSent"`
				BytesReceived                int     `json:"bytesReceived"`
				BytesLost                    int     `json:"bytesLost"`
				BytesDropped                 int     `json:"bytesDropped"`
				MbpsRate                     float64 `json:"mbpsRate"`
			} `json:"recv"`
		} `json:"stats"`
	} `json:"SrtReceivers"`
}

type SrtSenders struct {
	SrtSenders []struct {
		ID       string `json:"id"`
		Streamid string `json:"streamid"`
		State    string `json:"state"`
		Stats    struct {
			Time   int `json:"time"`
			Window struct {
				Flow       int `json:"flow"`
				Congestion int `json:"congestion"`
				Flight     int `json:"flight"`
			} `json:"window"`
			Link struct {
				Rtt              float64 `json:"rtt"`
				MbpsBandwidth    float64 `json:"mbpsBandwidth"`
				MbpsMaxBandwidth int     `json:"mbpsMaxBandwidth"`
			} `json:"link"`
			Recv struct {
				PacketsReceived              int     `json:"packetsReceived"`
				PacketsReceivedRetransmitted int     `json:"packetsReceivedRetransmitted"`
				PacketsLost                  int     `json:"packetsLost"`
				PacketsDropped               int     `json:"packetsDropped"`
				PacketsBelated               int     `json:"packetsBelated"`
				NAKsSent                     int     `json:"NAKsSent"`
				BytesReceived                int     `json:"bytesReceived"`
				BytesLost                    int     `json:"bytesLost"`
				BytesDropped                 int     `json:"bytesDropped"`
				MbpsRate                     float64 `json:"mbpsRate"`
			} `json:"recv"`
		} `json:"stats"`
	} `json:"SrtSenders"`
}

func processRequest(w http.ResponseWriter, r *http.Request) {
	resp, err := getMetrics("/manage/srt_sender_stats")
	if err == nil {
		var result SrtSenders
		if err := json.Unmarshal([]byte(resp), &result); err != nil {
			log.Errorf("Can not unmarshal JSON from /manage/srt_sender_stats")
		} else {
			for _, rec := range result.SrtSenders {
				fmt.Fprintf(w, "nimble_srt_sender_time { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Time)
				fmt.Fprintf(w, "nimble_srt_sender_window_flow { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Window.Flow)
				fmt.Fprintf(w, "nimble_srt_sender_window_congestion { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Window.Congestion)
				fmt.Fprintf(w, "nimble_srt_sender_window_flight { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Window.Flight)

				fmt.Fprintf(w, "nimble_srt_sender_link_rtt { stream_id=\"%s\", id=\"%s\"} %f\n", rec.Streamid, rec.ID, rec.Stats.Link.Rtt)
				fmt.Fprintf(w, "nimble_srt_sender_link_bandwidth { stream_id=\"%s\", id=\"%s\"} %f\n", rec.Streamid, rec.ID, rec.Stats.Link.MbpsBandwidth)
				fmt.Fprintf(w, "nimble_srt_sender_link_bandwidth_max { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Link.MbpsMaxBandwidth)

				fmt.Fprintf(w, "nimble_srt_sender_recv_packets_received { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsReceived)
				fmt.Fprintf(w, "nimble_srt_sender_recv_packets_received_retransmitted { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsReceivedRetransmitted)
				fmt.Fprintf(w, "nimble_srt_sender_recv_packets_lost { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsLost)
				fmt.Fprintf(w, "nimble_srt_sender_recv_packets_dropped { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsDropped)
				fmt.Fprintf(w, "nimble_srt_sender_recv_packets_belated { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsBelated)
				fmt.Fprintf(w, "nimble_srt_sender_recv_naks_sent { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.NAKsSent)
				fmt.Fprintf(w, "nimble_srt_sender_recv_bytes_lost { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.BytesLost)
				fmt.Fprintf(w, "nimble_srt_sender_recv_bytes_dropped { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.BytesDropped)
				fmt.Fprintf(w, "nimble_srt_sender_recv_mbps_rate { stream_id=\"%s\", id=\"%s\"} %f\n", rec.Streamid, rec.ID, rec.Stats.Recv.MbpsRate)
			}
		}
	}
	resp, err = getMetrics("/manage/srt_receiver_stats")
	if err == nil {
		var result SrtReceivers
		if err := json.Unmarshal([]byte(resp), &result); err != nil {
			log.Errorf("Can not unmarshal JSON from /manage/srt_receiver_stats")
		} else {
			for _, rec := range result.SrtReceivers {
				fmt.Fprintf(w, "nimble_srt_receiver_time { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Time)
				fmt.Fprintf(w, "nimble_srt_receiver_window_flow { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Window.Flow)
				fmt.Fprintf(w, "nimble_srt_receiver_window_congestion { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Window.Congestion)
				fmt.Fprintf(w, "nimble_srt_receiver_window_flight { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Window.Flight)

				fmt.Fprintf(w, "nimble_srt_receiver_link_rtt { stream_id=\"%s\", id=\"%s\"} %f\n", rec.Streamid, rec.ID, rec.Stats.Link.Rtt)
				fmt.Fprintf(w, "nimble_srt_receiver_link_bandwidth { stream_id=\"%s\", id=\"%s\"} %f\n", rec.Streamid, rec.ID, rec.Stats.Link.MbpsBandwidth)
				fmt.Fprintf(w, "nimble_srt_receiver_link_bandwidth_max { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Link.MbpsMaxBandwidth)

				fmt.Fprintf(w, "nimble_srt_receiver_recv_packets_received { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsReceived)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_packets_received_retransmitted { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsReceivedRetransmitted)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_packets_lost { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsLost)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_packets_dropped { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsDropped)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_packets_belated { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.PacketsBelated)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_naks_sent { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.NAKsSent)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_bytes_lost { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.BytesLost)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_bytes_dropped { stream_id=\"%s\", id=\"%s\"} %d\n", rec.Streamid, rec.ID, rec.Stats.Recv.BytesDropped)
				fmt.Fprintf(w, "nimble_srt_receiver_recv_mbps_rate { stream_id=\"%s\", id=\"%s\"} %f\n", rec.Streamid, rec.ID, rec.Stats.Recv.MbpsRate)
			}
		}
	}
	// server_status
	resp, err = getMetrics("/manage/server_status")
	if err == nil {
		var result ServerStatus
		if err := json.Unmarshal([]byte(resp), &result); err != nil {
			log.Errorf("Can not unmarshal JSON from /manage/server_status")
		} else {
			fmt.Fprintf(w, "nimble_connections %d\n", result.Connections)
			fmt.Fprintf(w, "nimble_outrate %d\n", result.OutRate)
			fmt.Fprintf(w, "nimble_ram_cache_size %d\n", result.RAMCacheSize)
			fmt.Fprintf(w, "nimble_file_cache_size %d\n", result.FileCacheSize)
			fmt.Fprintf(w, "nimble_ram_cache_size_max %d\n", result.MaxRAMCacheSize)
			fmt.Fprintf(w, "nimble_file_cache_size_max %d\n", result.MaxFileCacheSize)
			fmt.Fprintf(w, "nimble_sysinfo_ap %d\n", result.SysInfo.Ap)
			fmt.Fprintf(w, "nimble_sysinfo_scl %s\n", result.SysInfo.Scl)
			fmt.Fprintf(w, "nimble_sysinfo_tpms %d\n", result.SysInfo.Tpms)
			fmt.Fprintf(w, "nimble_sysinfo_fpms %d\n", result.SysInfo.Fpms)
			fmt.Fprintf(w, "nimble_sysinfo_tsss %d\n", result.SysInfo.Tsss)
			fmt.Fprintf(w, "nimble_sysinfo_fsss %d\n", result.SysInfo.Fsss)
		}
	}
}

func getMetrics(path string) (string, error) {
	auth_string := ""
	if len(*NimbleAuthSalt) > 0 && len(*NimbleAuthHash) > 0 {
		auth_string = "?salt=" + *NimbleAuthSalt + "&hash=" + *NimbleAuthHash
	}
	connect_string := *NimbleAddress + path + auth_string

	client := &http.Client{Timeout: 2 * time.Second}
	resp, err := client.Get(connect_string)

	log.Debugf("Connecting to %s", *NimbleAddress)
	if err != nil {
		log.Errorf("Cant connect to %s", *NimbleAddress)
		return "", errors.New("cant connect to nimble")
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		log.Errorf("Got %d response code", resp.StatusCode)
		return "", errors.New("cant get metrics")
	}
	body, _ := io.ReadAll(resp.Body)
	//fmt.Printf("client: response body: %s\n", body)
	return string(body), nil
}

func main() {
	flag.Parse()
	lvl, _ := log.ParseLevel(*LogLevel)
	log.SetLevel(lvl)
	switch *LogFmt {
	case "json":
		log.SetFormatter(&log.JSONFormatter{})
	}

	http.HandleFunc("/", processRequest)

	http.ListenAndServe(*ListenAddress, nil)
}
