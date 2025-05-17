setup:
	npm install
	make browsers

browsers:
	npx playwright install

server:
	-docker rm mechanus-server
	docker run --name mechanus-server --env-file .env -d -p 8080:8080 -p 8443:8443 -p 8666:8666 mechanus ./mechanus server --log.level debug --log.format text --log.report-caller

local-server:
	go run ../server/main.go server --log.level debug --web.static.folder "../client/build"

test:
	npm run test