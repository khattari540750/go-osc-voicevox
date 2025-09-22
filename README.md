# go-osc-voicevox

A Go application that receives text via OSC, synthesizes speech using VOICEVOX ENGINE (HTTP API), and plays it in real time.

## Features
- Receives text via OSC (`/text` address)
- Speech synthesis using VOICEVOX ENGINE HTTP API (`/audio_query`, `/synthesis`)
- Plays WAV data as PCM (no cgo required, works on both Windows and macOS)
- Command-line flags for engine URL, speaker ID, OSC port, and queue size
- Queue system for handling multiple OSC messages (configurable size 1-10, default 2)
- Sequential speech processing (one speech at a time, queued messages processed in order)

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
# Default (127.0.0.1:50021, speaker ID=1, port 9000, queue size=0)
./go-osc-voicevox

# Specify VOICEVOX ENGINE, speaker ID, OSC port, and queue size
./go-osc-voicevox -engine http://localhost:50021 -speaker 3 -port 9001 -queue 5
```

#### Command-line options
- `-engine`: VOICEVOX ENGINE URL (default: http://127.0.0.1:50021)
- `-speaker`: VOICEVOX speaker ID (default: 1)
- `-port`: OSC listen port (default: 9000)
- `-queue`: Queue size for OSC messages, 0-10 (default: 0, 0=no queue, ignore while speaking)

### 4. Example: Sending OSC from a client
Send text (including Japanese) to the `/text` address.

Example (Python):
```python
from pythonosc.udp_client import SimpleUDPClient
client = SimpleUDPClient('127.0.0.1', 9000)
client.send_message('/text', 'こんにちは、ずんだもんです')
```

## Queue System
The application supports two modes for handling multiple OSC messages:

### Queue Mode (queue size 1-10)
- OSC messages are queued and processed sequentially
- Queue size is configurable (1-10)
- When the queue is full, new OSC messages are ignored with a log message
- Each speech synthesis completes before the next queued message is processed

### No Queue Mode (queue size 0, default)
- OSC messages are processed immediately if not currently speaking
- While speaking, new OSC messages are ignored with a log message
- No queuing - only one speech at a time

## About speakerID
The VOICEVOX ENGINE speaker ID depends on the engine version and dictionary. You can check available IDs via the `/speakers` endpoint.

## License
MIT
