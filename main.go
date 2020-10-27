package main

func main() {
	wait := make(chan int, 1)
	go listenDns()
	go listenTls()
	go listenApi()
	<-wait
}
