package main

import (
	"fmt"
	"github.com/LeBronQ/Mobility"
	"github.com/LeBronQ/RadioChannelModel"
	"io/ioutil"
	"strconv"
	"bytes"
	"net/http"
	"encoding/json"
	"github.com/LeBronQ/kdtree"
	"github.com/LeBronQ/kdtree/points"
	//"sync"
	consulapi "github.com/hashicorp/consul/api"
)

const (
	NodeNum = 100
	consul_address = "127.0.0.1:8500"
)
var channel_se = Discovery("Default_ChannelModel")
var mobility_se = Discovery("Default_MobilityModel")

type Node struct {
	ID      int64
	MobNode Mobility.Node
	WNode   RadioChannelModel.WirelessNode
	Range   float64
}

type ChannelModel struct {
	LargeScaleModel string    `json:"largescalemodel"`
	SmallScaleModel string    `json:"smallscalemodel"`
}

type Position struct {
	X float64    `json:"x"`
	Y float64    `json:"y"`
	Z float64    `json:"z"`
}

type WirelessNode struct {
	Frequency  float64    `json:"frequency"`
	BitRate    float64    `json:"bitrate"`
	Modulation string     `json:"modulation"`
	BandWidth  float64    `json:"bandwidth"`
	M          float64    `json:"m"`
	PowerInDbm float64    `json:"powerindbm"`
}

type ChannelReqParams struct {
	LinkId 	      	  int64					`json:"linkid"`
	TxNode 		  RadioChannelModel.WirelessNode	`json:"txnode"`
	RxNode		  RadioChannelModel.WirelessNode	`json:"rxnode"`
	TxPosition 	  RadioChannelModel.Position		`json:"txposition"`
	RxPosition    	  RadioChannelModel.Position		`json:"rxposition"`
	Model 		  ChannelModel				`json:"model"`
}

type MobilityReqParams struct {
	Node    Mobility.Node    `json:"node"`
}

type TreeNodeData struct {
	ID int64
}

func GenerateNodes() []*Node {
	arr := make([]*Node, NodeNum)
	for i := 0; i < NodeNum; i++ {
		node := &Mobility.Node{
			Pos:  Mobility.Nbox.RandomPosition3D(),
			Time: 10,
			V: Mobility.Speed{
				X: 10., Y: 10., Z: 10.,
			},
			Model: "RandomWalk",
			Param: Mobility.RandomWalkParam{
				MinSpeed: 0,
				MaxSpeed: 20,
			},
		}
		wirelessNode := &RadioChannelModel.WirelessNode{
			Frequency:  2.4e+9,
			BitRate:    5.0e+7,
			Modulation: "BPSK",
			BandWidth:  2.0e+7,
			M:          0,
			PowerInDbm: 20,
		}
		n := &Node{
			ID:      int64(i),
			MobNode: *node,
			WNode:   *wirelessNode,
			Range:   2000.0,
		}
		arr[i] = n
	}
	return arr
}

func Discovery(serviceName string) []*consulapi.ServiceEntry {
	config := consulapi.DefaultConfig()
	config.Address = consul_address
	client, err := consulapi.NewClient(config)
	if err != nil {
		fmt.Printf("consul client error: %v", err)
	}
	service, _, err := client.Health().Service(serviceName, "", false, nil)
	if err != nil {
		fmt.Printf("consul client get serviceIp error: %v", err)
	}
	return service
}

func MobilityRequest(node Mobility.Node, se *consulapi.ServiceEntry) []byte {
	//se := Discovery("Default_MobilityModel")
	port := se.Service.Port
	address := se.Service.Address
	request := "http://" + address + ":" + strconv.Itoa(port) + "/mobility"
	param := MobilityReqParams{
		Node: node,
	}
	jsonData, err := json.Marshal(param)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return nil
	}

	requestBody := bytes.NewBuffer(jsonData)

	req, err := http.NewRequest("POST", request, requestBody)
    	if err != nil {
        	fmt.Println(err)
        	return nil
    	}
    	req.Header.Set("Content-Type", "application/json")
    	req.Header.Set("Content-Length", fmt.Sprintf("%d", requestBody.Len()))
 
    	client := &http.Client{}
    	resp, err := client.Do(req)
    	if err != nil {
        	fmt.Println(err)
        	return nil
    	}
    	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status code:", resp.StatusCode)
		return nil
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return nil
	}
	return body
}

func UpdatePosition(NodeArr []*Node, TreeNodeArr []kdtree.Point) []kdtree.Point{
	for _, node := range NodeArr {
		res := MobilityRequest(node.MobNode, mobility_se[0])
		var newNode MobilityReqParams
		err := json.Unmarshal(res, &newNode)
		if err != nil {
			fmt.Println("Error:", err)
		}
		node.MobNode = newNode.Node
		p := points.NewPoint([]float64{node.MobNode.Pos.X, node.MobNode.Pos.Y, node.MobNode.Pos.Z}, TreeNodeData{ID: node.ID})
		TreeNodeArr = append(TreeNodeArr, p)
	}
	return TreeNodeArr
}

func ChannelRequest(Tx RadioChannelModel.WirelessNode, Rx RadioChannelModel.WirelessNode, TxPos RadioChannelModel.Position, RxPos RadioChannelModel.Position, se *consulapi.ServiceEntry) {
	//se := Discovery("Default_ChannelModel")
	port := se.Service.Port
	address := se.Service.Address
	request := "http://" + address + ":" + strconv.Itoa(port) + "/model"
	mod := ChannelModel{
		LargeScaleModel: "FreeSpacePathLossModel",
		SmallScaleModel: "NakagamiFadingModel",
	}
	param := ChannelReqParams{
		LinkId:  0,
		TxNode: Tx,
		RxNode: Rx,
		TxPosition: TxPos,
		RxPosition: RxPos,
		Model: mod,
	}

	jsonData, err := json.Marshal(param)
	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return
	}

	requestBody := bytes.NewBuffer(jsonData)

	req, err := http.NewRequest("POST", request, requestBody)
    	if err != nil {
        	fmt.Println(err)
        	return
    	}
	
    	req.Header.Set("Content-Type", "application/json")
    	req.Header.Set("Content-Length", fmt.Sprintf("%d", requestBody.Len()))
 
    	client := &http.Client{}
   	resp, err := client.Do(req)
    	if err != nil {
        	fmt.Println(err)
        	return
    	}
    	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Unexpected status code:", resp.StatusCode)
		return
	}

	_, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		fmt.Println("Error reading response body:", err)
		return
	}

	//fmt.Println("Response:", string(body))
}

func UpdateNeighborsAndCalculatePLR(graph [NodeNum][]*Node, NodeArr []*Node, tree *kdtree.KDTree) {
	for i := 0; i < NodeNum; i++ {
		var neighbors []*Node
		node := NodeArr[i]
		distance := node.Range
		center := points.NewPoint([]float64{node.MobNode.Pos.X, node.MobNode.Pos.Y, node.MobNode.Pos.Z}, TreeNodeData{ID: node.ID})
		res := tree.QueryBallPoint(center, distance)
		//fmt.Printf("%v\n",res[0].GetData().(TreeNodeData).ID)
		for _, neigh := range res {
			neigh_ID := neigh.GetData().(TreeNodeData).ID
			if node.ID == neigh_ID {
				continue
			} else {
				neigh_node := NodeArr[neigh_ID]
				ChannelRequest(node.WNode, neigh_node.WNode, RadioChannelModel.Position(node.MobNode.Pos), RadioChannelModel.Position(neigh_node.MobNode.Pos),channel_se[0])
				neighbors = append(neighbors, neigh_node)
			} 
		}
		graph[i] = neighbors
		//fmt.Printf("%v\n", graph[i])
	}
}


func main() {
	var nodes []kdtree.Point
	NodeArr := make([]*Node, NodeNum)
	NodeArr = GenerateNodes()
	nodes = UpdatePosition(NodeArr, nodes)
	tree := kdtree.New(nodes)
	var graph [NodeNum][]*Node
	UpdateNeighborsAndCalculatePLR(graph, NodeArr, tree)
}
