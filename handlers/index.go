package handlers

import (
	"fmt"
	"log"
	"net/http"
	"strconv"

	"github.com/hoodyman/httpdbexample/commonvars"
	"github.com/hoodyman/httpdbexample/db"
)

type modelData struct {
	Csrftoken  string
	InputData  string
	OutputData []outputData
	DbStatus   string
}

type outputData struct {
	DataId    string
	DataValue string
}

func HandlerIndex(w http.ResponseWriter, r *http.Request) {
	m := modelData{}

	conn, err := db.AcquireConn()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	defer conn.Release()

	if r.Method == http.MethodPost {
		err := r.ParseForm()
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			log.Println(fmt.Errorf("form parse error: %v", err))
		} else {
			token := r.PostForm.Get(commonvars.TokenTag)
			if commonvars.Csrf.CheckoutAndDrop(token) {

				if r.PostForm.Has("putdata") {
					inputData := r.PostForm.Get("InputData")
					if len(inputData) != 0 {
						data := db.DbTableData{}
						data.Value = inputData
						err := conn.AppendData(data)
						if err != nil {
							log.Println(fmt.Errorf("append data error: %v", err))
						}
					}
				} else if r.PostForm.Has("deletedata") {
					if _, ok := r.PostForm["DeleteData"]; ok {
						for _, v := range r.PostForm["DeleteData"] {
							data := db.DbTableData{}
							data.Id, err = strconv.Atoi(v)
							if err == nil {
								d_err := conn.DeleteData(data)
								if d_err != nil {
									log.Println(fmt.Errorf("delete data error: %v", d_err))
								}
							} else {
								log.Println(fmt.Errorf("delete data str to int conv error: %v", err))
							}
						}
					}
				} else if r.PostForm.Has("createtable") {
					conn.CreateTable()
				} else if r.PostForm.Has("deletetable") {
					conn.DeleteTable()
				}

			} else {
				w.WriteHeader(http.StatusBadRequest)
				log.Println("invalid token")
			}
		}
	}

	b, err := conn.IsTableExist()
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		log.Println(fmt.Errorf("table exist check error: %v", err))
		return
	}
	if b {
		m.DbStatus = "Ok"
	} else {
		m.DbStatus = "NoTable"
	}

	scan, err := conn.GetDataScanner()
	if err == nil {
		for {
			data, err := scan.Scan()
			if err == nil {
				m.OutputData = append(m.OutputData, outputData{
					DataId:    strconv.Itoa(data.Id),
					DataValue: data.Value,
				})
			} else {
				break
			}
		}
	} else {
		log.Println(fmt.Errorf("data scan error: %v", err))
	}

	m.Csrftoken = fmt.Sprintf(`<input type="hidden", name="%v", value="%v">`, commonvars.TokenTag, commonvars.Csrf.New())
	commonvars.Templ.Apply(w, "index.html", m)
}
