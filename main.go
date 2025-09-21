package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/hypebeast/go-osc/osc"
)

var (
	voicevoxEngineURL string
	speakerID         int
	oscListenPort     int
	queueSize         int
	textQueue         chan string
)

func main() {
	flag.StringVar(&voicevoxEngineURL, "engine", "http://127.0.0.1:50021", "VOICEVOX ENGINE URL")
	flag.IntVar(&speakerID, "speaker", 1, "VOICEVOX speaker ID")
	flag.IntVar(&oscListenPort, "port", 9000, "OSC listen port")
	flag.IntVar(&queueSize, "queue", 2, "Queue size (1-10)")
	flag.Parse()
	
	// Validate queue size
	if queueSize < 1 || queueSize > 10 {
		log.Fatal("Queue size must be between 1 and 10")
	}
	
	// Initialize text queue
	textQueue = make(chan string, queueSize)
	
	// Start speech worker
	go speechWorker()
	
	// Start OSC server
	addr := fmt.Sprintf(":%d", oscListenPort)
	server := newOSCServer(addr)
	log.Printf("OSC server started: %s (listening for /text, queue size: %d)", addr, queueSize)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// Create an OSC server and call speakFunc when /text is received
func newOSCServer(addr string) *osc.Server {
	dispatcher := osc.NewStandardDispatcher()
	dispatcher.AddMsgHandler("/text", func(msg *osc.Message) {
		if len(msg.Arguments) == 0 {
			return
		}
		text, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}
		// Try to add text to queue, ignore if queue is full
		select {
		case textQueue <- text:
			log.Printf("Text queued: %s (queue: %d/%d)", text, len(textQueue), cap(textQueue))
		default:
			log.Printf("Queue full, OSC message ignored: %s", text)
		}
	})
	return &osc.Server{
		Addr:       addr,
		Dispatcher: dispatcher,
	}
}

// Speech worker processes queued texts one by one
func speechWorker() {
	for text := range textQueue {
		speak(text)
	}
}

// Synthesize and play speech from text using VOICEVOX
func speak(text string) {
	log.Printf("Received text: %s", text)
	audioQueryJSON, err := fetchAudioQuery(text, speakerID)
	if err != nil {
		log.Printf("audio_query failed: %v", err)
		return
	}
	wavData, err := fetchSynthesis(audioQueryJSON, speakerID)
	if err != nil {
		log.Printf("synthesis failed: %v", err)
		return
	}
	if err := playWav(wavData); err != nil {
		log.Printf("playback failed: %v", err)
	}
}

// Call VOICEVOX ENGINE /audio_query API and get synthesis JSON
func fetchAudioQuery(text string, speaker int) ([]byte, error) {
	apiUrl := fmt.Sprintf("%s/audio_query", voicevoxEngineURL)
	data := url.Values{}
	data.Set("text", text)
	data.Set("speaker", fmt.Sprintf("%d", speaker))
	req, err := http.NewRequest("POST", apiUrl+"?"+data.Encode(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var query map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&query); err != nil {
		return nil, err
	}
	return json.Marshal(query)
}

// Call VOICEVOX ENGINE /synthesis API and get WAV data
func fetchSynthesis(audioQueryJSON []byte, speaker int) ([]byte, error) {
	url := fmt.Sprintf("%s/synthesis?speaker=%d", voicevoxEngineURL, speaker)
	resp, err := http.Post(url, "application/json", bytes.NewReader(audioQueryJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// Play WAV data as PCM
func playWav(wavData []byte) error {
	streamer, format, err := wav.Decode(bytes.NewReader(wavData))
	if err != nil {
		return err
	}
	defer streamer.Close()
	if err := speaker.Init(format.SampleRate, format.SampleRate.N(time.Second/10)); err != nil {
		return err
	}
	done := make(chan bool)
	speaker.Play(beep.Seq(streamer, beep.Callback(func() {
		done <- true
	})))
	<-done
	return nil
}
