# go-osc-voicevox

OSCで受信したテキストをVOICEVOX ENGINE（HTTP API）で合成し、リアルタイムで再生するGoアプリです。

## 特徴
- OSCでテキスト受信（/text アドレス）
- VOICEVOX ENGINE HTTP API（/audio_query, /synthesis）で音声合成
- WAVデータをPCM再生（cgo不要、Windows/macOS両対応）
- コマンドライン引数でエンジンURL・話者ID・OSCポート指定可

## 使い方

### 1. VOICEVOX ENGINEの起動
公式のVOICEVOX ENGINE（FastAPIサーバ）を起動してください。

例: 
```
voicevox_engine --host 0.0.0.0 --port 50021
```

### 2. このアプリのビルド
#### macOS
```
go build -o go-osc-voicevox main.go
```
#### Windows用クロスビルド（macから）
```
GOOS=windows GOARCH=amd64 go build -o go-osc-voicevox.exe main.go
```

### 3. 実行例
```
# デフォルト（127.0.0.1:50021, 話者ID=1, ポート9000）
./go-osc-voicevox

# 任意のVOICEVOX ENGINEや話者ID、OSCポートを指定
./go-osc-voicevox -engine http://localhost:50021 -speaker 3 -port 9001
```

### 4. OSCクライアントからの送信例
`/text` アドレスでテキスト（日本語可）を送信してください。

例: (Python)
```python
from pythonosc.udp_client import SimpleUDPClient
client = SimpleUDPClient('127.0.0.1', 9000)
client.send_message('/text', 'こんにちは、ずんだもんです')
```

## 話者ID（speakerID）について
VOICEVOX ENGINEの話者IDはエンジンのバージョンや辞書によって異なります。
`/speakers` エンドポイントで確認できます。

## ライセンス
MIT
