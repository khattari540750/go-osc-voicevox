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
	"sync/atomic"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/faiface/beep/wav"
	"github.com/hypebeast/go-osc/osc"
)

var isSpeaking int32 // 0: not speaking, 1: speaking

var (
	voicevoxEngineURL string
	speakerID         int
	oscListenPort     int
)

func main() {
	flag.StringVar(&voicevoxEngineURL, "engine", "http://127.0.0.1:50021", "VOICEVOX ENGINEのURL")
	flag.IntVar(&speakerID, "speaker", 1, "VOICEVOXの話者ID")
	flag.IntVar(&oscListenPort, "port", 9000, "OSC受信ポート")
	flag.Parse()
	// OSCサーバ起動
	addr := fmt.Sprintf(":%d", oscListenPort)
	server := newOSCServer(addr, speak)
	log.Printf("OSCサーバ起動: %s (/text でテキスト受信)", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

// OSCサーバを生成し、/text受信時にspeakFuncを呼ぶ
func newOSCServer(addr string, speakFunc func(string)) *osc.Server {
	dispatcher := osc.NewStandardDispatcher()
	dispatcher.AddMsgHandler("/text", func(msg *osc.Message) {
		if len(msg.Arguments) == 0 {
			return
		}
		text, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}
		// 発話中なら無視
		if !atomic.CompareAndSwapInt32(&isSpeaking, 0, 1) {
			log.Printf("発話中のためOSC受信を無視: %s", text)
			return
		}
		go func() {
			speakFunc(text)
			atomic.StoreInt32(&isSpeaking, 0)
		}()
	})
	return &osc.Server{
		Addr:       addr,
		Dispatcher: dispatcher,
	}
}

// テキストをVOICEVOXで音声合成し再生
func speak(text string) {
	log.Printf("受信テキスト: %s", text)
	audioQueryJSON, err := fetchAudioQuery(text, speakerID)
	if err != nil {
		log.Printf("audio_query失敗: %v", err)
		return
	}
	wavData, err := fetchSynthesis(audioQueryJSON, speakerID)
	if err != nil {
		log.Printf("synthesis失敗: %v", err)
		return
	}
	if err := playWav(wavData); err != nil {
		log.Printf("再生失敗: %v", err)
	}
}

// VOICEVOX ENGINE /audio_query APIを叩き、音声合成用JSONを取得
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

// VOICEVOX ENGINE /synthesis APIを叩き、WAVデータを取得
func fetchSynthesis(audioQueryJSON []byte, speaker int) ([]byte, error) {
	url := fmt.Sprintf("%s/synthesis?speaker=%d", voicevoxEngineURL, speaker)
	resp, err := http.Post(url, "application/json", bytes.NewReader(audioQueryJSON))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

// WAVデータをPCM再生
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
