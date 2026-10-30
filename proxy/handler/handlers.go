package handlers

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type Handler struct{
	BaseURL string
}
func NewHandler(BaseURL string)*Handler{
	return &Handler{
		BaseURL: BaseURL,
	}
}
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request){
	path := "/tasks"
	url := h.BaseURL + path
	if r.URL.RawQuery != ""{
		url = url + "?" + r.URL.RawQuery
	}
	resp, err := http.Get(url)
	if err != nil{
		http.Error(w, "502",http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w,resp.Body)


	
}
func (h *Handler) GetTaskByID(w http.ResponseWriter, r *http.Request){
	url := h.BaseURL + r.URL.Path
	resp,err := http.Get(url)
	if err != nil{
		http.Error(w, "502", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}
func (h *Handler) PostTask(w http.ResponseWriter, r *http.Request){
	path := "/tasks"
	url := h.BaseURL + path
	body,err := io.ReadAll(r.Body)
	if err != nil{
		http.Error(w, "502", http.StatusBadGateway)
		return
	}
	req,err := http.NewRequest(http.MethodPost, url,bytes.NewReader(body))
	if err != nil{
		http.Error(w, "502", http.StatusBadGateway)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	client := http.DefaultClient
	resp,err := client.Do(req)
	
	if err != nil{
		http.Error(w, "502", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w,resp.Body)
	
	

}

func (h *Handler) PutTask (w http.ResponseWriter, r*http.Request){
	url := h.BaseURL + r.URL.Path

	body,err := io.ReadAll(r.Body)
	if err != nil{
		http.Error(w ,"502",http.StatusBadGateway)
		return
	}
	req, err := http.NewRequest(http.MethodPut,url,bytes.NewReader(body))
	if err != nil{
		http.Error(w,"405",http.StatusMethodNotAllowed)
		return
	}
	req.Header.Set("content-type","application/json")
	clinet := http.DefaultClient
	resp, err := clinet.Do(req)
	if err != nil{
		http.Error(w, "502", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w,resp.Body)

}

func (h *Handler) DeleteTasks(w http.ResponseWriter, r *http.Request){
	url := h.BaseURL + r.URL.Path
	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil{
		http.Error(w, "502",http.StatusBadGateway)
	}
	req.Header.Set("content-type", "application/josn")
	client := http.DefaultClient
	resp,err := client.Do(req)
	if err != nil{
		http.Error(w, "502", http.StatusBadGateway)
	}
	defer resp.Body.Close()
	w.WriteHeader(resp.StatusCode)
	io.Copy(w,resp.Body)

}
func (h *Handler)GetHealth(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w, "405", http.StatusMethodNotAllowed)
		return
	}
	path := "/health"
	url := h.BaseURL + path
	resp, err := http.Get(url)
	if err != nil{
		http.Error(w, "502", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(resp.StatusCode)
	io.Copy(w,resp.Body)
}
func (h *Handler)GetApiHealth(w http.ResponseWriter, r *http.Request){
	if r.Method != http.MethodGet{
		http.Error(w , "405", http.StatusMethodNotAllowed)
		return
	}
	w.Header().Set("content-type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(map[string]string{"status":"ok"})
	
}