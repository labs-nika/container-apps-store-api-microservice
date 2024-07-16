package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"time"

	dapr "github.com/dapr/go-sdk/client"
	"github.com/gorilla/mux"
)

type App struct {
	Router     *mux.Router
	daprClient dapr.Client
}

func (a *App) Initialize(client dapr.Client) {
	a.daprClient = client
	a.Router = mux.NewRouter()

	a.Router.HandleFunc("/", a.Hello).Methods("GET")
	a.Router.HandleFunc("/inventory", a.GetInventory).Methods("GET")
}

func (a *App) Hello(w http.ResponseWriter, r *http.Request) {
	w.Write([]byte("Hello world! It's me"))
}

func (a *App) GetInventory(w http.ResponseWriter, r *http.Request) {
	// 从查询字符串中获取产品ID
	productID := r.URL.Query().Get("id")
	if productID == "" {
		http.Error(w, "Product ID is required", http.StatusBadRequest)
		return
	}

	// 创建库存数据
	inventoryData := map[string]interface{}{
		"productID": productID,
		"quantity":  100, // 假设数量，你可以根据需要调整或从请求中获取
	}

	// 将数据转换为 JSON 字节
	data, err := json.Marshal(inventoryData)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取当前时间并格式化为 yyyyMMddHHmmss 格式
	currentTime := time.Now()
	timePrefix := currentTime.Format("20060102150405") // Go 中的时间格式化布局

	// 使用 Dapr 客户端通过 Azure Blob Storage 绑定存储数据
	resp, err := a.daprClient.InvokeBinding(context.Background(), &dapr.InvokeBindingRequest{
		Name:      "inventory", // 绑定的名称，应与你的配置文件中的名称相匹配
		Operation: "create",    // 操作类型，对于 Azure Blob Storage，通常是 "create"
		Data:      data,        // 要存储的数据
		Metadata: map[string]string{
			"blobName": timePrefix + "-inventory-item-" + productID, // Blob 的名称，包含时间戳前缀和产品ID
		},
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// 可以根据需要使用响应对象 resp
	_ = resp // 这里我们不需要使用响应对象，所以用 _ 来忽略它

	// 响应客户端
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("Inventory in store"))
}

func (a *App) Run(addr string) {
	log.Fatal(http.ListenAndServe(addr, a.Router))
}
