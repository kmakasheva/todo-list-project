package handlers

//
//import (
//	"net/http"
//)
//
//func GetData(w http.ResponseWriter, r *http.Request) {
//	now := r.URL.Query().Get("now")
//	if len(now) != 8 {
//		w.WriteHeader(http.StatusBadRequest)
//	}
//
//	date := r.URL.Query().Get("date")
//	if len(date) != 8 {
//		w.WriteHeader(http.StatusBadRequest)
//	}
//
//	repeat := r.URL.Query().Get("repeat")
//	if len(repeat) == 0 {
//		w.WriteHeader(http.StatusBadRequest)
//	}
//}
