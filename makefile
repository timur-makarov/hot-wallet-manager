.PHONY: api_gen proto_gen

api_gen:
	cd api && go generate generate.go && cd .. 

proto_gen:
	cd proto && go generate generate.go && cd ..
