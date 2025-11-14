package main

func main() {

	cloudantImport, err := NewCloudantImport()
	if err != nil {
		panic(err)
	}

	cloudantImport.Run()
}
