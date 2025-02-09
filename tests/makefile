setup:
	npm install
	make browsers

browsers:
	npx playwright install

server:
	docker run -d -p 8080:8080 -p 8443:8443 -p 8666:8666 mechanus server --log.level debug

local-server:
	go run ../server/main.go server --log.level debug --web.static.folder "../client/build"