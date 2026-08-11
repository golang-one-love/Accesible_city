.PHONY: dev build test lint clean

dev:
	npm run dev

build:
	npm run build

test:
	npm run test

lint:
	npm run lint

clean:
	rm -rf dist node_modules

docker-build:
	docker build -t accessible-path-frontend .

docker-run:
	docker run -p 3000:80 accessible-path-frontend