package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"
)

type ViaCEPData struct {
	Cep         string `json:"cep"`
	Logradouro  string `json:"logradouro"`
	Complemento string `json:"complemento"`
	Bairro      string `json:"bairro"`
	Localidade  string `json:"localidade"`
	Uf          string `json:"uf"`
	Ibge        string `json:"ibge"`
	Gia         string `json:"gia"`
	Ddd         string `json:"ddd"`
	Siafi       string `json:"siafi"`
}

type ApiCEPData struct {
	Status   int    `json:"status"`
	Code     string `json:"code"`
	State    string `json:"state"`
	City     string `json:"city"`
	District string `json:"district"`
	Address  string `json:"address"`
}

func ViaCEP(ch chan ViaCEPData, cep string) {
	req, err := http.NewRequest("GET", "https://viacep.com.br/ws/"+cep+"/json/", nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var response ViaCEPData
	err = json.Unmarshal(body, &response)
	ch <- response
}

func ApiCEP(ch chan ApiCEPData, cep string) {
	req, err := http.NewRequest("GET", "https://cdn.apicep.com/file/apicep/"+cep+".json", nil)
	if err != nil {
		panic(err)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		panic(err)
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {
			panic(err)
		}
	}(resp.Body)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}

	var response ApiCEPData
	err = json.Unmarshal(body, &response)
	ch <- response
}

func main() {
	for _, cep := range os.Args[1:] {
		c1 := make(chan ViaCEPData)
		c2 := make(chan ApiCEPData)

		if strings.Contains(cep, string(".")) {
			cep = strings.Replace(cep, ".", "", 1)
		}

		if strings.Contains(cep, string('-')) {
			go ViaCEP(c1, strings.Replace(cep, "-", "", 1))
			go ApiCEP(c2, cep)
		} else {
			go ViaCEP(c1, cep)
			go ApiCEP(c2, cep[:5]+"-"+cep[5:])
		}

		select {
		case msg := <-c1:
			fmt.Println("Received from ViaCEP")
			err := json.NewEncoder(os.Stdout).Encode(msg)
			if err != nil {
				panic(err)
			}
		case msg := <-c2:
			fmt.Println("Received from ApiCEP")
			err := json.NewEncoder(os.Stdout).Encode(msg)
			if err != nil {
				panic(err)
			}
		case <-time.After(time.Second * 1):
			panic(errors.New("timeout"))
		}
	}
}
