package main

import (
	"context"
	"encoding/json"
	"log"
	"os"
	"strings"
)

var (
	user     string
	pass     string
	baseUrl  string
	clientId string
)

func init() {
	user = os.Getenv("AGFA_USER")
	if user == "" {
		log.Fatalf("please provide AGFA_USER env")
	}
	pass = os.Getenv("AGFA_PASS")
	if pass == "" {
		log.Fatalf("please provide AGFA_PASS env")
	}
	baseUrl = os.Getenv("AGFA_URL")
	if baseUrl == "" {
		log.Fatalf("please provide AGFA_URL env")
	}
	clientId = os.Getenv("AGFA_CLIENT")
	if clientId == "" {
		log.Fatalf("please provide AGFA_CLIENT env")
	}
}

func main() {
	auth, err := NewManualAuth(user, pass, loadDefaults)
	if err != nil {
		log.Fatalln(err)
	}

	if _, err = auth.GetAuthHeader(); err != nil {
		log.Fatalln("auth error:", err)
	}

	ctx := context.Background()
	log.Println("fetching worklist...")
	r, err := auth.FetchListById(ctx, "ql-130738")
	if err != nil {
		log.Fatalln("error getting list:", err)
	}

	var bundle Bundle
	if err := json.NewDecoder(r).Decode(&bundle); err != nil {
		log.Fatalln("json.Decode:", err)
	}

	var entries []ListEntry
	for _, entry := range bundle.Entry {
		if entry.Resource.ResourceType == "List" {
			entries = entry.Resource.Entry
			break
		}
	}
	if len(entries) == 0 {
		log.Println("no records in list")
		os.Exit(0)
	}

	log.Printf("found %d entries\n", len(entries))
	taskId := ""
	for _, e := range entries {
		if strings.HasPrefix(e.Item.Reference, "Task/") {
			taskId = e.Item.Reference[strings.Index(e.Item.Reference, "/"):]
			break
		}
	}

	log.Printf("attempting to fetch task ID %q\n", taskId)
	task, err := auth.FetchTaskById(ctx, taskId)
	if err != nil {
		log.Fatalln("error getting task:", err)
	}

	if len(task.Input) == 0 {
		log.Fatalln("no service requests to process!")
	}

	reqId := task.Input[0].ValueReference.Reference[strings.Index(task.Input[0].ValueReference.Reference, "/"):]

	log.Printf("attempting to fetch service request with ID %q\n", reqId)
	svcReq, err := auth.FetchServiceRequestById(ctx, reqId)
	if err != nil {
		log.Fatalln("error getting service request:", err)
	}

	log.Println(svcReq)
}

func loadDefaults(auth *ManualAuth) {
	auth.Domain = DefaultDomain
	auth.BaseUrl = baseUrl
	auth.ClientId = clientId
}
