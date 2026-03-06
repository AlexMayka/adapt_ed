run:
	echo "Запуск программы"
	go run ./cmd/main.go

build:
	echo "Сборка бинарника"
	go build -o bin/backend-sales-radar ./cmd


run-cmd-make:
	echo "Запуск бинарного файла"
	./bin/backend-sales-radar