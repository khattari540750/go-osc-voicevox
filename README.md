# go-osc-voicevox

A Go application that receives text via OSC, synthesizes speech using VOICEVOX ENGINE (HTTP API), and plays it in real time.

## Features
- Receives text via OSC (`/text` address)
- Speech synthesis using VOICEVOX ENGINE HTTP API (`/audio_query`, `/synthesis`)
- Plays WAV data as PCM (no cgo required, works on both Windows and macOS)
- Command-line flags for engine URL, speaker ID, and OSC port

## Usage

### 1. Start VOICEVOX ENGINE
Launch the official VOICEVOX ENGINE (FastAPI server).

Example:
```
voicevox_engine --host 0.0.0.0 --port 50021
```

### 2. Build this app
#### macOS
```
go build -o go-osc-voicevox main.go
```
#### Cross-build for Windows (from macOS)
```
GOOS=windows GOARCH=amd64 go build -o go-osc-voicevox.exe main.go
```

### 3. Run
```
# Default (127.0.0.1:50021, speaker ID=1, port 9000)
./go-osc-voicevox

# Specify VOICEVOX ENGINE, speaker ID, and OSC port
./go-osc-voicevox -engine http://localhost:50021 -speaker 3 -port 9001
```

### 4. Example: Sending OSC from a client
Send text (including Japanese) to the `/text` address.

Example (Python):
```python
from pythonosc.udp_client import SimpleUDPClient
client = SimpleUDPClient('127.0.0.1', 9000)
client.send_message('/text', 'こんにちは、ずんだもんです')
```

## About speakerID
The VOICEVOX ENGINE speaker ID depends on the engine version and dictionary. You can check available IDs via the `/speakers` endpoint.

## License
MIT
