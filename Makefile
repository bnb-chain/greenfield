
proto-gen:
	ignite generate proto-go
	#cd proto && buf generate && cp -r github.com/bnb-chain/bfs/x/* ../x && rm -rf github.com

proto-format:
	buf format -w
