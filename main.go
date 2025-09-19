package main

import (
	"bytes"
	"encoding/json"
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

const (
	voicevoxEngineURL       = "http://127.0.0.1:50021"
	zundamonNormalSpeakerID = 1 // ずんだもんノーマルのID（VOICEVOX ENGINE v0.14.5以降）
	oscListenPort           = 9000
)

type AudioQuery struct {
	// VOICEVOX ENGINEのaudio_query APIのレスポンス構造体
}

func main() {
	// OSCサーバ起動
	addr := fmt.Sprintf(":%d", oscListenPort)
	dispatcher := osc.NewStandardDispatcher()
	dispatcher.AddMsgHandler("/text", func(msg *osc.Message) {
		if len(msg.Arguments) == 0 {
			return
		}
		text, ok := msg.Arguments[0].(string)
		if !ok {
			return
		}
		go speak(text)
	})
	server := &osc.Server{
		Addr:       addr,
		Dispatcher: dispatcher,
	}
	log.Printf("OSCサーバ起動: %s (/text でテキスト受信)", addr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func speak(text string) {
	log.Printf("受信テキスト: %s", text)
	// 1. audio_query
	query, err := audioQuery(text, zundamonNormalSpeakerID)
	if err != nil {
		log.Printf("audio_query失敗: %v", err)
		return
	}
	// 2. synthesis
	wavData, err := synthesis(query, zundamonNormalSpeakerID)
	if err != nil {
		log.Printf("synthesis失敗: %v", err)
		return
	}
	// 3. 再生
	if err := playWav(wavData); err != nil {
		log.Printf("再生失敗: %v", err)
	}
}

func audioQuery(text string, speaker int) ([]byte, error) {
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
	// JSONとして一度デコードし、再エンコードして返す
	var query map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&query); err != nil {
		return nil, err
	}
	return json.Marshal(query)
}

func synthesis(query []byte, speaker int) ([]byte, error) {
	url := fmt.Sprintf("%s/synthesis?speaker=%d", voicevoxEngineURL, speaker)
	resp, err := http.Post(url, "application/json", bytes.NewReader(query))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	return io.ReadAll(resp.Body)
}

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
