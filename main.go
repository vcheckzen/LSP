package main

func main() {
	go listenDns()
	go listenTls()
	listenApi()
}
