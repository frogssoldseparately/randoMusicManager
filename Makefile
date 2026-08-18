build:
	go build -o bin/randoMusicManager.exe cmd/randoMusicManager/main.go
parser:
	hopper resources/grammar.txt pkg/parser