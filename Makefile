build:
	cd web && npm run build
	ENV=prod go build -buildvcs=false -o ./bin/liift ./main.go

dev:
	cd web && npm run dev & air && fg