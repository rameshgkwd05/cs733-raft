package main

import (
	"encoding/json"
	"errors"
	handler "github.com/rameshgkwd05/cs733-raft/handler"
	raft "github.com/rameshgkwd05/cs733-raft/raft"
	types "github.com/rameshgkwd05/cs733-raft/types"
	"io/ioutil"
	"log"
	"sync"
)

type NullWriter int
var clusterConfig *types.ClusterConfig
var sharedMap  types.SharedMaptype //shared map of eventCh identified by server ID
var wg *sync.WaitGroup

func (NullWriter) Write([]byte) (int, error) {
	return 0, nil
}

func ReadConfig(path string) (*types.ClusterConfig, error) {
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var conf types.ClusterConfig
	err = json.Unmarshal(data, &conf)
	if err == nil && len(conf.Servers) < 1 {
		err = errors.New("No Server Configuration found")
	}
	return &conf, err
}

//start the server Id will be prespecified
func StartServer(id int) {
	defer wg.Done()
	log.SetOutput(new(NullWriter))

	//TODO: Read the config.json file to get all the server configurations
	clusterConfig, err := ReadConfig("config.json")
	if err != nil {
		log.Println("Error parsing config file : ", err.Error())
		return
	}

	log.Println("Starting server with ID ", id)

	commitCh := make(chan raft.LogEntry, 10000)
	raftInstance, err := raft.NewRaft(clusterConfig, id, commitCh,&sharedMap)
	if err != nil {
		log.Println("Error creating server instance : ", err.Error())
	}

	
	var clientPort int
	for _, server := range raftInstance.ClusterConfig.Servers {
		
		//Initially no one will be leader. All servers will start off as follower
		//raftInstance.LeaderID = types.UnknownLeader
		//raftInstance.CurrentState =types.FOLLOWER
	
		//for testing I am putting server 1 as leader
		raftInstance.LeaderID = 1
		raftInstance.CurrentState =types.LEADER
		
	
		//Initialize the connection handler module
		if server.Id == raftInstance.ServerID {
			clientPort = server.ClientPort
		}
	}
	
	if clientPort <= 0 {
		log.Println("Server's client port not valid")
	} else {
		go handler.StartConnectionHandler(clientPort, raftInstance.AppendRequestChannel)
	}

	//Inititialize the KV Store Module
	go raft.InitializeKVStore(raftInstance.CommitCh)

	//Now start the SharedLog module
	raftInstance.StartServer()

	log.Println("Started raft Instance for server ID ", raftInstance.ServerID)
}

func main() {
	serverConfig := []types.ServerConfig{
		{1, "localhost", 5001, 6001},
		{2, "localhost", 5002, 6002},
		{3, "localhost", 5003, 6003},
		{4, "localhost", 5004, 6004},
		{5, "localhost", 5005, 6005}}
	clusterConfig = &types.ClusterConfig{"/log", serverConfig}

	data, _ := json.Marshal(*clusterConfig)
	ioutil.WriteFile("config.json", data, 0644)
	
	serverReplicas := len((*clusterConfig).Servers)
	
	//sharedMap will be map of eventCh shared among all server goroutines
	//mapping will be serverID <-> eventch
	sharedMap = make(map[int]chan int)
	
	wg = new(sync.WaitGroup)
	//start all 5 server goroutines
	index := 1
	for index <= serverReplicas {
		wg.Add(1)
		//eventCh
		sharedMap[index] = make(chan int,1000)
		go StartServer(index)
		index++
	}
	wg.Wait()
	
}
