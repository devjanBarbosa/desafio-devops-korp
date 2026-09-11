package main

import (
"encoding/json"
"log"
"net/http"
"time"

"github.com/prometheus/client_golang/prometheus"
"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
httpRequestsTotal = prometheus.NewCounterVec(
prometheus.CounterOpts{
Name: "http_requests_total",
Help: "Total de requisicoes HTTP recebidas",
},
[]string{"path", "status"},
)

serviceUp = prometheus.NewGauge(
prometheus.GaugeOpts{
Name: "http_server_up",
Help: "Status do servico (1 para up, 0 para down)",
},
)
)

func init() {
prometheus.MustRegister(httpRequestsTotal)
prometheus.MustRegister(serviceUp)
serviceUp.Set(1)
}

type KorpResponse struct {
Nome    string `json:"nome"`
Horario string `json:"horario"`
}

func korpHandler(w http.ResponseWriter, r *http.Request) {
if r.Method != http.MethodGet {
http.Error(w, "Metodo nao permitido", http.StatusMethodNotAllowed)
httpRequestsTotal.WithLabelValues(r.URL.Path, "405").Inc()
return
}

response := KorpResponse{
Nome:    "Projeto Korp",
Horario: time.Now().UTC().Format(time.RFC3339),
}

w.Header().Set("Content-Type", "application/json")
w.WriteHeader(http.StatusOK)
if err := json.NewEncoder(w).Encode(response); err != nil {
httpRequestsTotal.WithLabelValues(r.URL.Path, "500").Inc()
return
}

httpRequestsTotal.WithLabelValues(r.URL.Path, "200").Inc()
}

func main() {
mux := http.NewServeMux()

mux.HandleFunc("/projeto-korp", korpHandler)
mux.Handle("/metrics", promhttp.Handler())

server := &http.Server{
Addr:         ":8080",
Handler:      mux,
ReadTimeout:  5 * time.Second,
WriteTimeout: 10 * time.Second,
}

log.Println("Servidor iniciado na porta 8080...")
if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
serviceUp.Set(0)
log.Fatalf("Erro ao iniciar servidor: %v", err)
}
}